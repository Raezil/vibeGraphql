package vibeGraphql

import (
	"testing"
)

func TestParser_OperationWithVariables(t *testing.T) {
	input := `
		query TestQuery($x: Int) {
			hello
		}
	`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	if len(doc.Definitions) == 0 {
		t.Fatal("expected at least one definition")
	}

	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an OperationDefinition")
	}

	if op.Name != "TestQuery" {
		t.Errorf("expected operation name 'TestQuery', got %q", op.Name)
	}

	if len(op.VariableDefinitions) != 1 {
		t.Errorf("expected one variable definition, got %d", len(op.VariableDefinitions))
	}

	varDef := op.VariableDefinitions[0]
	if varDef.Variable != "x" {
		t.Errorf("expected variable name 'x', got %q", varDef.Variable)
	}
	if varDef.Type.Name != "Int" {
		t.Errorf("expected variable type 'Int', got %q", varDef.Type.Name)
	}
}

func TestParser_TypeDefinition(t *testing.T) {
	input := `
		type Person {
			name: String,
			age: Int
		}
	`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	def := parser.ParseDocument().Definitions[0]

	typeDef, ok := def.(*TypeDefinition)
	if !ok {
		t.Fatal("expected a TypeDefinition")
	}

	if typeDef.Name != "Person" {
		t.Errorf("expected type name 'Person', got %q", typeDef.Name)
	}

	if len(typeDef.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(typeDef.Fields))
	}

	fieldNames := []string{typeDef.Fields[0].Name, typeDef.Fields[1].Name}
	expected := []string{"name", "age"}
	for i, name := range expected {
		if fieldNames[i] != name {
			t.Errorf("expected field %d to be %q, got %q", i, name, fieldNames[i])
		}
	}
}

func TestParser_ParseObjectValue(t *testing.T) {
	input := `{ key: "value", number: 123 }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()

	if val.Kind != "Object" {
		t.Fatalf("expected Kind 'Object', got %q", val.Kind)
	}

	if v, ok := val.ObjectFields["key"]; !ok || v.Literal != "value" {
		t.Errorf("expected object field 'key' to have value 'value'")
	}

	if v, ok := val.ObjectFields["number"]; !ok || v.Literal != "123" {
		t.Errorf("expected object field 'number' to have value '123', got %q", v.Literal)
	}
}

func TestParser_ParseArrayValue(t *testing.T) {
	input := `[ "one", "two", "three" ]`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	val := parser.parseValue()

	if val.Kind != "Array" {
		t.Fatalf("expected Kind 'Array', got %q", val.Kind)
	}

	if len(val.List) != 3 {
		t.Fatalf("expected array length 3, got %d", len(val.List))
	}

	expected := []string{"one", "two", "three"}
	for i, exp := range expected {
		if val.List[i].Literal != exp {
			t.Errorf("expected element %d to be %q, got %q", i, exp, val.List[i].Literal)
		}
	}
}

func TestParser_ParseArguments(t *testing.T) {
	input := `query { greet(name: "Alice", age: 30) }`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an OperationDefinition")
	}

	if len(op.SelectionSet.Selections) != 1 {
		t.Fatalf("expected one selection, got %d", len(op.SelectionSet.Selections))
	}

	field, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		t.Fatal("expected a Field in selection set")
	}

	if len(field.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(field.Arguments))
	}

	argNames := []string{field.Arguments[0].Name, field.Arguments[1].Name}
	expectedNames := []string{"name", "age"}
	for i, name := range expectedNames {
		if argNames[i] != name {
			t.Errorf("expected argument %d name to be %q, got %q", i, name, argNames[i])
		}
	}
}

func TestParser_NestedSelectionSet(t *testing.T) {
	input := `
		query {
			user {
				name,
				email
			}
		}
	`
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	doc := parser.ParseDocument()

	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatal("expected an OperationDefinition")
	}

	if len(op.SelectionSet.Selections) != 1 {
		t.Fatalf("expected one top-level field, got %d", len(op.SelectionSet.Selections))
	}

	userField, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		t.Fatal("expected a Field for user")
	}

	if userField.Name != "user" {
		t.Errorf("expected top-level field name 'user', got %q", userField.Name)
	}

	if userField.SelectionSet == nil || len(userField.SelectionSet.Selections) != 2 {
		t.Fatalf("expected nested selection set with 2 fields, got %d", len(userField.SelectionSet.Selections))
	}

	nestedFieldNames := []string{}
	for _, sel := range userField.SelectionSet.Selections {
		f, ok := sel.(*Field)
		if ok {
			nestedFieldNames = append(nestedFieldNames, f.Name)
		}
	}
	expectedNested := []string{"name", "email"}
	for i, exp := range expectedNested {
		if nestedFieldNames[i] != exp {
			t.Errorf("expected nested field %d to be %q, got %q", i, exp, nestedFieldNames[i])
		}
	}
}

func TestParseOperationDefinition(t *testing.T) {
	input := `query test { hello }`
	l := NewLexer(input)
	p := NewParser(l)
	doc := p.ParseDocument()

	if len(doc.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(doc.Definitions))
	}

	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatalf("expected OperationDefinition, got %T", doc.Definitions[0])
	}

	if op.Operation != "query" {
		t.Errorf("expected operation 'query', got %s", op.Operation)
	}
	if op.Name != "test" {
		t.Errorf("expected operation name 'test', got %s", op.Name)
	}
	if op.SelectionSet == nil || len(op.SelectionSet.Selections) != 1 {
		t.Fatalf("expected one selection in selection set, got %v", op.SelectionSet)
	}

	field, ok := op.SelectionSet.Selections[0].(*Field)
	if !ok {
		t.Errorf("expected selection to be a Field, got %T", op.SelectionSet.Selections[0])
	}
	if field.Name != "hello" {
		t.Errorf("expected field name 'hello', got %s", field.Name)
	}
}
