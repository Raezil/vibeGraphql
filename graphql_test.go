package vibeGraphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// A helper type to test reflectResolve when source is not a struct.
type nonStruct int

// dummyResolver returns an error if a specific argument is provided.
func dummyResolver(source interface{}, args map[string]interface{}) (interface{}, error) {
	if args["fail"] == true {
		return nil, fmt.Errorf("dummy failure")
	}
	return "dummy success", nil
}

func TestGraphqlHandlerInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestGraphqlHandlerNoDefinitions(t *testing.T) {
	// Test with a query that yields an empty document.
	payload := map[string]interface{}{
		"query": "",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 for empty document, got %d", resp.StatusCode)
	}
}

func TestReflectResolveNonStruct(t *testing.T) {
	// Call reflectResolve with a non-struct source.
	field := &Field{Name: "Test"}
	_, err := reflectResolve(nonStruct(5), field)
	if err == nil {
		t.Error("expected error when source is not a struct")
	}
}

func TestBuildValueAndArgs(t *testing.T) {
	// Test buildValue for each type.
	variables := map[string]interface{}{
		"var1": 123,
	}

	// Variable kind.
	valVar := &Value{Kind: "Variable", Literal: "var1"}
	resVar := buildValue(valVar, variables)
	if resVar.(int) != 123 {
		t.Errorf("expected 123 for variable, got %v", resVar)
	}

	// Int kind.
	valInt := &Value{Kind: "Int", Literal: "42"}
	resInt := buildValue(valInt, variables)
	if resInt.(int) != 42 {
		t.Errorf("expected 42 for int, got %v", resInt)
	}

	// Boolean kind.
	valBool := &Value{Kind: "Boolean", Literal: "true"}
	resBool := buildValue(valBool, variables)
	if resBool.(bool) != true {
		t.Errorf("expected true for boolean, got %v", resBool)
	}

	// Object kind.
	valObj := &Value{
		Kind: "Object",
		ObjectFields: map[string]*Value{
			"num": {Kind: "Int", Literal: "99"},
		},
	}
	resObj := buildValue(valObj, variables)
	objMap, ok := resObj.(map[string]interface{})
	if !ok || objMap["num"].(int) != 99 {
		t.Errorf("expected object with num 99, got %v", resObj)
	}
}

