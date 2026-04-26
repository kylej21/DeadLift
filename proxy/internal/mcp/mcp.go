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
	FixedPayload    string
	ErrorClass      string
	ConfidenceScore int
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

const rcaSystemPrompt = `You are a root cause analysis expert for Pub/Sub pipeline failures.

Given a failed message and its repaired version, produce a detailed root cause analysis in plain text covering:
1. What went wrong and why
2. Which upstream system or code path likely caused this
3. How the fix addresses the root cause
4. Recommendations to prevent recurrence

Be specific and technical. Reference field names and values from the payload.`

func (c *Client) CallRCA(ctx context.Context, orgID, messageID, rawPayload, fixedPayload, errorClass string) (string, error) {
	userMsg := fmt.Sprintf(
		"Organization: %s\nMessage ID: %s\nError class: %s\n\nOriginal failed payload:\n%s\n\nRepaired payload:\n%s\n\nProvide a thorough root cause analysis.",
		orgID, messageID, errorClass, rawPayload, fixedPayload,
	)

	body, _ := json.Marshal(map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": rcaSystemPrompt},
			{"role": "user", "content": userMsg},
		},
		"temperature": 0.2,
		"max_tokens":  2048,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("vllm rca request: %w", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vllm HTTP %d: %s", resp.StatusCode, b)
	}

	var cr struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &cr); err != nil || len(cr.Choices) == 0 {
		return "", fmt.Errorf("vllm rca parse: %w", err)
	}
	return cr.Choices[0].Message.Content, nil
}

const systemPrompt = `You are a Pub/Sub dead-letter queue repair agent. A message failed delivery and needs to be diagnosed and fixed.

You have access to the following tools to help diagnose and repair the message:
- fetch_gcp_logs: Fetch GCP Cloud Run / GCE logs filtered by resource type and severity. Use this to find errors or patterns related to the failed message.
- gcp_list_log_resource_types: List available GCP monitored resource types to use as filters when calling fetch_gcp_logs.
- graph_rag_query: Query a knowledge graph built from codebase and incident data. Use method='local' for specific entity/error questions (handlers, configs, schemas). Use method='global' for broad system-wide questions.
- bigquery_last_n_query: Fetch recent rows from a BigQuery table to inspect expected schemas or historical data patterns.

Analyze the raw message payload using these tools to understand the expected schema, then return a JSON object with exactly these three fields:
- "error_class": a short string describing the type of issue found (e.g. "type_mismatch", "schema_drift", "malformed_json", "missing_field", "encoding", or any other appropriate label)
- "fixed_payload": the corrected JSON payload as a string
- "confidence_score": an integer from 0 to 100 representing how confident you are in the diagnosis and repair

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
		ErrorClass      string `json:"error_class"`
		FixedPayload    string `json:"fixed_payload"`
		ConfidenceScore int    `json:"confidence_score"`
	}
	if err := json.Unmarshal([]byte(cr.Choices[0].Message.Content), &result); err != nil {
		return Result{}, fmt.Errorf("vllm result parse: %w", err)
	}

	return Result{
		ErrorClass:      result.ErrorClass,
		FixedPayload:    result.FixedPayload,
		ConfidenceScore: result.ConfidenceScore,
	}, nil
}
