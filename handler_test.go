package vibeGraphql

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGraphqlHandler(t *testing.T) {
	// Register a simple query resolver for testing.
	RegisterQueryResolver("hello", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "world", nil
	})

	// Prepare a test GraphQL query.
	payload := map[string]interface{}{
		"query": "{ hello }",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	GraphqlHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %v", result["data"])
	}
	if data["hello"] != "world" {
		t.Errorf("expected hello: world, got %v", data["hello"])
	}
}
