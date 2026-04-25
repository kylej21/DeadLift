package graphrag

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	serverURL string
	rp        *httputil.ReverseProxy
}

func New(serverURL string) *Proxy {
	if serverURL == "" {
		return &Proxy{}
	}
	target, err := url.Parse(serverURL)
	if err != nil {
		panic("invalid GRAPHRAG_SERVER_URL: " + err.Error())
	}
	return &Proxy{serverURL: serverURL, rp: httputil.NewSingleHostReverseProxy(target)}
}

// TriggerOnboard fires a POST /onboard to the graphrag server for the given org.
// Intended to be called in a goroutine — logs errors but does not return them.
func (p *Proxy) TriggerOnboard(orgID, repoURL string) {
	if p.serverURL == "" {
		log.Printf("graphrag: skipping onboard trigger for org %s — server not configured", orgID)
		return
	}
	body, _ := json.Marshal(map[string]string{
		"repo_url":  repoURL,
		"client_id": orgID,
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
func (p *Proxy) HandleUpdate(w http.ResponseWriter, r *http.Request)  { p.serve(w, r) }
func (p *Proxy) HandleStatus(w http.ResponseWriter, r *http.Request)  { p.serve(w, r) }