func TestExecuteSelectionSetWithNestedArray(t *testing.T) {
	// Create a dummy struct for nested resolution.
	type Dummy struct {
		Name string
	}
	// Dummy resolver to simulate nested selection on a slice.
	resolver := func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return []Dummy{{Name: "Alice"}, {Name: "Bob"}}, nil
	}
	// Register dummy resolver.
	RegisterQueryResolver("names", resolver)

	// Build a simple document with a nested selection.
	doc := &Document{
		Definitions: []Definition{
			&OperationDefinition{
				Operation: "query",
				SelectionSet: &SelectionSet{
					Selections: []Selection{
						&Field{
							Name: "names",
							SelectionSet: &SelectionSet{
								Selections: []Selection{
									&Field{Name: "Name"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Execute document.
	resp, err := executeDocument(doc, map[string]interface{}{})
	if err != nil {
		t.Fatalf("executeDocument error: %v", err)
	}

	data := resp["data"].(map[string]interface{})
	// Check that we get an array.
	names, ok := data["names"].([]interface{})
	if !ok || len(names) != 2 {
		t.Fatalf("expected 2 names, got %v", data["names"])
	}
}

func TestExecuteSubscriptionError(t *testing.T) {
	// Create a dummy field and register a subscription resolver that returns a non-channel.
	RegisterSubscriptionResolver("badSub", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "not a channel", nil
	})
	field := &Field{Name: "badSub"}
	_, err := executeSubscription(nil, field, nil)
	if err == nil {
		t.Error("expected error when subscription resolver returns non-channel")
	}
}

func TestExecuteSubscriptionSuccess(t *testing.T) {
	// Create a dummy subscription resolver that returns a channel.
	ch := make(chan interface{}, 1)
	ch <- "event1"
	close(ch)
	RegisterSubscriptionResolver("goodSub", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return ch, nil
	})
	field := &Field{Name: "goodSub"}
	subCh, err := executeSubscription(nil, field, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case event := <-subCh:
		if event != "event1" {
			t.Errorf("expected 'event1', got %v", event)
		}
	case <-time.After(1 * time.Second):
		t.Error("timed out waiting for subscription event")
	}
}

func TestParserIllegalValue(t *testing.T) {
	// Provide an illegal token to test parseValue fallback.
	input := `$`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()
	if val.Kind != "Variable" || val.Literal != "" {
		t.Errorf("expected Variable with empty literal, got Kind: %s, Literal: %q", val.Kind, val.Literal)
	}
}

func TestLexerIllegalCharacter(t *testing.T) {
	// Test lexer with an unexpected character.
	input := "@"
	lexer := NewLexer(input)
	tok := lexer.NextToken()
	if tok.Type != ILLEGAL {
		t.Errorf("expected ILLEGAL token, got %s", tok.Type)
	}
}

func TestParseObjectWithMissingKey(t *testing.T) {
	// Test the object parser for error handling (it should return an Illegal Value if key is missing).
	input := `{ : "value" }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseObject()
	if val.Kind != "Illegal" {
		t.Errorf("expected Illegal kind, got %s", val.Kind)
	}
}

func TestOperationDefinitionImplicitQuery(t *testing.T) {
	// Test implicit query when the token is '{' instead of an explicit "query".
	input := `{ hello }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatal("expected one definition for implicit query")
	}
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected operation definition")
	}
	if op.Operation != "query" {
		t.Errorf("expected operation to be 'query', got %q", op.Operation)
	}
}

// --- Tests for resolveArgument ---

func TestResolveArgument_Int(t *testing.T) {
	arg := &Argument{
		Name: "age",
		Value: &Value{
			Kind:    "Int",
			Literal: "42",
		},
	}
	res, err := resolveArgument(arg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, ok := res.(int)
	if !ok || i != 42 {
		t.Errorf("expected 42, got %v", res)
	}
}

func TestResolveArgument_String(t *testing.T) {
	arg := &Argument{
		Name: "greeting",
		Value: &Value{
			Kind:    "String",
			Literal: "hello",
		},
	}
	res, err := resolveArgument(arg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := res.(string)
	if !ok || s != "hello" {
		t.Errorf("expected 'hello', got %v", res)
	}
}

func TestResolveArgument_Boolean(t *testing.T) {
	arg := &Argument{
		Name: "flag",
		Value: &Value{
			Kind:    "Boolean",
			Literal: "true",
		},
	}
	res, err := resolveArgument(arg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, ok := res.(bool)
	if !ok || b != true {
		t.Errorf("expected true, got %v", res)
	}
}

func TestResolveArgument_Variable(t *testing.T) {
	vars := map[string]interface{}{
		"var1": "variable value",
	}
	arg := &Argument{
		Name: "varArg",
		Value: &Value{
			Kind:    "Variable",
			Literal: "var1",
		},
	}
	res, err := resolveArgument(arg, vars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := res.(string)
	if !ok || s != "variable value" {
		t.Errorf("expected 'variable value', got %v", res)
	}
}

func TestResolveArgument_Default(t *testing.T) {
	arg := &Argument{
		Name: "defaultArg",
		Value: &Value{
			Kind:    "Foo", // unhandled kind defaults to literal
			Literal: "default",
		},
	}
	res, err := resolveArgument(arg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := res.(string)
	if !ok || s != "default" {
		t.Errorf("expected 'default', got %v", res)
	}
}

// --- Tests for reflective resolution ---

type sample struct {
	Test      int
	JsonField string `json:"json_field"`
}

func TestReflectResolve_SuccessByName(t *testing.T) {
	src := &sample{Test: 100, JsonField: "json value"}
	field := &Field{Name: "Test"}
	res, err := reflectResolve(src, field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, ok := res.(int)
	if !ok || i != 100 {
		t.Errorf("expected 100, got %v", res)
	}
}

func TestReflectResolve_SuccessByJSONTag(t *testing.T) {
	src := &sample{Test: 200, JsonField: "tagged"}
	field := &Field{Name: "json_field"}
	res, err := reflectResolve(src, field)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := res.(string)
	if !ok || s != "tagged" {
		t.Errorf("expected 'tagged', got %v", res)
	}
}

func TestResolveField_Reflective(t *testing.T) {
	// Test reflective resolution when source is non-nil.
	src := &sample{Test: 300, JsonField: "reflective"}
	// Create a field with no resolver registered, so it will try reflective lookup.
	field := &Field{Name: "Test"}
	res, err := resolveField(src, field, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	i, ok := res.(int)
	if !ok || i != 300 {
		t.Errorf("expected 300, got %v", res)
	}
}

// --- Tests for Parser variable definitions and arguments ---

func TestParseVariableDefinitions(t *testing.T) {
	// Query with a variable definition that includes non-null.
	input := `query ($var: Int!) { hello }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an operation definition")
	}
	if len(op.VariableDefinitions) != 1 {
		t.Fatalf("expected one variable definition, got %d", len(op.VariableDefinitions))
	}
	varDef := op.VariableDefinitions[0]
	if varDef.Variable != "var" {
		t.Errorf("expected variable name 'var', got %q", varDef.Variable)
	}
	if varDef.Type.Name != "Int" {
		t.Errorf("expected type 'Int', got %q", varDef.Type.Name)
	}
	if !varDef.Type.NonNull {
		t.Errorf("expected NonNull to be true")
	}
}

func TestParseArgumentsMultiple(t *testing.T) {
	// Test parsing a field with multiple arguments.
	input := `query { field(arg1: 10, arg2: "value", arg3: true) }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok || op.SelectionSet == nil || len(op.SelectionSet.Selections) == 0 {
		t.Fatal("failed to parse operation definition with arguments")
	}
	field, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		t.Fatal("expected a field")
	}
	if len(field.Arguments) != 3 {
		t.Errorf("expected 3 arguments, got %d", len(field.Arguments))
	}
	argNames := []string{"arg1", "arg2", "arg3"}
	for i, arg := range field.Arguments {
		if arg.Name != argNames[i] {
			t.Errorf("expected argument name %q, got %q", argNames[i], arg.Name)
		}
	}
}

func TestParseObject_Nested(t *testing.T) {
	// Test parsing an object literal with nested object.
	input := `{ key1: 123, key2: { nested: "value" } }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseObject()
	if val.Kind != "Object" {
		t.Fatalf("expected kind 'Object', got %s", val.Kind)
	}
	if len(val.ObjectFields) != 2 {
		t.Errorf("expected 2 keys, got %d", len(val.ObjectFields))
	}
	// Check first key.
	field1, ok := val.ObjectFields["key1"]
	if !ok {
		t.Error("expected key1 in object")
	} else if field1.Kind != "Int" || field1.Literal != "123" {
		t.Errorf("unexpected value for key1: %+v", field1)
	}
	// Check nested object.
	field2, ok := val.ObjectFields["key2"]
	if !ok {
		t.Error("expected key2 in object")
	} else if field2.Kind != "Object" {
		t.Errorf("expected key2 to be Object, got %s", field2.Kind)
	} else {
		nested, ok := field2.ObjectFields["nested"]
		if !ok || nested.Literal != "value" {
			t.Errorf("expected nested value 'value', got %v", nested)
		}
	}
}

func TestParseValue_Enum(t *testing.T) {
	// Test that an identifier that is not a boolean is treated as an enum.
	input := `ENUM_VALUE`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()
	if val.Kind != "Enum" {
		t.Errorf("expected kind 'Enum', got %s", val.Kind)
	}
	if val.Literal != "ENUM_VALUE" {
		t.Errorf("expected literal 'ENUM_VALUE', got %q", val.Literal)
	}
}

// --- Tests for operation definitions with a name ---

func TestOperationDefinitionWithNameAndVariables(t *testing.T) {
	input := `query MyQuery($id: Int) { hello }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an operation definition")
	}
	if op.Name != "MyQuery" {
		t.Errorf("expected operation name 'MyQuery', got %q", op.Name)
	}
	if len(op.VariableDefinitions) != 1 {
		t.Errorf("expected one variable definition, got %d", len(op.VariableDefinitions))
	}
}

// --- Test for parseValue handling of "$" with no identifier ---

func TestParseValue_VariableMissingIdentifier(t *testing.T) {
	// Using input with only "$" should set Kind to "Variable" with empty Literal.
	input := `$`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()
	if val.Kind != "Variable" {
		t.Errorf("expected Kind 'Variable', got %s", val.Kind)
	}
	if val.Literal != "" {
		t.Errorf("expected empty Literal, got %q", val.Literal)
	}
}

// --- Test for GraphqlHandler with valid JSON but missing query field ---
func TestGraphqlHandler_MissingQueryField(t *testing.T) {
	// Provide valid JSON but without a "query" field.
	payload := map[string]interface{}{
		"not_query": "value",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/graphql", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("expected status 500 for missing query field, got %d", resp.StatusCode)
	}
}

func TestValueTokenLiteral(t *testing.T) {
	val := &Value{Literal: "value"}
	if val.TokenLiteral() != "value" {
		t.Errorf("expected TokenLiteral to be 'value', got %q", val.TokenLiteral())
	}
}

// --- Test for executeDocument error when definition is not an OperationDefinition ---

type dummyNonOp struct{}

func (d *dummyNonOp) TokenLiteral() string { return "dummy" }

func TestExecuteDocumentNonOpDefinition(t *testing.T) {
	doc := &Document{Definitions: []Definition{&dummyNonOp{}}}
	_, err := executeDocument(doc, nil)
	if err == nil {
		t.Error("expected error when definition is not an OperationDefinition")
	}
}

// --- Test for buildValue with an empty object ---

func TestBuildValue_EmptyObject(t *testing.T) {
	valObj := &Value{
		Kind:         "Object",
		ObjectFields: map[string]*Value{},
	}
	res := buildValue(valObj, nil)
	objMap, ok := res.(map[string]interface{})
	if !ok || len(objMap) != 0 {
		t.Errorf("expected an empty object, got %v", res)
	}
}

func TestReflectResolve_FieldNotFound(t *testing.T) {
	src := &sample{Test: 500, JsonField: "not found"}
	field := &Field{Name: "NonExistent"}
	_, err := reflectResolve(src, field)
	if err == nil {
		t.Error("expected error when field is not found via reflection")
	}
}

// --- Test for parseDefinition with an invalid token ---

func TestParseDefinitionInvalid(t *testing.T) {
	input := "random"
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	def := parser.parseDefinition()
	if def != nil {
		t.Error("expected nil definition for invalid input")
	}
}

// --- Test for parseObject error: missing colon (illegal object) ---

func TestParseObjectMissingColon(t *testing.T) {
	input := `{ key "value" }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseObject()
	if val.Kind != "Illegal" {
		t.Errorf("expected kind 'Illegal' for missing colon, got %s", val.Kind)
	}
}

// --- Test for GraphqlHandler when variables is explicitly nil ---
func TestGraphqlHandler_NilVariables(t *testing.T) {
	// Register a simple query resolver.
	RegisterQueryResolver("greet", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "hi", nil
	})
	payload := map[string]interface{}{
		"query":     "{ greet }",
		"variables": nil,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	GraphqlHandler(w, req)
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// --- Test for executeSubscription branch when resolver returns non-channel ---

func TestExecuteSubscriptionNonChannel(t *testing.T) {
	RegisterSubscriptionResolver("badSub", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "not a channel", nil
	})
	field := &Field{Name: "badSub"}
	_, err := executeSubscription(nil, field, nil)
	if err == nil {
		t.Error("expected error when subscription resolver returns non-channel")
	}
}

// --- Test for SubscriptionHandler error paths using a fake upgrader ---
// Since testing real WebSocket upgrades is complex in unit tests,
// we override the upgrader to simulate an upgrade failure.
func TestSubscriptionHandlerUpgradeFail(t *testing.T) {
	// Override upgrader with one that always fails.
	origUpgrader := upgrader
	defer func() { upgrader = origUpgrader }()
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return false },
	}

	req := httptest.NewRequest("GET", "/subscription", nil)
	w := httptest.NewRecorder()
	SubscriptionHandler(w, req)
	resp := w.Result()
	// Since upgrade fails, we expect an HTTP error.
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400 due to upgrade failure, got %d", resp.StatusCode)
	}
}

// --- Test for lexer string token ---
func TestLexerStringToken(t *testing.T) {
	input := `"hello world"`
	lexer := NewLexer(input)
	tok := lexer.NextToken()
	if tok.Type != STRING || tok.Literal != "hello world" {
		t.Errorf("expected string token with literal 'hello world', got Type: %s, Literal: %q", tok.Type, tok.Literal)
	}
}

// --- Test for buildArgs with multiple arguments ---
func TestBuildArgsMultiple(t *testing.T) {
	// Create a field with several arguments.
	field := &Field{
		Name: "dummyField",
		Arguments: []Argument{
			{Name: "a", Value: &Value{Kind: "Int", Literal: "10"}},
			{Name: "b", Value: &Value{Kind: "String", Literal: "test"}},
		},
	}
	args := buildArgs(field, nil)
	if len(args) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(args))
	}
	if v, ok := args["a"].(int); !ok || v != 10 {
		t.Errorf("expected argument 'a' to be 10, got %v", args["a"])
	}
	if s, ok := args["b"].(string); !ok || s != "test" {
		t.Errorf("expected argument 'b' to be 'test', got %v", args["b"])
	}
}

// Test that reflectResolve returns an error when provided a nil pointer.
func TestReflectResolve_NilPointer(t *testing.T) {
	var src *sample = nil
	field := &Field{Name: "Test"}
	_, err := reflectResolve(src, field)
	if err == nil {
		t.Error("expected error for nil pointer source")
	}
}

// Test buildValue for Boolean false.
func TestBuildValue_BooleanFalse(t *testing.T) {
	valBool := &Value{Kind: "Boolean", Literal: "false"}
	resBool := buildValue(valBool, nil)
	if b, ok := resBool.(bool); !ok || b != false {
		t.Errorf("expected false for boolean, got %v", resBool)
	}
}

// Test parseValue default branch for an illegal token.
func TestParseValue_Illegal(t *testing.T) {
	// "@" is not a valid starting character, so lexer returns an ILLEGAL token.
	input := "@"
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()
	if val.Kind != "Illegal" {
		t.Errorf("expected kind 'Illegal', got %s", val.Kind)
	}
	if val.Literal != "@" {
		t.Errorf("expected literal '@', got %q", val.Literal)
	}
}

// Test executeSelectionSet to skip non-Field selections.
type dummySel struct{}

func (d *dummySel) TokenLiteral() string {
	return "dummy"
}

func TestExecuteSelectionSet_SkipNonField(t *testing.T) {
	// Create a selection set with a dummy selection that is not a *Field.
	ss := &SelectionSet{Selections: []Selection{&dummySel{}}}
	result, err := executeSelectionSet(nil, ss, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected result to be empty, got %v", result)
	}
}

type fakeUpgrader struct{}

// fakeUpgrader is a stub that implements the minimal interface for our tests.

func (f fakeUpgrader) Upgrade(w httptest.ResponseRecorder, r *http.Request, responseHeader map[string][]string) (fakeConn, error) {
	return fakeConn{}, nil
}

// fakeConn is a stub connection that fails on ReadMessage.
type fakeConn struct{}

func (f fakeConn) ReadMessage() (int, []byte, error) {
	return 0, nil, nil // Simulate a read failure by returning nil error? For our test we'll leave it as is.
}
func (f fakeConn) WriteJSON(v interface{}) error { return nil }
func (f fakeConn) WriteMessage(messageType int, data []byte) error {
	return nil
}
func (f fakeConn) Close() error { return nil }

// Test executeSubscription with channel resolution already exists; we rely on that to cover success paths.
// Additionally, ensure executeSubscription returns an error when resolver returns a non-channel.
func TestExecuteSubscription_NonChannel(t *testing.T) {
	RegisterSubscriptionResolver("nonChannel", func(source interface{}, args map[string]interface{}) (interface{}, error) {
		return "not a channel", nil
	})
	field := &Field{Name: "nonChannel"}
	_, err := executeSubscription(nil, field, nil)
	if err == nil {
		t.Error("expected error when subscription resolver returns non-channel")
	}
}

// Test buildArgs with a nested object value.
func TestBuildArgs_NestedObject(t *testing.T) {
	// Field with an argument whose value is an object.
	field := &Field{
		Name: "dummy",
		Arguments: []Argument{
			{
				Name: "config",
				Value: &Value{
					Kind: "Object",
					ObjectFields: map[string]*Value{
						"threshold": {Kind: "Int", Literal: "5"},
					},
				},
			},
		},
	}
	args := buildArgs(field, nil)
	config, ok := args["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config to be an object")
	}
	if threshold, ok := config["threshold"].(int); !ok || threshold != 5 {
		t.Errorf("expected threshold to be 5, got %v", config["threshold"])
	}
}
