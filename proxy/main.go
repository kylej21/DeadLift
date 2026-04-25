package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"

	"proxy/internal/graphrag"
	"proxy/internal/onboard"
	"proxy/internal/store"
	"proxy/internal/tasks"
	"proxy/internal/worker"
)

func corsAllowed(origin, clientURL string) bool {
	if origin == clientURL {
		return true
	}
	return strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
}

func corsMiddleware(clientURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if corsAllowed(origin, clientURL) {
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

	gcpProject := os.Getenv("GCP_PROJECT_ID")
	clientURL := os.Getenv("CLIENT_URL")
	repairSA := os.Getenv("REPAIR_SA_EMAIL")

	fs, err := firestore.NewClient(ctx, gcpProject)
	if err != nil {
		log.Fatalf("firestore init: %v", err)
	}
	defer fs.Close()

	var stateStore sync.Map
	st := store.New(fs, gcpProject)

	gr := graphrag.New(os.Getenv("GRAPHRAG_SERVER_URL"))

	ob := &onboard.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("REDIRECT_URI"),
		RepairSA:     repairSA,
		ClientURL:    clientURL,
		GCPProject:   gcpProject,
		Store:        st,
		StateStore:   &stateStore,
		Graphrag:     gr,
	}

	th := &tasks.Handler{
		RepairSA: repairSA,
		Store:    st,
	}

	w := &worker.Worker{
		RepairSA: repairSA,
		Store:    st,
	}

	mux := http.NewServeMux()

	// Onboarding & auth
	mux.HandleFunc("POST /api/onboard/connect", ob.HandleConnect)
	mux.HandleFunc("GET /api/onboard/signin", ob.HandleSignIn)
	mux.HandleFunc("GET /api/onboard/callback", ob.HandleCallback)
	mux.HandleFunc("GET /api/users", ob.HandleGetUser)

	// Task management (DLQ repair & republish)
	mux.HandleFunc("GET /api/tasks", th.HandleList)
	mux.HandleFunc("POST /api/tasks/{task_id}/approve", th.HandleApprove)
	mux.HandleFunc("POST /api/tasks/{task_id}/deny", th.HandleDeny)

	// GraphRAG context ingestion
	mux.HandleFunc("POST /api/graphrag/onboard", gr.HandleOnboard)
	mux.HandleFunc("POST /api/graphrag/update", gr.HandleUpdate)
	mux.HandleFunc("GET /api/graphrag/status/{job_id}", gr.HandleStatus)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	go w.Start(ctx)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("proxy listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(clientURL, mux)))
}
