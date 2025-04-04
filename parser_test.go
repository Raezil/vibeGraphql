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

func TestParseTypeDefinition(t *testing.T) {
	input := `type MyType { field1: Int, field2: String }`
	l := NewLexer(input)
	p := NewParser(l)
	doc := p.ParseDocument()

	if len(doc.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(doc.Definitions))
	}

	td, ok := doc.Definitions[0].(*TypeDefinition)
	if !ok {
		t.Fatalf("expected TypeDefinition, got %T", doc.Definitions[0])
	}

	if td.Name != "MyType" {
		t.Errorf("expected type name 'MyType', got %s", td.Name)
	}

	if len(td.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(td.Fields))
	}

	if td.Fields[0].Name != "field1" {
		t.Errorf("expected first field to be 'field1', got %s", td.Fields[0].Name)
	}

	if td.Fields[1].Name != "field2" {
		t.Errorf("expected second field to be 'field2', got %s", td.Fields[1].Name)
	}
}

func TestSkipParenBlock(t *testing.T) {
	// The input starts with a parenthesized block followed by an identifier.
	input := `(arg1: Int, arg2: String) nextToken`
	l := NewLexer(input)
	p := NewParser(l)

	// Before calling skipParenBlock, the current token should be LPAREN.
	if p.curToken.Type != LPAREN {
		t.Fatalf("expected current token to be LPAREN, got %s", p.curToken.Type)
	}

	p.skipParenBlock()

	// After skipping, we expect the next token to be an IDENT with literal "nextToken".
	if p.curToken.Type != IDENT || p.curToken.Literal != "nextToken" {
		t.Fatalf("expected token IDENT 'nextToken' after skipping paren block, got %s (%s)",
			p.curToken.Type, p.curToken.Literal)
	}
}

func TestSkipBlock(t *testing.T) {
	// The input has a block followed by an identifier.
	input := `{ field1: Int } nextToken`
	l := NewLexer(input)
	p := NewParser(l)

	// Before calling skipBlock, current token should be LBRACE.
	if p.curToken.Type != LBRACE {
		t.Fatalf("expected current token to be LBRACE, got %s", p.curToken.Type)
	}

	p.skipBlock()

	// After skipping, the current token should be IDENT with literal "nextToken".
	if p.curToken.Type != IDENT || p.curToken.Literal != "nextToken" {
		t.Fatalf("expected token IDENT 'nextToken' after skipping block, got %s (%s)",
			p.curToken.Type, p.curToken.Literal)
	}
}

func TestParseType(t *testing.T) {
	// Test a simple type: Int!
	input := `Int!`
	l := NewLexer(input)
	p := NewParser(l)
	typ := p.parseType()
	if typ == nil {
		t.Fatalf("expected non-nil type")
	}
	if typ.Name != "Int" {
		t.Errorf("expected type name 'Int', got %s", typ.Name)
	}
	if !typ.NonNull {
		t.Errorf("expected type to be non-null")
	}

	// Test a list type: [String!]!
	input = `[String!]!`
	l = NewLexer(input)
	p = NewParser(l)
	typ = p.parseType()
	if typ == nil {
		t.Fatalf("expected non-nil type")
	}
	if !typ.IsList {
		t.Errorf("expected type to be a list")
	}
	if !typ.NonNull {
		t.Errorf("expected list type to be non-null")
	}
	if typ.Elem == nil {
		t.Errorf("expected list element type, got nil")
	} else {
		if typ.Elem.Name != "String" {
			t.Errorf("expected element type name 'String', got %s", typ.Elem.Name)
		}
		if !typ.Elem.NonNull {
			t.Errorf("expected element type to be non-null")
		}
	}
}

// TestParser_ImplicitQuery verifies that an implicit query (starting with '{')
// is correctly parsed as a query operation with a selection set.
func TestParser_ImplicitQuery(t *testing.T) {
	input := `{ hello world }`
	l := NewLexer(input)
	p := NewParser(l)
	doc := p.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected one definition, got %d", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*OperationDefinition)
	if !ok {
		t.Fatalf("expected OperationDefinition, got %T", doc.Definitions[0])
	}
	if op.Operation != "query" {
		t.Errorf("expected implicit operation to be 'query', got %s", op.Operation)
	}
	// Expect two fields: "hello" and "world"
	if op.SelectionSet == nil {
		t.Fatal("expected a selection set")
	}
	if len(op.SelectionSet.Selections) != 2 {
		t.Errorf("expected 2 selections, got %d", len(op.SelectionSet.Selections))
	}
	fieldNames := []string{}
	for _, sel := range op.SelectionSet.Selections {
		if f, ok := sel.(*Field); ok {
			fieldNames = append(fieldNames, f.Name)
		}
	}
	expected := []string{"hello", "world"}
	for i, exp := range expected {
		if fieldNames[i] != exp {
			t.Errorf("expected field %d to be %q, got %q", i, exp, fieldNames[i])
		}
	}
}

