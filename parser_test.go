package vibeGraphql

import "testing"

func TestParser(t *testing.T) {
	input := `query { hello }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	if len(doc.Definitions) == 0 {
		t.Fatal("expected at least one definition")
	}

	// Check that the first definition is an operation definition with a valid selection.
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an operation definition")
	}
	if op.SelectionSet == nil || len(op.SelectionSet.Selections) != 1 {
		t.Fatalf("expected one selection in the selection set, got %d", len(op.SelectionSet.Selections))
	}

	field, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		t.Fatal("expected the selection to be a field")
	}
	if field.Name != "hello" {
		t.Errorf("expected field name 'hello', got %q", field.Name)
	}
}
