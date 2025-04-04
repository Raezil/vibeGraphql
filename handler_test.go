package vibeGraphql

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildValue_Object(t *testing.T) {
	objValue := &Value{
		Kind: "Object",
		ObjectFields: map[string]*Value{
			"key": {Kind: "String", Literal: "value"},
			"num": {Kind: "Int", Literal: "42"},
		},
	}
	result := buildValue(objValue, nil)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", result)
	}
	if m["key"] != "value" {
		t.Errorf("expected key 'value', got %v", m["key"])
	}
	// buildValue for Int converts to int.
	if m["num"] != 42 {
		t.Errorf("expected num 42, got %v", m["num"])
	}
}

func TestGraphqlHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer([]byte("invalid json")))
	w := httptest.NewRecorder()

	GraphqlHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestGraphqlHandler_NoQuery(t *testing.T) {
	// Send a JSON payload without the "query" field.
	payload := map[string]interface{}{
		"notQuery": "test",
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

	// When query is missing, the parser will likely create an empty document and return an error.
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500 for missing query, got %d", resp.StatusCode)
	}
}

func TestGraphqlUploadHandler_MissingOperations(t *testing.T) {
	// Simulate a multipart request missing the "operations" field.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// Create a dummy file field.
	part, err := writer.CreateFormFile("file1", "dummy.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("dummy content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/graphql/upload", &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	w := httptest.NewRecorder()
	GraphqlUploadHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing operations field, got %d", resp.StatusCode)
	}
}

func TestGraphqlUploadHandler_InvalidMap(t *testing.T) {
	// Simulate a multipart request with invalid JSON in the "map" field.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// Write a valid operations field.
	writer.WriteField("operations", `{"query": "{ hello }", "variables": {}}`)
	// Write an invalid map field.
	writer.WriteField("map", "not a valid json")
	writer.Close()

	req := httptest.NewRequest("POST", "/graphql/upload", &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	w := httptest.NewRecorder()
	GraphqlUploadHandler(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid map JSON, got %d", resp.StatusCode)
	}
}

func TestSetNestedValueAndArrayValue(t *testing.T) {
	var vars = make(map[string]interface{})
	// Test setting a nested value.
	setNestedValue(vars, "user.name", "Alice")
	if vars["user"] == nil {
		t.Fatal("expected nested map for user")
	}
	userMap, ok := vars["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map for user, got %T", vars["user"])
	}
	if userMap["name"] != "Alice" {
		t.Errorf("expected user.name to be 'Alice', got %v", userMap["name"])
	}

	// Test setting an array value.
	setNestedArrayValue(vars, "files.0", "file1.txt")
	arr, ok := vars["files"].([]interface{})
	if !ok {
		t.Fatalf("expected files to be an array, got %T", vars["files"])
	}
	if arr[0] != "file1.txt" {
		t.Errorf("expected files[0] to be 'file1.txt', got %v", arr[0])
	}
}

func TestExecuteDocument_NoDefinitions(t *testing.T) {
	// Create a Document with no definitions.
	doc := &Document{}
	_, err := executeDocument(doc, nil)
	if err == nil {
		t.Error("expected error for document with no definitions, got nil")
	}
}

func TestResolveField_NoResolver(t *testing.T) {
	// Ensure that resolveField returns an error when no resolver is registered.
	field := &Field{Name: "nonexistent"}
	_, err := resolveField(nil, field, nil)
	if err == nil {
		t.Error("expected error for unresolved field, got nil")
	}
}

func dummyHelloResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	return "world", nil
}

func TestGraphqlHandler(t *testing.T) {
	// Register the dummy resolver for the "hello" field.
	QueryResolvers["hello"] = dummyHelloResolver

	// Create a GraphQL query that requests the "hello" field.
	reqBody, err := json.Marshal(map[string]interface{}{
		"query": "query { hello }",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response.
	rr := httptest.NewRecorder()
	GraphqlHandler(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %v", res.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected response to contain 'data' field")
	}
	if data["hello"] != "world" {
		t.Errorf("expected field 'hello' to be 'world', got %v", data["hello"])
	}
}
