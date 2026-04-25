package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"proxy/internal/models"
)

type Store struct {
	fs         *firestore.Client
	gcpProject string
}

func New(fs *firestore.Client, gcpProject string) *Store {
	return &Store{fs: fs, gcpProject: gcpProject}
}

// ── Users ────────────────────────────────────────────────────────────────────

func (s *Store) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.fs.Collection("users").Doc(user.OrgID).Set(ctx, user)
	return err
}

func (s *Store) GetUserByOrgID(ctx context.Context, orgID string) (*models.User, error) {
	doc, err := s.fs.Collection("users").Doc(orgID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &user, nil
}

func (s *Store) GetUserByGoogleSub(ctx context.Context, sub string) (*models.User, error) {
	docs, err := s.fs.Collection("users").Where("google_sub", "==", sub).Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("get user by sub: %w", err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	var user models.User
	if err := docs[0].DataTo(&user); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &user, nil
}

func (s *Store) GetAllUsers(ctx context.Context) ([]models.User, error) {
	docs, err := s.fs.Collection("users").Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]models.User, 0, len(docs))
	for _, doc := range docs {
		var u models.User
		if err := doc.DataTo(&u); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// ── Tasks ────────────────────────────────────────────────────────────────────

func (s *Store) CreateTask(ctx context.Context, task models.Task) error {
	_, err := s.fs.Collection("tasks").Doc(task.TaskID).Set(ctx, task)
	return err
}

func (s *Store) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	doc, err := s.fs.Collection("tasks").Doc(taskID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	var task models.Task
	if err := doc.DataTo(&task); err != nil {
		return nil, fmt.Errorf("parse task: %w", err)
	}
	return &task, nil
}

func (s *Store) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	_, err := s.fs.Collection("tasks").Doc(taskID).Update(ctx, []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

func (s *Store) ListTasksByOrg(ctx context.Context, orgID string) ([]models.Task, error) {
	docs, err := s.fs.Collection("tasks").Where("org_id", "==", orgID).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	tasks := make([]models.Task, 0, len(docs))
	for _, doc := range docs {
		var t models.Task
		if err := doc.DataTo(&t); err == nil {
			tasks = append(tasks, t)
		}
	}
	// Sort newest first in Go — avoids needing a composite Firestore index.
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[j].CreatedAt.After(tasks[i].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}
	return tasks, nil
}

// ── BigQuery subscription setup ──────────────────────────────────────────────

func (s *Store) GetProjectNumber(ctx context.Context, token, projectID string) (string, error) {
	url := fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v1/projects/%s", projectID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get project HTTP %d: %s", resp.StatusCode, b)
	}
	var result struct {
		ProjectNumber string `json:"projectNumber"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return "", err
	}
	return result.ProjectNumber, nil
}

func (s *Store) GrantPubSubSABQAccess(ctx context.Context, proxyToken, pubsubSAEmail string) error {
	datasetURL := fmt.Sprintf(
		"https://bigquery.googleapis.com/bigquery/v2/projects/%s/datasets/deadlift",
		s.gcpProject,
	)
	getReq, err := http.NewRequestWithContext(ctx, "GET", datasetURL, nil)
	if err != nil {
		return err
	}
	getReq.Header.Set("Authorization", "Bearer "+proxyToken)
	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return err
	}
	defer getResp.Body.Close()
	b, _ := io.ReadAll(getResp.Body)
	if getResp.StatusCode != http.StatusOK {
		return fmt.Errorf("get dataset HTTP %d: %s", getResp.StatusCode, b)
	}
	var dataset map[string]interface{}
	if err := json.Unmarshal(b, &dataset); err != nil {
		return err
	}
	access, _ := dataset["access"].([]interface{})
	newEntries := []map[string]interface{}{
		{"role": "roles/bigquery.dataEditor", "iamMember": "serviceAccount:" + pubsubSAEmail},
		{"role": "roles/bigquery.metadataViewer", "iamMember": "serviceAccount:" + pubsubSAEmail},
	}
	for _, newEntry := range newEntries {
		found := false
		for _, entry := range access {
			if e, ok := entry.(map[string]interface{}); ok {
				if e["iamMember"] == newEntry["iamMember"] && e["role"] == newEntry["role"] {
					found = true
					break
				}
			}
		}
		if !found {
			access = append(access, newEntry)
		}
	}
	dataset["access"] = access
	patchBody, _ := json.Marshal(map[string]interface{}{"access": dataset["access"]})
	patchReq, err := http.NewRequestWithContext(ctx, "PATCH", datasetURL, bytes.NewReader(patchBody))
	if err != nil {
		return err
	}
	patchReq.Header.Set("Authorization", "Bearer "+proxyToken)
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		return err
	}
	defer patchResp.Body.Close()
	if patchResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(patchResp.Body)
		return fmt.Errorf("patch dataset HTTP %d: %s", patchResp.StatusCode, b)
	}
	return nil
}

func (s *Store) CreateBQSubscription(ctx context.Context, token, customerProjectID, mainTopic, orgID string) error {
	topicResource := mainTopic
	if len(mainTopic) < 9 || mainTopic[:9] != "projects/" {
		topicResource = fmt.Sprintf("projects/%s/topics/%s", customerProjectID, mainTopic)
	}
	subName := fmt.Sprintf("projects/%s/subscriptions/deadlift-analytics-%s", customerProjectID, orgID)
	table := fmt.Sprintf("%s:deadlift.success_logs", s.gcpProject)

	mkBody := func() []byte {
		b, _ := json.Marshal(map[string]interface{}{
			"topic": topicResource,
			"bigqueryConfig": map[string]interface{}{
				"table":         table,
				"writeMetadata": true,
			},
		})
		return b
	}

	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "PUT",
			"https://pubsub.googleapis.com/v1/"+subName,
			bytes.NewReader(mkBody()),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
			return nil
		}
		lastErr = fmt.Errorf("create BQ subscription HTTP %d: %s", resp.StatusCode, b)
		time.Sleep(time.Duration(attempt*15) * time.Second)
	}
	return lastErr
}
