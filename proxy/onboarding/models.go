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
	GithubURL         string          `firestore:"github_url"`
	WebURL            string          `firestore:"web_url"`
	CreatedAt         time.Time       `firestore:"created_at"`
}

type statePayload struct {
	Mode              string // "onboard" | "signin"
	OrgID             string
	ProjectID         string
	DLQSubscription   string
	MainTopic         string
	AutoRepublish     map[string]bool
	BatchingThreshold int
	NotificationEmail string
	GithubURL         string
	WebURL            string
}

type userInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

type Task struct {
	TaskID       string            `firestore:"task_id"`
	OrgID        string            `firestore:"org_id"`
	MessageID    string            `firestore:"message_id"`
	RawPayload   string            `firestore:"raw_payload"`
	Attributes   map[string]string `firestore:"attributes"`
	FixedPayload string            `firestore:"fixed_payload"`
	Status       string            `firestore:"status"` // pending_approval | approved | denied | failed
	CreatedAt    time.Time         `firestore:"created_at"`
	UpdatedAt    time.Time         `firestore:"updated_at"`
}
