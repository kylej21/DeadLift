package main

import (
	"context"
	"encoding/json"
	"fmt"
)

// callMCP sends the raw DLQ payload to the MCP server and returns a fixed payload.
// TODO: replace with real MCP HTTP call once the server endpoint is known.
func callMCP(_ context.Context, rawPayload string) (string, error) {
	// Dummy: parse the payload as JSON and add a "_repaired" marker, then return it.
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawPayload), &data); err != nil {
		// Not JSON — return as-is with a wrapper indicating it was processed.
		fixed, _ := json.Marshal(map[string]interface{}{
			"_repaired":      true,
			"original":       rawPayload,
			"repair_note":    "non-JSON payload — passed through unchanged",
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
