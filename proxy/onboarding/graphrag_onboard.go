package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

var graphragProxy *httputil.ReverseProxy

func initGraphragProxy() {
	if graphragServerURL == "" {
		return
	}
	target, err := url.Parse(graphragServerURL)
	if err != nil {
		panic("invalid GRAPHRAG_SERVER_URL: " + err.Error())
	}
	graphragProxy = httputil.NewSingleHostReverseProxy(target)
}

func handleGraphragOnboard(w http.ResponseWriter, r *http.Request) {
	if graphragProxy == nil {
		http.Error(w, `{"error":"graphrag server not configured"}`, http.StatusServiceUnavailable)
		return
	}
	graphragProxy.ServeHTTP(w, r)
}

func handleGraphragUpdate(w http.ResponseWriter, r *http.Request) {
	if graphragProxy == nil {
		http.Error(w, `{"error":"graphrag server not configured"}`, http.StatusServiceUnavailable)
		return
	}
	graphragProxy.ServeHTTP(w, r)
}

func handleGraphragStatus(w http.ResponseWriter, r *http.Request) {
	if graphragProxy == nil {
		http.Error(w, `{"error":"graphrag server not configured"}`, http.StatusServiceUnavailable)
		return
	}
	graphragProxy.ServeHTTP(w, r)
}
