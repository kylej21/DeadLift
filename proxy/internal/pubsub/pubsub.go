package pubsub

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"proxy/internal/models"
)

var (
	repairTokenMu     sync.Mutex
	repairTokenValue  string
	repairTokenExpiry time.Time
)

func GetMetadataToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("metadata server unreachable (not running on GCP?): %w", err)
	}
	defer resp.Body.Close()
	var tr struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	return tr.AccessToken, nil
}

func GetRepairSAToken(ctx context.Context, repairSA string) (string, error) {
	repairTokenMu.Lock()
	defer repairTokenMu.Unlock()

	// Reuse the cached token until 5 minutes before it expires.
	if repairTokenValue != "" && time.Now().Before(repairTokenExpiry.Add(-5*time.Minute)) {
		return repairTokenValue, nil
	}

	metaToken, err := GetMetadataToken(ctx)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", repairSA)
	body := `{"scope":["https://www.googleapis.com/auth/pubsub"]}`
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+metaToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("generateAccessToken HTTP %d: %s", resp.StatusCode, b)
	}
	var tr struct {
		AccessToken string `json:"accessToken"`
		ExpireTime  string `json:"expireTime"`
	}
	if err := json.Unmarshal(b, &tr); err != nil {
		return "", err
	}
	repairTokenValue = tr.AccessToken
	if expiry, err := time.Parse(time.RFC3339, tr.ExpireTime); err == nil {
		repairTokenExpiry = expiry
	} else {
		repairTokenExpiry = time.Now().Add(55 * time.Minute)
	}
	return repairTokenValue, nil
}

func PullMessages(ctx context.Context, token, subscription string) ([]models.PubSubMessage, error) {
	url := fmt.Sprintf("https://pubsub.googleapis.com/v1/%s:pull", subscription)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(`{"maxMessages":10}`))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pull HTTP %d: %s", resp.StatusCode, b)
	}
	var result struct {
		ReceivedMessages []models.PubSubMessage `json:"receivedMessages"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result.ReceivedMessages, nil
}

func AckMessages(ctx context.Context, token, subscription string, ackIDs []string) error {
	url := fmt.Sprintf("https://pubsub.googleapis.com/v1/%s:acknowledge", subscription)
	body, _ := json.Marshal(map[string][]string{"ackIds": ackIDs})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ack HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}

func PublishMessage(ctx context.Context, token, topic, data string, attributes map[string]string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	body, _ := json.Marshal(map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{
				"data":       encoded,
				"attributes": attributes,
			},
		},
	})
	url := fmt.Sprintf("https://pubsub.googleapis.com/v1/%s:publish", topic)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}
