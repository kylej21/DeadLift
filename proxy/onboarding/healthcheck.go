package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// healthCheck verifies the subscription and topic exist using the customer's token.
// IAM propagation takes ~60s so we don't test repairSA access here — just that
// the resources the user provided are real.
func healthCheck(ctx context.Context, token, projectID, dlqSub, mainTopic string) error {
	subURL := fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/subscriptions/%s", projectID, dlqSub)
	if err := getResource(ctx, token, subURL); err != nil {
		return fmt.Errorf("DLQ subscription %q not found: %w", dlqSub, err)
	}

	topicURL := fmt.Sprintf("https://pubsub.googleapis.com/v1/projects/%s/topics/%s", projectID, mainTopic)
	if err := getResource(ctx, token, topicURL); err != nil {
		return fmt.Errorf("main topic %q not found: %w", mainTopic, err)
	}

	return nil
}

func getResource(ctx context.Context, token, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	return nil
}
