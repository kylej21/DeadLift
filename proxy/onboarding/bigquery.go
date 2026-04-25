package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// getProjectNumber resolves a GCP project ID → project number using the customer's token.
// The Pub/Sub service agent email is service-{number}@gcp-sa-pubsub.iam.gserviceaccount.com
func getProjectNumber(ctx context.Context, token, projectID string) (string, error) {
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

// grantPubSubSABQAccess grants the customer's Pub/Sub service agent
// bigquery.dataEditor on our dataset so it can write to success_logs.
// Uses the proxy's own ADC (metadata server) since this modifies our project.
func grantPubSubSABQAccess(ctx context.Context, pubsubSAEmail string) error {
	proxyToken, err := getMetadataToken(ctx)
	if err != nil {
		return fmt.Errorf("get proxy token: %w", err)
	}

	datasetURL := fmt.Sprintf(
		"https://bigquery.googleapis.com/bigquery/v2/projects/%s/datasets/deadlift",
		gcpProject,
	)

	// GET current dataset to read existing access entries.
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

	// Add the new member if not already present.
	access, _ := dataset["access"].([]interface{})
	newEntry := map[string]interface{}{
		"role":      "roles/bigquery.dataEditor",
		"iamMember": "serviceAccount:" + pubsubSAEmail,
	}
	for _, entry := range access {
		if e, ok := entry.(map[string]interface{}); ok {
			if e["iamMember"] == newEntry["iamMember"] {
				return nil // already granted
			}
		}
	}
	dataset["access"] = append(access, newEntry)

	// PATCH the dataset with updated access.
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

// createBQSubscription creates a BigQuery Pub/Sub subscription on the customer's
// main topic that streams directly into our success_logs table.
// Named deadlift-analytics-{orgID} so we can identify the tenant in BigQuery queries.
func createBQSubscription(ctx context.Context, token, customerProjectID, mainTopic, orgID string) error {
	// mainTopic may already be a full resource path or just a name.
	topicResource := mainTopic
	if !strings.HasPrefix(mainTopic, "projects/") {
		topicResource = fmt.Sprintf("projects/%s/topics/%s", customerProjectID, mainTopic)
	}

	subName := fmt.Sprintf("projects/%s/subscriptions/deadlift-analytics-%s", customerProjectID, orgID)
	table := fmt.Sprintf("projects/%s/datasets/deadlift/tables/success_logs", gcpProject)

	body, _ := json.Marshal(map[string]interface{}{
		"topic": topicResource,
		"bigqueryConfig": map[string]interface{}{
			"table":         table,
			"writeMetadata": true,
		},
	})

	req, err := http.NewRequestWithContext(ctx, "PUT",
		"https://pubsub.googleapis.com/v1/"+subName,
		bytes.NewReader(body),
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
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	// 409 = already exists, that's fine.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("create BQ subscription HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}
