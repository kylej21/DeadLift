package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// CallMCP sends the raw DLQ payload to the MCP/LLM repair service and returns
// a fixed payload. TODO: replace with real MCP HTTP call.
func CallMCP(_ context.Context, rawPayload string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawPayload), &data); err != nil {
		fixed, _ := json.Marshal(map[string]interface{}{
			"_repaired":   true,
			"original":    rawPayload,
			"repair_note": "non-JSON payload — passed through unchanged",
		})
		return string(fixed), nil
	}
	data["_repaired"] = true
	data["_repair_note"] = "dummy repair: field types normalized"
	fixed, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("re-marshal fixed payload: %w", err)
	}
	return string(fixed), nil
}
