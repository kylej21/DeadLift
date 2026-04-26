package rca

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"proxy/internal/mcp"
	"proxy/internal/store"
)

type Handler struct {
	Store         *store.Store
	MCPClient     *mcp.Client
	GraphragURL   string
	httpClient    *http.Client
}

func New(s *store.Store, m *mcp.Client, graphragURL string) *Handler {
	return &Handler{
		Store:       s,
		MCPClient:   m,
		GraphragURL: graphragURL,
		httpClient:  &http.Client{Timeout: 120 * time.Second},
	}
}

const rcaSystemPrompt = `You are a root cause analysis expert for Pub/Sub pipeline failures.

Given a failed message and its repaired version, produce a detailed root cause analysis in plain text covering:
1. What went wrong and why
2. Which upstream system or code path likely caused this
3. How the fix addresses the root cause
4. Recommendations to prevent recurrence

Be specific and technical. Reference field names and values from the payload.`

func (h *Handler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")
	if taskID == "" {
		http.Error(w, `{"error":"task_id required"}`, http.StatusBadRequest)
		return
	}

	task, err := h.Store.GetTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}

	analysis, err := h.MCPClient.CallRCA(r.Context(), task.OrgID, task.MessageID, task.RawPayload, task.FixedPayload, task.ErrorClass)
	if err != nil {
		log.Printf("rca: vllm error for task %s: %v", taskID, err)
		http.Error(w, `{"error":"analysis failed"}`, http.StatusInternalServerError)
		return
	}

	if err := h.persist(r.Context(), task.OrgID, task.MessageID, task.ErrorClass, task.RawPayload, task.FixedPayload, analysis); err != nil {
		log.Printf("rca: persist error for task %s: %v", taskID, err)
		// non-fatal — still return the analysis to the client
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"task_id":  taskID,
		"analysis": analysis,
	})
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.httpClient.Get(h.GraphragURL + "/rca/get/" + orgID)
	if err != nil {
		log.Printf("rca: fetch from graphrag error: %v", err)
		http.Error(w, `{"error":"could not fetch reports"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (h *Handler) persist(ctx context.Context, orgID, messageID, errorClass, rawPayload, fixedPayload, analysis string) error {
	payload, _ := json.Marshal(map[string]string{
		"org_id":        orgID,
		"message_id":    messageID,
		"error_class":   errorClass,
		"raw_payload":   rawPayload,
		"fixed_payload": fixedPayload,
		"analysis":      analysis,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", h.GraphragURL+"/rca/create", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post to graphrag: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("graphrag HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}
