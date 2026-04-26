package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Result struct {
	FixedPayload string
	ErrorClass   string
}

type Client struct {
	serverURL string
	apiKey    string
	model     string
	http      *http.Client
}

func New(serverURL, apiKey, model string) *Client {
	return &Client{
		serverURL: serverURL,
		apiKey:    apiKey,
		model:     model,
		http:      &http.Client{Timeout: 120 * time.Second},
	}
}

const systemPrompt = `You are a Pub/Sub dead-letter queue repair agent. A message failed delivery and needs to be diagnosed and fixed.

Analyze the raw message payload using your available tools to understand the expected schema, then return a JSON object with exactly these two fields:
- "error_class": one of "type_mismatch", "schema_drift", "malformed_json", "missing_field", "encoding", "unknown"
- "fixed_payload": the corrected JSON payload as a string

Respond with only the JSON object, no extra text.`

func (c *Client) Call(ctx context.Context, orgID, messageID, rawPayload string) (Result, error) {
	userMsg := fmt.Sprintf(
		"Organization: %s\nMessage ID: %s\n\nRaw DLQ payload:\n%s\n\nDiagnose the error and return the repaired payload.",
		orgID, messageID, rawPayload,
	)

	body, _ := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMsg},
		},
		"temperature":     0,
		"max_tokens":      1024,
		"response_format": map[string]string{"type": "json_object"},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("vllm request: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("vllm HTTP %d: %s", resp.StatusCode, b)
	}

	var cr struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &cr); err != nil || len(cr.Choices) == 0 {
		return Result{}, fmt.Errorf("vllm parse: %w", err)
	}

	var result struct {
		ErrorClass   string `json:"error_class"`
		FixedPayload string `json:"fixed_payload"`
	}
	if err := json.Unmarshal([]byte(cr.Choices[0].Message.Content), &result); err != nil {
		return Result{}, fmt.Errorf("vllm result parse: %w", err)
	}

	return Result{
		ErrorClass:   result.ErrorClass,
		FixedPayload: result.FixedPayload,
	}, nil
}
