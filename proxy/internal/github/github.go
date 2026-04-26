package github

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Handler manages GitHub OAuth state and token exchange.
type Handler struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	ClientURL    string
	tokens       sync.Map // state → access_token string
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleAuthURL handles GET /api/github/auth-url.
// It generates a random state, stores it, and returns the GitHub OAuth URL.
func (h *Handler) HandleAuthURL(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// Store state → "" placeholder (token not yet obtained)
	h.tokens.Store(state, "")

	params := url.Values{
		"client_id":    {h.ClientID},
		"redirect_uri": {h.RedirectURI},
		"scope":        {"repo"},
		"state":        {state},
	}
	oauthURL := "https://github.com/login/oauth/authorize?" + params.Encode()

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]string{
		"oauth_url": oauthURL,
		"state_id":  state,
	})
}

// HandleCallback handles GET /api/github/callback?code=...&state=...
// It validates the state, exchanges the code for an access token, and redirects.
func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Validate that the state exists (use Load, not LoadAndDelete — GetToken handles deletion)
	if _, ok := h.tokens.Load(state); !ok {
		http.Error(w, `{"error":"invalid or expired state"}`, http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	accessToken, err := exchangeCode(h.ClientID, h.ClientSecret, h.RedirectURI, code)
	if err != nil {
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusInternalServerError)
		return
	}

	// Store the real token
	h.tokens.Store(state, accessToken)

	http.Redirect(w, r, h.ClientURL+"/github_connected.html?state_id="+url.QueryEscape(state), http.StatusTemporaryRedirect)
}

// GetToken retrieves and consumes the access token associated with stateID.
// Returns ("", false) if not found or if the token has not yet been obtained.
func (h *Handler) GetToken(stateID string) (string, bool) {
	val, ok := h.tokens.Load(stateID)
	if !ok {
		return "", false
	}
	token, _ := val.(string)
	// Delete after consuming to prevent reuse
	h.tokens.Delete(stateID)
	if token == "" {
		return "", false
	}
	return token, true
}

// exchangeCode exchanges a GitHub OAuth code for an access token.
func exchangeCode(clientID, clientSecret, redirectURI, code string) (string, error) {
	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token",
		strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", &oauthError{result.Error}
	}
	return result.AccessToken, nil
}

type oauthError struct{ msg string }

func (e *oauthError) Error() string { return "github oauth error: " + e.msg }
