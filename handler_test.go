package vibeGraphql

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestResolveArgument_Variable_Found(t *testing.T) {
	vars := map[string]interface{}{
		"var1": "variableValue",
	}
	arg := &Argument{Value: &Value{Kind: "Variable", Literal: "var1"}}
	result, err := resolveArgument(arg, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "variableValue" {
		t.Errorf("expected 'variableValue', got %v", result)
	}
}

func TestResolveArgument_Variable_NotFound(t *testing.T) {
	arg := &Argument{Value: &Value{Kind: "Variable", Literal: "missingVar"}}
	_, err := resolveArgument(arg, nil)
	if err == nil {
		t.Error("expected an error for missing variable, got nil")
	}
}

// ---------- Additional test for reflectResolve ----------

// DummyStruct is used to test reflective field resolution.
type DummyStruct struct {
	FieldA string `json:"fieldA"`
	FieldB int
}

func TestReflectResolve(t *testing.T) {
	dummy := DummyStruct{FieldA: "valueA", FieldB: 99}
	// Test by using the JSON tag.
	field := &Field{Name: "fieldA"}
	res, err := reflectResolve(dummy, field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "valueA" {
		t.Errorf("expected 'valueA', got %v", res)
	}

	// Test by matching struct field name directly (case-insensitive).
	field = &Field{Name: "FIELDB"}
	res, err = reflectResolve(dummy, field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 99 {
		t.Errorf("expected 99, got %v", res)
	}

	// Test with a non-existent field.
	field = &Field{Name: "nonexistent"}
	_, err = reflectResolve(dummy, field)
	if err == nil {
		t.Error("expected error for nonexistent field, got nil")
	}
}

// ---------- Additional test for executeSelectionSet with nested selections ----------

// DummyUser is used to simulate nested field resolution.
type DummyUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestExecuteSelectionSet_Nested(t *testing.T) {
	// Create a dummy user.
	user := DummyUser{Name: "Alice", Age: 30}

	// Manually create a selection set requesting "name" and "age".
	selectionSet := &SelectionSet{
		Selections: []Selection{
			&Field{Name: "name"},
			&Field{Name: "age"},
		},
	}

	// Call executeSelectionSet with the dummy user as source.
	result, err := executeSelectionSet(user, selectionSet, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that the resulting map contains the expected values.
	if result["name"] != "Alice" {
		t.Errorf("expected name 'Alice', got %v", result["name"])
	}
	if result["age"] != 30 {
		t.Errorf("expected age 30, got %v", result["age"])
	}
}

// ---------- Additional test for buildArgs ----------

func TestBuildArgs(t *testing.T) {
	// Create a field with arguments.
	field := &Field{
		Name: "dummy",
		Arguments: []Argument{
			{Name: "arg1", Value: &Value{Kind: "String", Literal: "hello"}},
			{Name: "arg2", Value: &Value{Kind: "Int", Literal: "42"}},
		},
	}

	args := buildArgs(field, nil)
	if args["arg1"] != "hello" {
		t.Errorf("expected arg1 to be 'hello', got %v", args["arg1"])
	}
	if args["arg2"] != 42 {
		t.Errorf("expected arg2 to be 42, got %v", args["arg2"])
	}
}

// ---------- Additional test for executeDocument error handling ----------

func TestExecuteDocument_UnsupportedDefinition(t *testing.T) {
	// Create a Document with an unsupported definition type.
	doc := &Document{
		Definitions: []Definition{
			// Using a dummy implementation of Definition.
			struct{ Node }{},
		},
	}
	_, err := executeDocument(doc, nil)
	if err == nil {
		t.Error("expected error for unsupported definition type, got nil")
	}
}

// ---------- Additional multipart upload tests can be added if needed ----------

// For example, you might simulate a successful file upload with valid operations and map fields.
// (This would likely require a temporary file or buffer with the correct multipart data.)
// Here’s an example placeholder test that you can expand upon:

func TestGraphqlUploadHandler_Success(t *testing.T) {
	// Register a dummy resolver for "hello"
	QueryResolvers["hello"] = func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "world", nil
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// Write valid operations field.
	writer.WriteField("operations", `{"query": "{ hello }", "variables": {}}`)
	// Write a valid map field.
	writer.WriteField("map", `{"file1": ["variables.file"]}`)
	// Create a dummy file.
	part, err := writer.CreateFormFile("file1", "dummy.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("dummy content"))
	writer.Close()

	// Use the proper upload endpoint.
	req := httptest.NewRequest("POST", "/graphql/upload", &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	rr := httptest.NewRecorder()
	GraphqlUploadHandler(rr, req)
	res := rr.Result()
	defer res.Body.Close()

	// Expect 200 OK since the query resolves correctly.
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
}

// TestSetNestedValue verifies that setNestedValue properly updates a nested map.
func TestSetNestedValue(t *testing.T) {
	vars := make(map[string]interface{})
	setNestedValue(vars, "a.b", "nestedValue")
	m, ok := vars["a"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected key 'a' to be a map, got %T", vars["a"])
	}
	if m["b"] != "nestedValue" {
		t.Errorf("expected nested value 'nestedValue', got %v", m["b"])
	}
}

// TestSetNestedArrayValue verifies that setNestedArrayValue correctly creates or extends an array.
func TestSetNestedArrayValue(t *testing.T) {
	vars := make(map[string]interface{})
	// Set an element at index 0.
	setNestedArrayValue(vars, "files.0", "file0")
	arr, ok := vars["files"].([]interface{})
	if !ok {
		t.Fatalf("expected key 'files' to be an array, got %T", vars["files"])
	}
	if len(arr) != 1 || arr[0] != "file0" {
		t.Errorf("expected array with ['file0'], got %v", arr)
	}

	// Extend the array to index 2.
	setNestedArrayValue(vars, "files.2", "file2")
	arr, ok = vars["files"].([]interface{})
	if !ok || len(arr) < 3 {
		t.Fatalf("expected array of length at least 3, got %v", vars["files"])
	}
	if arr[2] != "file2" {
		t.Errorf("expected element at index 2 to be 'file2', got %v", arr[2])
	}
}

// TestGraphqlUploadHandler_MissingOperations verifies that GraphqlUploadHandler returns an error
// when the "operations" field is missing in a multipart upload.
func TestGraphqlUploadHandler_MissingOperations(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	// Omit operations field.
	writer.WriteField("map", `{"file1": ["variables.file"]}`)
	part, err := writer.CreateFormFile("file1", "dummy.txt")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("dummy content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/graphql/upload", &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	rr := httptest.NewRecorder()
	GraphqlUploadHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing operations, got %d", rr.Code)
	}
}

// TestGraphqlUploadHandler_MissingMap verifies that GraphqlUploadHandler returns an error
// when the "map" field is missing in a multipart upload.
func TestGraphqlUploadHandler_MissingMap(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("operations", `{"query": "{ hello }", "variables": {}}`)
	// Omit map field.
	writer.Close()

	req := httptest.NewRequest("POST", "/graphql/upload", &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
	rr := httptest.NewRecorder()
	GraphqlUploadHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing map field, got %d", rr.Code)
	}
}

// TestSubscriptionHandler_UpgradeFailure simulates a failure in upgrading the HTTP connection to a WebSocket.
func TestSubscriptionHandler_UpgradeFailure(t *testing.T) {
	// Override the upgrader to always fail by rejecting the origin.
	origUpgrader := upgrader
	defer func() { upgrader = origUpgrader }()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return false },
	}

	req := httptest.NewRequest("GET", "/subscription", nil)
	rr := httptest.NewRecorder()
	SubscriptionHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for upgrade failure, got %d", rr.Code)
	}
}
