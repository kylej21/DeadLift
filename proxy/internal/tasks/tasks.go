package tasks

import (
	"encoding/json"
	"log"
	"net/http"

	"proxy/internal/pubsub"
	"proxy/internal/store"
)

// Handler holds the dependencies needed for task management endpoints.
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
	tasks, err := h.Store.ListTasksByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("list tasks error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")
	task, err := h.Store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}
	if task.Status != "pending_approval" {
		http.Error(w, `{"error":"task is not pending approval"}`, http.StatusBadRequest)
		return
	}
	user, err := h.Store.GetUserByOrgID(r.Context(), task.OrgID)
	if err != nil {
		http.Error(w, `{"error":"org not found"}`, http.StatusNotFound)
		return
	}

	token, err := pubsub.GetRepairSAToken(r.Context(), h.RepairSA)
	if err != nil {
		log.Printf("approve: get token error: %v", err)
		http.Error(w, `{"error":"could not get publish token"}`, http.StatusInternalServerError)
		return
	}

	// Stamp all original attributes through plus our repaired marker.
	outAttrs := make(map[string]string, len(task.Attributes)+1)
	for k, v := range task.Attributes {
		if k != "_deadlift_confidence" {
			outAttrs[k] = v
		}
	}
	outAttrs["_deadlift_repaired"] = "true"

	payload := task.FixedPayload
	if payload == "" {
		payload = task.RawPayload // fall back to original if LLM produced nothing
	}

	if err := pubsub.PublishMessage(r.Context(), token, user.MainTopic, payload, outAttrs); err != nil {
		log.Printf("approve: publish error: %v", err)
		http.Error(w, `{"error":"publish failed: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if err := h.Store.UpdateTaskStatus(r.Context(), taskID, "approved"); err != nil {
		log.Printf("approve: update task error: %v", err)
	}

	log.Printf("approve: republished task %s to topic %s", taskID, user.MainTopic)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (h *Handler) HandleDeny(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")
	task, err := h.Store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}
	if task.Status != "pending_approval" {
		http.Error(w, `{"error":"task is not pending approval"}`, http.StatusBadRequest)
		return
	}
	if err := h.Store.UpdateTaskStatus(r.Context(), taskID, "denied"); err != nil {
		log.Printf("deny: update task error: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
