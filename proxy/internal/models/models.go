package models

import "time"

type User struct {
	OrgID             string          `firestore:"org_id"             json:"org_id"`
	GoogleSub         string          `firestore:"google_sub"         json:"google_sub"`
	Email             string          `firestore:"email"              json:"email"`
	ProjectID         string          `firestore:"project_id"         json:"project_id"`
	DLQSubscription   string          `firestore:"dlq_subscription"   json:"dlq_subscription"`
	MainTopic         string          `firestore:"main_topic"         json:"main_topic"`
	RepairSAGranted   bool            `firestore:"repair_sa_granted"  json:"repair_sa_granted"`
	AutoRepublish     map[string]bool `firestore:"auto_republish"     json:"auto_republish"`
	BatchingThreshold int             `firestore:"batching_threshold" json:"batching_threshold"`
	NotificationEmail string          `firestore:"notification_email" json:"notification_email"`
	GithubURL         string          `firestore:"github_url"         json:"github_url"`
	WebURL            string          `firestore:"web_url"            json:"web_url"`
	CreatedAt         time.Time       `firestore:"created_at"         json:"created_at"`
}

type Task struct {
	TaskID       string            `firestore:"task_id"       json:"task_id"`
	OrgID        string            `firestore:"org_id"        json:"org_id"`
	MessageID    string            `firestore:"message_id"    json:"message_id"`
	RawPayload   string            `firestore:"raw_payload"   json:"raw_payload"`
	Attributes   map[string]string `firestore:"attributes"    json:"attributes"`
	FixedPayload string            `firestore:"fixed_payload" json:"fixed_payload"`
	ErrorClass   string            `firestore:"error_class"   json:"error_class"`
	Status       string            `firestore:"status"        json:"status"` // pending_approval | approved | denied | failed
	CreatedAt    time.Time         `firestore:"created_at"    json:"created_at"`
	UpdatedAt    time.Time         `firestore:"updated_at"    json:"updated_at"`
}

type StatePayload struct {
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

type UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

type PubSubMessage struct {
	AckID   string `json:"ackId"`
	Message struct {
		Data        string            `json:"data"`
		Attributes  map[string]string `json:"attributes"`
		MessageID   string            `json:"messageId"`
		PublishTime string            `json:"publishTime"`
	} `json:"message"`
}
