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
