package main

import "time"

type User struct {
	OrgID             string          `firestore:"org_id"`
	GoogleSub         string          `firestore:"google_sub"`
	Email             string          `firestore:"email"`
	ProjectID         string          `firestore:"project_id"`
	DLQSubscription   string          `firestore:"dlq_subscription"`
	MainTopic         string          `firestore:"main_topic"`
	RepairSAGranted   bool            `firestore:"repair_sa_granted"`
	AutoRepublish     map[string]bool `firestore:"auto_republish"`
	BatchingThreshold int             `firestore:"batching_threshold"`
	NotificationEmail string          `firestore:"notification_email"`
	CreatedAt         time.Time       `firestore:"created_at"`
}

type statePayload struct {
	OrgID             string
	ProjectID         string
	DLQSubscription   string
	MainTopic         string
	AutoRepublish     map[string]bool
	BatchingThreshold int
	NotificationEmail string
}

type userInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}
