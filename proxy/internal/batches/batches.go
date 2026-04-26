package batches

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"proxy/internal/pubsub"
	"proxy/internal/store"
)

type Handler struct {
	RepairSA string
	Store    *store.Store
}

// BatchSummary is derived at query time by grouping tasks on error_class.
// No separate Firestore collection — batches are always in sync with task state.
type BatchSummary struct {
	ErrorClass   string    `json:"error_class"`
	PendingCount int       `json:"pending_count"`
	TotalCount   int       `json:"total_count"`
	FirstSeen    time.Time `json:"first_seen"`
	Status       string    `json:"status"` // "pending" | "resolved"
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}

	tasks, err := h.Store.ListTasksByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("batches list: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	user, _ := h.Store.GetUserByOrgID(r.Context(), orgID)
	threshold := 2
	if user != nil && user.BatchingThreshold > 0 {
		threshold = user.BatchingThreshold
	}

	type group struct {
		pending   int
		total     int
		firstSeen time.Time
	}
	groups := map[string]*group{}
	for _, t := range tasks {
		if t.ErrorClass == "" {
			continue
		}
		g, ok := groups[t.ErrorClass]
		if !ok {
			g = &group{firstSeen: t.CreatedAt}
			groups[t.ErrorClass] = g
		}
		g.total++
		if t.Status == "pending_approval" {
			g.pending++
		}
		if t.CreatedAt.Before(g.firstSeen) {
			g.firstSeen = t.CreatedAt
		}
	}

	summaries := make([]BatchSummary, 0, len(groups))
	for class, g := range groups {
		if g.total < threshold {
			continue
		}
		status := "resolved"
		if g.pending > 0 {
			status = "pending"
		}
		summaries = append(summaries, BatchSummary{
			ErrorClass:   class,
			PendingCount: g.pending,
			TotalCount:   g.total,
			FirstSeen:    g.firstSeen,
			Status:       status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	errorClass := r.PathValue("error_class")
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserByOrgID(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"org not found"}`, http.StatusNotFound)
		return
	}

	token, err := pubsub.GetRepairSAToken(r.Context(), h.RepairSA)
	if err != nil {
		log.Printf("batch approve: get token error: %v", err)
		http.Error(w, `{"error":"could not get publish token"}`, http.StatusInternalServerError)
		return
	}

	tasks, err := h.Store.ListTasksByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("batch approve: list tasks error: %v", err)
		http.Error(w, `{"error":"could not load tasks"}`, http.StatusInternalServerError)
		return
	}

	published, failed := 0, 0
	for _, task := range tasks {
		if task.ErrorClass != errorClass || task.Status != "pending_approval" {
			continue
		}
		outAttrs := make(map[string]string, len(task.Attributes)+1)
		for k, v := range task.Attributes {
			if k != "_deadlift_confidence" && k != "simulate_failure" {
				outAttrs[k] = v
			}
		}
		outAttrs["_deadlift_repaired"] = "true"

		payload := task.FixedPayload
		if payload == "" {
			payload = task.RawPayload
		}

		if err := pubsub.PublishMessage(r.Context(), token, user.MainTopic, payload, outAttrs); err != nil {
			log.Printf("batch approve: publish task %s error: %v", task.TaskID, err)
			_ = h.Store.UpdateTaskStatus(r.Context(), task.TaskID, "failed")
			failed++
			continue
		}
		_ = h.Store.UpdateTaskStatus(r.Context(), task.TaskID, "approved")
		published++
	}

	log.Printf("batch approve: error_class=%s org=%s published=%d failed=%d", errorClass, orgID, published, failed)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "published": published, "failed": failed})
}

func (h *Handler) HandleDeny(w http.ResponseWriter, r *http.Request) {
	errorClass := r.PathValue("error_class")
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}

	tasks, err := h.Store.ListTasksByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("batch deny: list tasks error: %v", err)
		http.Error(w, `{"error":"could not load tasks"}`, http.StatusInternalServerError)
		return
	}

	denied := 0
	for _, task := range tasks {
		if task.ErrorClass == errorClass && task.Status == "pending_approval" {
			_ = h.Store.UpdateTaskStatus(r.Context(), task.TaskID, "denied")
			denied++
		}
	}

	log.Printf("batch deny: error_class=%s org=%s denied=%d", errorClass, orgID, denied)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "denied": denied})
}
