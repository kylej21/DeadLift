package graphrag

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	rp *httputil.ReverseProxy
}

func New(serverURL string) *Proxy {
	if serverURL == "" {
		return &Proxy{}
	}
	target, err := url.Parse(serverURL)
	if err != nil {
		panic("invalid GRAPHRAG_SERVER_URL: " + err.Error())
	}
	return &Proxy{rp: httputil.NewSingleHostReverseProxy(target)}
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
