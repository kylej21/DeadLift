package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
)

var (
	clientID     = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI  = os.Getenv("REDIRECT_URI")
	repairSA     = os.Getenv("REPAIR_SA_EMAIL")
	clientURL    = os.Getenv("CLIENT_URL")
	gcpProject   = os.Getenv("GCP_PROJECT_ID")

	fsClient   *firestore.Client
	stateStore sync.Map
)

func main() {
	ctx := context.Background()

	var err error
	fsClient, err = firestore.NewClient(ctx, gcpProject)
	if err != nil {
		log.Fatalf("firestore init: %v", err)
	}
	defer fsClient.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/onboard/connect", handleConnect)
	mux.HandleFunc("GET /api/onboard/callback", handleCallback)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("proxy listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