// TestParser_ParseValue_Invalid checks that parseValue produces an Illegal value
// when encountering an unrecognized token.
func TestParser_ParseValue_Invalid(t *testing.T) {
	input := `@invalid`
	l := NewLexer(input)
	p := NewParser(l)
	val := p.parseValue()
	if val.Kind != "Illegal" {
		t.Errorf("expected Kind 'Illegal', got %q", val.Kind)
	}
	if val.Literal != "@" {
		t.Errorf("expected literal '@', got %q", val.Literal)
	}
}

// TestParser_ParseIncompleteObject tests behavior when an object literal is missing its closing brace.
func TestParser_ParseIncompleteObject(t *testing.T) {
	input := `{ key: "value"`
	l := NewLexer(input)
	p := NewParser(l)
	val := p.parseValue()
	if val.Kind != "Object" {
		t.Fatalf("expected Kind 'Object', got %q", val.Kind)
	}
	if _, ok := val.ObjectFields["key"]; !ok {
		t.Errorf("expected object to have key 'key'")
	}
	// Depending on error handling, the parser may simply return what it read.
}

// TestParser_ParseVariableWithoutIdentifier checks that a variable token with no identifier
// after '$' is handled as a Variable with an empty literal.
func TestParser_ParseVariableWithoutIdentifier(t *testing.T) {
	input := `$`
	l := NewLexer(input)
	p := NewParser(l)
	val := p.parseValue()
	if val.Kind != "Variable" {
		t.Errorf("expected Kind 'Variable', got %q", val.Kind)
	}
	if val.Literal != "" {
		t.Errorf("expected empty literal for variable, got %q", val.Literal)
	}
}

// TestParser_NestedArrayInObject tests parsing of an object that contains an array and a nested object.
func TestParser_NestedArrayInObject(t *testing.T) {
	input := `{ items: [1, 2, 3], info: { name: "test" } }`
	l := NewLexer(input)
	p := NewParser(l)
	val := p.parseValue()
	if val.Kind != "Object" {
		t.Fatalf("expected Kind 'Object', got %q", val.Kind)
	}
	items, ok := val.ObjectFields["items"]
	if !ok {
		t.Fatalf("expected key 'items' in object")
	}
	if items.Kind != "Array" {
		t.Errorf("expected 'items' to be Array, got %q", items.Kind)
	}
	if len(items.List) != 3 {
		t.Errorf("expected array length 3, got %d", len(items.List))
	}
	info, ok := val.ObjectFields["info"]
	if !ok {
		t.Fatalf("expected key 'info' in object")
	}
	if info.Kind != "Object" {
		t.Errorf("expected 'info' to be Object, got %q", info.Kind)
	}
	nameField, ok := info.ObjectFields["name"]
	if !ok || nameField.Literal != "test" {
		t.Errorf("expected 'info.name' to be 'test', got %q", nameField.Literal)
	}
}

// TestParser_ComplexTypeDefinition checks that a complex type definition
// with multiple fields, list types, and nested object types is parsed correctly.
func TestParser_ComplexTypeDefinition(t *testing.T) {
	input := `
		type Complex {
			id: ID!,
			names: [String!]!,
			meta: Meta
		}
	`
	l := NewLexer(input)
	p := NewParser(l)
	doc := p.ParseDocument()
	if len(doc.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(doc.Definitions))
	}
	td, ok := doc.Definitions[0].(*TypeDefinition)
	if !ok {
		t.Fatalf("expected TypeDefinition, got %T", doc.Definitions[0])
	}
	if td.Name != "Complex" {
		t.Errorf("expected type name 'Complex', got %s", td.Name)
	}
	if len(td.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(td.Fields))
	}
	if td.Fields[0].Name != "id" {
		t.Errorf("expected first field to be 'id', got %s", td.Fields[0].Name)
	}
	if td.Fields[1].Name != "names" {
		t.Errorf("expected second field to be 'names', got %s", td.Fields[1].Name)
	}
	if td.Fields[2].Name != "meta" {
		t.Errorf("expected third field to be 'meta', got %s", td.Fields[2].Name)
	}
}

// TestParser_ParseBooleanValue verifies that boolean literals are correctly recognized.
func TestParser_ParseBooleanValue(t *testing.T) {
	input := `true false`
	l := NewLexer(input)
	p := NewParser(l)
	// Parse first boolean.
	val1 := p.parseValue()
	if val1.Kind != "Boolean" {
		t.Errorf("expected Kind 'Boolean' for true, got %q", val1.Kind)
	}
	if val1.Literal != "true" {
		t.Errorf("expected literal 'true', got %q", val1.Literal)
	}
	// Parse second boolean.
	val2 := p.parseValue()
	if val2.Kind != "Boolean" {
		t.Errorf("expected Kind 'Boolean' for false, got %q", val2.Kind)
	}
	if val2.Literal != "false" {
		t.Errorf("expected literal 'false', got %q", val2.Literal)
	}
}
