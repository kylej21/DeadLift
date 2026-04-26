package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// Result holds both the repaired payload and the stub-classified error class.
// When the real on-prem server is wired in, it will return these fields directly.
type Result struct {
	FixedPayload string
	ErrorClass   string
}

// CallMCP sends the raw DLQ payload to the MCP/LLM repair service.
// TODO: replace stub with real HTTP call to the on-prem graphrag server.
func CallMCP(_ context.Context, rawPayload string) (Result, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawPayload), &data); err != nil {
		fixed, _ := json.Marshal(map[string]interface{}{
			"_repaired":   true,
			"original":    rawPayload,
			"repair_note": "non-JSON payload — passed through unchanged",
		})
		return Result{FixedPayload: string(fixed), ErrorClass: "malformed_json"}, nil
	}

	errorClass := classifyPayload(data)

	data["_repaired"] = true
	data["_repair_note"] = "stub repair: field types normalized"
	fixed, err := json.Marshal(data)
	if err != nil {
		return Result{}, fmt.Errorf("re-marshal fixed payload: %w", err)
	}
	return Result{FixedPayload: string(fixed), ErrorClass: errorClass}, nil
}

// classifyPayload derives a best-guess error class from the payload structure.
// The real server will replace this with LLM-based classification.
func classifyPayload(data map[string]interface{}) string {
	for _, v := range data {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if s == "true" || s == "false" {
			return "type_mismatch"
		}
		if _, err := strconv.ParseFloat(s, 64); err == nil {
			return "type_mismatch"
		}
	}
	if len(data) <= 2 {
		return "missing_field"
	}
	return "schema_drift"
}
