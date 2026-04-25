package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
)

var (
	clientID          = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret      = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI       = os.Getenv("REDIRECT_URI")
	repairSA          = os.Getenv("REPAIR_SA_EMAIL")
	clientURL         = os.Getenv("CLIENT_URL")
	gcpProject        = os.Getenv("GCP_PROJECT_ID")
	graphragServerURL = os.Getenv("GRAPHRAG_SERVER_URL")

	fsClient   *firestore.Client
	stateStore sync.Map
)

func corsAllowed(origin string) bool {
	if origin == clientURL {
		return true
	}
	// Allow any localhost origin for local development
	return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if corsAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()

	var err error
	fsClient, err = firestore.NewClient(ctx, gcpProject)
	if err != nil {
		log.Fatalf("firestore init: %v", err)
	}
	defer fsClient.Close()

	initGraphragProxy()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/onboard/connect", handleConnect)
	mux.HandleFunc("GET /api/onboard/signin", handleSignIn)
	mux.HandleFunc("GET /api/onboard/callback", handleCallback)
	mux.HandleFunc("GET /api/users", handleGetUser)
	mux.HandleFunc("GET /api/tasks", handleListTasks)
	mux.HandleFunc("POST /api/tasks/{task_id}/approve", handleApproveTask)
	mux.HandleFunc("POST /api/tasks/{task_id}/deny", handleDenyTask)
	mux.HandleFunc("POST /api/graphrag/onboard", handleGraphragOnboard)
	mux.HandleFunc("POST /api/graphrag/update", handleGraphragUpdate)
	mux.HandleFunc("GET /api/graphrag/status/{job_id}", handleGraphragStatus)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go startWorker(ctx)

	log.Printf("proxy listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(mux)))
}
