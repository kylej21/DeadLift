package batches

import (
	"encoding/json"
	"log"
	"net/http"

	"proxy/internal/pubsub"
	"proxy/internal/store"
)

type Handler struct {
	RepairSA string
	Store    *store.Store
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}
	batches, err := h.Store.ListBatchesByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("list batches error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batches)
}

func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batch_id")
	batch, err := h.Store.GetBatch(r.Context(), batchID)
	if err != nil {
		http.Error(w, `{"error":"batch not found"}`, http.StatusNotFound)
		return
	}
	if batch.Status != "pending" {
		http.Error(w, `{"error":"batch is not pending"}`, http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetUserByOrgID(r.Context(), batch.OrgID)
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

	tasks, err := h.Store.GetTasksByBatch(r.Context(), batchID)
	if err != nil {
		log.Printf("batch approve: get tasks error: %v", err)
		http.Error(w, `{"error":"could not load tasks"}`, http.StatusInternalServerError)
		return
	}

	failed := 0
	for _, task := range tasks {
		if task.Status != "pending_approval" {
			continue
		}
		outAttrs := make(map[string]string, len(task.Attributes)+1)
		for k, v := range task.Attributes {
			outAttrs[k] = v
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
	}

	newStatus := "approved"
	if failed > 0 && failed == len(tasks) {
		newStatus = "failed"
	}
	if err := h.Store.UpdateBatchStatus(r.Context(), batchID, newStatus); err != nil {
		log.Printf("batch approve: update batch status error: %v", err)
	}

	log.Printf("batch approve: batch %s — %d tasks published, %d failed", batchID, len(tasks)-failed, failed)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "published": len(tasks) - failed, "failed": failed})
}

func (h *Handler) HandleDeny(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batch_id")
	batch, err := h.Store.GetBatch(r.Context(), batchID)
	if err != nil {
		http.Error(w, `{"error":"batch not found"}`, http.StatusNotFound)
		return
	}
	if batch.Status != "pending" {
		http.Error(w, `{"error":"batch is not pending"}`, http.StatusBadRequest)
		return
	}

	tasks, err := h.Store.GetTasksByBatch(r.Context(), batchID)
	if err != nil {
		log.Printf("batch deny: get tasks error: %v", err)
		http.Error(w, `{"error":"could not load tasks"}`, http.StatusInternalServerError)
		return
	}

	for _, task := range tasks {
		if task.Status == "pending_approval" {
			_ = h.Store.UpdateTaskStatus(r.Context(), task.TaskID, "denied")
		}
	}

	if err := h.Store.UpdateBatchStatus(r.Context(), batchID, "denied"); err != nil {
		log.Printf("batch deny: update batch status error: %v", err)
	}

	log.Printf("batch deny: batch %s denied (%d tasks)", batchID, len(tasks))
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
