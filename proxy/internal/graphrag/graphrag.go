package graphrag

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"proxy/internal/store"
)

type Proxy struct {
	serverURL string
	rp        *httputil.ReverseProxy
	store     *store.Store
}

func New(serverURL string, st *store.Store) *Proxy {
	if serverURL == "" {
		return &Proxy{store: st}
	}
	target, err := url.Parse(serverURL)
	if err != nil {
		panic("invalid GRAPHRAG_SERVER_URL: " + err.Error())
	}
	return &Proxy{serverURL: serverURL, rp: httputil.NewSingleHostReverseProxy(target), store: st}
}

// TriggerOnboard fires a POST /onboard to the graphrag server for the given org.
// Intended to be called in a goroutine — logs errors but does not return them.
func (p *Proxy) TriggerOnboard(orgID, repoURL, githubToken string) {
	if p.serverURL == "" {
		log.Printf("graphrag: skipping onboard trigger for org %s — server not configured", orgID)
		return
	}
	body, _ := json.Marshal(map[string]string{
		"repo_url":     repoURL,
		"client_id":    orgID,
		"github_token": githubToken,
	})
	resp, err := http.Post(p.serverURL+"/onboard", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("graphrag: trigger onboard for org %s: %v", orgID, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("graphrag: triggered onboard for org %s, status %d", orgID, resp.StatusCode)
}

func (p *Proxy) serve(w http.ResponseWriter, r *http.Request) {
	if p.rp == nil {
		http.Error(w, `{"error":"graphrag server not configured"}`, http.StatusServiceUnavailable)
		return
	}
	p.rp.ServeHTTP(w, r)
}

func (p *Proxy) HandleOnboard(w http.ResponseWriter, r *http.Request) { p.serve(w, r) }
func (p *Proxy) HandleStatus(w http.ResponseWriter, r *http.Request)  { p.serve(w, r) }

func (p *Proxy) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if p.rp == nil {
		http.Error(w, `{"error":"graphrag server not configured"}`, http.StatusServiceUnavailable)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if p.store != nil {
		clientID, _ := req["client_id"].(string)
		if clientID != "" {
			user, err := p.store.GetUserByOrgID(r.Context(), clientID)
			if err == nil && user.GithubToken != "" {
				req["github_token"] = user.GithubToken
			}
		}
	}

	newBody, _ := json.Marshal(req)
	r.Body = io.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))
	p.rp.ServeHTTP(w, r)
}
