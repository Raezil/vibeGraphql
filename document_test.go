package vibeGraphql

import "testing"

// dummyDefinition implements Definition for testing.
type dummyDefinition struct {
	token string
}

func (d *dummyDefinition) TokenLiteral() string {
	return d.token
}

func TestDocumentTokenLiteral(t *testing.T) {
	// Test that a Document returns the token literal of its first definition.
	def := &dummyDefinition{token: "dummy"}
	doc := &Document{Definitions: []Definition{def}}
	if got := doc.TokenLiteral(); got != "dummy" {
		t.Errorf("expected 'dummy', got %q", got)
	}

	// Test when no definitions exist.
	emptyDoc := &Document{}
	if got := emptyDoc.TokenLiteral(); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestOperationDefinitionTokenLiteral(t *testing.T) {
	// When Name is provided, TokenLiteral should return it.
	opWithName := &OperationDefinition{
		Operation: "query",
		Name:      "GetUser",
	}
	if got := opWithName.TokenLiteral(); got != "GetUser" {
		t.Errorf("expected 'GetUser', got %q", got)
	}

	// When Name is empty, TokenLiteral should return Operation.
	opWithoutName := &OperationDefinition{
		Operation: "mutation",
		Name:      "",
	}
	if got := opWithoutName.TokenLiteral(); got != "mutation" {
		t.Errorf("expected 'mutation', got %q", got)
	}
}

func TestVariableDefinitionTokenLiteral(t *testing.T) {
	varDef := &VariableDefinition{
		Variable: "$id",
	}
	if got := varDef.TokenLiteral(); got != "$id" {
		t.Errorf("expected '$id', got %q", got)
	}
}

func TestFieldTokenLiteral(t *testing.T) {
	field := &Field{
		Name: "username",
	}
	if got := field.TokenLiteral(); got != "username" {
		t.Errorf("expected 'username', got %q", got)
	}
}

func TestArgumentTokenLiteral(t *testing.T) {
	arg := &Argument{
		Name: "limit",
	}
	if got := arg.TokenLiteral(); got != "limit" {
		t.Errorf("expected 'limit', got %q", got)
	}
}

func TestTypeDefinitionTokenLiteral(t *testing.T) {
	typeDef := &TypeDefinition{
		Name: "Query",
	}
	if got := typeDef.TokenLiteral(); got != "Query" {
		t.Errorf("expected 'Query', got %q", got)
	}
}

func TestValueTokenLiteral_ObjectAndArray(t *testing.T) {
	// Even if the Value represents an object, TokenLiteral returns the Literal field.
	objVal := &Value{
		Kind:         "Object",
		Literal:      "objectLiteral",
		ObjectFields: map[string]*Value{"key": {Literal: "val"}},
	}
	if got := objVal.TokenLiteral(); got != "objectLiteral" {
		t.Errorf("expected 'objectLiteral', got %q", got)
	}

	// When the Value represents an array, TokenLiteral still returns the Literal field.
	arrVal := &Value{
		Kind:    "Array",
		Literal: "arrayLiteral",
		List: []*Value{
			{Literal: "item1"},
			{Literal: "item2"},
		},
	}
	if got := arrVal.TokenLiteral(); got != "arrayLiteral" {
		t.Errorf("expected 'arrayLiteral', got %q", got)
	}
}

// TestBuildValue_Array verifies that an Array kind Value is correctly built into a slice.
func TestBuildValue_Array(t *testing.T) {
	valArr := &Value{
		Kind: "Array",
		List: []*Value{
			{Kind: "Int", Literal: "1"},
			{Kind: "Int", Literal: "2"},
			{Kind: "Int", Literal: "3"},
		},
	}
	res := buildValue(valArr, nil)
	arr, ok := res.([]interface{})
	if !ok {
		t.Fatal("expected result to be a slice")
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
	expected := []int{1, 2, 3}
	for i, v := range arr {
		num, ok := v.(int)
		if !ok {
			t.Errorf("expected element %d to be int, got %T", i, v)
		}
		if num != expected[i] {
			t.Errorf("expected element %d to be %d, got %d", i, expected[i], num)
		}
	}
}

// TestLexerNumberToken checks that the lexer correctly recognizes a numeric token.
func TestLexerNumberToken(t *testing.T) {
	// Assume that the lexer returns a token of type INT for numeric input.
	input := "12345"
	lexer := NewLexer(input)
	tok := lexer.NextToken()
	// Check the token type. (INT constant must be defined in your lexer implementation.)
	if tok.Type != "INT" {
		t.Errorf("expected token type INT, got %s", tok.Type)
	}
	if tok.Literal != "12345" {
		t.Errorf("expected literal '12345', got %q", tok.Literal)
	}
}
