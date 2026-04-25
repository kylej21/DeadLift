package onboard

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"proxy/internal/models"
	"proxy/internal/pubsub"
	"proxy/internal/store"
)

const googleAuthURL = "https://accounts.google.com/o/oauth2/v2/auth"

// Config holds all dependencies the onboarding handlers need.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	RepairSA     string
	ClientURL    string
	GCPProject   string
	Store        *store.Store
	StateStore   *sync.Map
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (c *Config) oauthURL(state, scope string) string {
	params := url.Values{
		"client_id":     {c.ClientID},
		"redirect_uri":  {c.RedirectURI},
		"response_type": {"code"},
		"scope":         {scope},
		"state":         {state},
		"access_type":   {"online"},
		"prompt":        {"select_account"},
	}
	return googleAuthURL + "?" + params.Encode()
}

func (c *Config) HandleConnect(w http.ResponseWriter, r *http.Request) {
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
	c.StateStore.Store(state, models.StatePayload{
		Mode:              "onboard",
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

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]string{
		"oauth_url": c.oauthURL(state, "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email"),
	})
}

func (c *Config) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	state, err := generateState()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	c.StateStore.Store(state, models.StatePayload{Mode: "signin"})
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]string{
		"oauth_url": c.oauthURL(state, "https://www.googleapis.com/auth/userinfo.email"),
	})
}

func (c *Config) HandleCallback(w http.ResponseWriter, r *http.Request) {
	redirect := func(errMsg string) {
		http.Redirect(w, r, c.ClientURL+"/#/onboarding?error="+url.QueryEscape(errMsg), http.StatusTemporaryRedirect)
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		redirect(errParam)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	val, ok := c.StateStore.LoadAndDelete(state)
	if !ok {
		redirect("invalid_or_expired_session")
		return
	}
	payload := val.(models.StatePayload)

	accessToken, info, err := exchangeCode(c.ClientID, c.ClientSecret, c.RedirectURI, code)
	if err != nil {
		log.Printf("token exchange error: %v", err)
		redirect("token_exchange_failed")
		return
	}

	// Sign-in mode: look up existing user and redirect.
	if payload.Mode == "signin" {
		user, err := c.Store.GetUserByGoogleSub(r.Context(), info.Sub)
		if err != nil {
			log.Printf("signin: user not found for sub %s: %v", info.Sub, err)
			redirect("user_not_found")
			return
		}
		http.Redirect(w, r, c.ClientURL+"/#/app?org_id="+user.OrgID, http.StatusTemporaryRedirect)
		return
	}

	// Onboard mode: grant IAM, health-check, set up BQ, create user.
	if err := grantPermissions(r.Context(), accessToken, payload.ProjectID, payload.DLQSubscription, payload.MainTopic, c.RepairSA); err != nil {
		log.Printf("IAM grant error: %v", err)
		redirect("iam_grant_failed: " + err.Error())
		return
	}

	if err := healthCheck(r.Context(), accessToken, payload.ProjectID, payload.DLQSubscription, payload.MainTopic); err != nil {
		log.Printf("health check failed: %v", err)
		redirect("health_check_failed: " + err.Error())
		return
	}

	// BigQuery streaming subscription (non-fatal).
	projectNumber, err := c.Store.GetProjectNumber(r.Context(), accessToken, payload.ProjectID)
	if err != nil {
		log.Printf("bigquery setup: could not get project number: %v", err)
	} else {
		proxyToken, _ := pubsub.GetMetadataToken(r.Context())
		pubsubSA := fmt.Sprintf("service-%s@gcp-sa-pubsub.iam.gserviceaccount.com", projectNumber)
		if err := c.Store.GrantPubSubSABQAccess(r.Context(), proxyToken, pubsubSA); err != nil {
			log.Printf("bigquery setup: grant access error: %v", err)
		} else if err := c.Store.CreateBQSubscription(r.Context(), accessToken, payload.ProjectID, payload.MainTopic, payload.OrgID); err != nil {
			log.Printf("bigquery setup: create subscription error: %v", err)
		} else {
			log.Printf("bigquery setup: streaming subscription created for org %s", payload.OrgID)
		}
	}

	user := models.User{
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

	if err := c.Store.CreateUser(r.Context(), user); err != nil {
		log.Printf("firestore write error: %v", err)
		redirect("db_write_failed")
		return
	}

	http.Redirect(w, r, c.ClientURL+"/#/app?org_id="+payload.OrgID, http.StatusTemporaryRedirect)
}

func (c *Config) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, `{"error":"org_id required"}`, http.StatusBadRequest)
		return
	}
	user, err := c.Store.GetUserByOrgID(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
