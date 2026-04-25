package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	oauthScope    = "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email"
)

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		ProjectID         string          `json:"project_id"`
		DLQSubscription   string          `json:"dlq_subscription"`
		MainTopic         string          `json:"main_topic"`
		AutoRepublish     map[string]bool `json:"auto_republish"`
		BatchingThreshold int             `json:"batching_threshold"`
		NotificationEmail string          `json:"notification_email"`
		GithubURL         string          `json:"github_url"`
		WebURL            string          `json:"web_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.ProjectID = strings.TrimSpace(req.ProjectID)
	req.DLQSubscription = strings.TrimSpace(req.DLQSubscription)
	req.MainTopic = strings.TrimSpace(req.MainTopic)

	if req.ProjectID == "" || req.DLQSubscription == "" || req.MainTopic == "" {
		http.Error(w, `{"error":"project_id, dlq_subscription, and main_topic are required"}`, http.StatusBadRequest)
		return
	}
	if req.BatchingThreshold == 0 {
		req.BatchingThreshold = 5
	}

	state, err := generateState()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	stateStore.Store(state, statePayload{
		OrgID:             uuid.New().String(),
		ProjectID:         req.ProjectID,
		DLQSubscription:   req.DLQSubscription,
		MainTopic:         req.MainTopic,
		AutoRepublish:     req.AutoRepublish,
		BatchingThreshold: req.BatchingThreshold,
		NotificationEmail: req.NotificationEmail,
		GithubURL:         req.GithubURL,
		WebURL:            req.WebURL,
	})

	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {oauthScope},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]string{
		"oauth_url": googleAuthURL + "?" + params.Encode(),
	})
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	redirect := func(errMsg string) {
		http.Redirect(w, r, clientURL+"/#/onboarding?error="+url.QueryEscape(errMsg), http.StatusTemporaryRedirect)
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		redirect(errParam)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	val, ok := stateStore.LoadAndDelete(state)
	if !ok {
		redirect("invalid_or_expired_session")
		return
	}
	payload := val.(statePayload)

	accessToken, info, err := exchangeCode(code)
	if err != nil {
		log.Printf("token exchange error: %v", err)
		redirect("token_exchange_failed")
		return
	}

	if _, err := grantPermissions(r.Context(), accessToken, payload.ProjectID, payload.DLQSubscription, payload.MainTopic, repairSA); err != nil {
		log.Printf("IAM grant error: %v", err)
		redirect("iam_grant_failed: " + err.Error())
		return
	}

	if err := healthCheck(r.Context(), accessToken, payload.ProjectID, payload.DLQSubscription, payload.MainTopic); err != nil {
		log.Printf("health check failed: %v", err)
		redirect("health_check_failed: " + err.Error())
		return
	}

	// Set up BigQuery streaming subscription on the customer's main topic.
	// Non-fatal — log and continue if it fails.
	projectNumber, err := getProjectNumber(r.Context(), accessToken, payload.ProjectID)
	if err != nil {
		log.Printf("bigquery setup: could not get project number: %v", err)
	} else {
		pubsubSA := fmt.Sprintf("service-%s@gcp-sa-pubsub.iam.gserviceaccount.com", projectNumber)
		if err := grantPubSubSABQAccess(r.Context(), pubsubSA); err != nil {
			log.Printf("bigquery setup: grant access error: %v", err)
		} else if err := createBQSubscription(r.Context(), accessToken, payload.ProjectID, payload.MainTopic, payload.OrgID); err != nil {
			log.Printf("bigquery setup: create subscription error: %v", err)
		} else {
			log.Printf("bigquery setup: streaming subscription created for org %s", payload.OrgID)
		}
	}

	user := User{
		OrgID:             payload.OrgID,
		GoogleSub:         info.Sub,
		Email:             info.Email,
		ProjectID:         payload.ProjectID,
		DLQSubscription:   payload.DLQSubscription,
		MainTopic:         payload.MainTopic,
		RepairSAGranted:   true,
		AutoRepublish:     payload.AutoRepublish,
		BatchingThreshold: payload.BatchingThreshold,
		NotificationEmail: payload.NotificationEmail,
		GithubURL:         payload.GithubURL,
		WebURL:            payload.WebURL,
		CreatedAt:         time.Now(),
	}

	if err := createUser(r.Context(), user); err != nil {
		log.Printf("firestore write error: %v", err)
		redirect("db_write_failed")
		return
	}

	http.Redirect(w, r, clientURL+"/#/app?org_id="+payload.OrgID, http.StatusTemporaryRedirect)
}

func handleListTasks(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}
	tasks, err := listTasksByOrg(r.Context(), orgID)
	if err != nil {
		log.Printf("list tasks error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func handleApproveTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")
	task, err := getTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}
	if task.Status != "pending_approval" {
		http.Error(w, `{"error":"task is not pending approval"}`, http.StatusBadRequest)
		return
	}
	user, err := getUserByOrgID(r.Context(), task.OrgID)
	if err != nil {
		http.Error(w, `{"error":"org not found"}`, http.StatusNotFound)
		return
	}
	token, err := getRepairSAToken(r.Context())
	if err != nil {
		log.Printf("approve: get token error: %v", err)
		http.Error(w, `{"error":"could not get publish token"}`, http.StatusInternalServerError)
		return
	}
	outAttrs := make(map[string]string, len(task.Attributes)+1)
	for k, v := range task.Attributes {
		outAttrs[k] = v
	}
	outAttrs["_deadlift_repaired"] = "true"
	if err := publishMessage(r.Context(), token, user.MainTopic, task.FixedPayload, outAttrs); err != nil {
		log.Printf("approve: publish error: %v", err)
		http.Error(w, `{"error":"publish failed"}`, http.StatusInternalServerError)
		return
	}
	if err := updateTaskStatus(r.Context(), taskID, "approved"); err != nil {
		log.Printf("approve: update task error: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func handleDenyTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")
	task, err := getTask(r.Context(), taskID)
	if err != nil {
		http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		return
	}
	if task.Status != "pending_approval" {
		http.Error(w, `{"error":"task is not pending approval"}`, http.StatusBadRequest)
		return
	}
	if err := updateTaskStatus(r.Context(), taskID, "denied"); err != nil {
		log.Printf("deny: update task error: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
