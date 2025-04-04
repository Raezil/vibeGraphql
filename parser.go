package vibeGraphql

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	// initialize two tokens
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseDocument() *Document {
	doc := &Document{}
	for p.curToken.Type != EOF {
		def := p.parseDefinition()
		if def != nil {
			doc.Definitions = append(doc.Definitions, def)
		}
		// Always advance at least one token to ensure progress.
		if p.curToken.Type != EOF {
			p.nextToken()
		}
	}
	return doc
}

// In parser.go, add:

func (p *Parser) parseTypeDefinition() Definition {
	// Assume the current token is "type"
	p.nextToken() // Skip "type"
	if p.curToken.Type != IDENT {
		// Error handling could go here.
		return nil
	}
	typeName := p.curToken.Literal
	p.nextToken() // Move past type name

	// Expect an opening brace
	if p.curToken.Type != LBRACE {
		return nil // or handle error
	}
	p.nextToken() // Skip '{'

	var fields []*Field
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		field := p.parseField()
		if field != nil {
			fields = append(fields, field)
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // Skip '}'

	return &TypeDefinition{
		Name:   typeName,
		Fields: fields,
	}
}

func (p *Parser) parseDefinition() Definition {
	// Handle operation definitions.
	if p.curToken.Literal == "query" ||
		p.curToken.Literal == "mutation" ||
		p.curToken.Literal == "subscription" {
		return p.parseOperationDefinition()
	}
	// Handle implicit queries (starting with '{').
	if p.curToken.Type == LBRACE {
		return p.parseOperationDefinition()
	}
	// When a "type" keyword is encountered, use skipTypeDefinition to parse it.
	if p.curToken.Literal == "type" {
		return p.skipTypeDefinition()
	}
	// If the token isn't recognized, advance and return nil.
	p.nextToken()
	return nil
}

// skipTypeDefinition advances the parser past a type definition block.
func (p *Parser) skipTypeDefinition() Definition {
	// Skip the "type" keyword.
	p.nextToken()
	if p.curToken.Type != IDENT {
		// Expected type name.
		return nil
	}
	typeName := p.curToken.Literal
	p.nextToken() // move past type name

	// Expect an opening brace.
	if p.curToken.Type != LBRACE {
		return nil
	}
	p.nextToken() // Skip '{'

	var fields []*Field
	iterations := 0
	maxIterations := 10000 // safeguard to prevent infinite loops

	// Parse fields until we hit the closing brace.
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		iterations++
		if iterations > maxIterations {
			break
		}
		field := p.parseTypeField()
		if field != nil {
			fields = append(fields, field)
		} else {
			// If no field is returned, still advance the token to avoid a freeze.
			p.nextToken()
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	// Skip the closing brace.
	if p.curToken.Type == RBRACE {
		p.nextToken()
	}
	return &TypeDefinition{
		Name:   typeName,
		Fields: fields,
	}
}

// skipBlock skips over a block delimited by '{' and '}'.
func (p *Parser) skipBlock() {
	// Assume the current token is LBRACE.
	depth := 0
	iterations := 0
	maxIterations := 10000 // safeguard to prevent infinite loop
	for p.curToken.Type != EOF {
		if iterations > maxIterations {
			// Break out if we've looped too many times.
			break
		}
		if p.curToken.Type == LBRACE {
			depth++
		} else if p.curToken.Type == RBRACE {
			depth--
			if depth == 0 {
				p.nextToken() // Move past the closing brace.
				return
			}
		}
		p.nextToken()
		iterations++
	}
}

func (p *Parser) parseOperationDefinition() *OperationDefinition {
	op := &OperationDefinition{}
	if p.curToken.Literal == "query" ||
		p.curToken.Literal == "mutation" ||
		p.curToken.Literal == "subscription" {
		op.Operation = p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == IDENT {
			op.Name = p.curToken.Literal
			p.nextToken()
		}
		if p.curToken.Type == LPAREN {
			op.VariableDefinitions = p.parseVariableDefinitions()
		}
	} else {
		op.Operation = "query"
	}
	if p.curToken.Type == LBRACE {
		op.SelectionSet = p.parseSelectionSet()
	}
	return op
}

func (p *Parser) parseTypeField() *Field {
	// Expect an identifier for the field name.
	if p.curToken.Type != IDENT {
		return nil
	}
	field := &Field{Name: p.curToken.Literal}
	p.nextToken()
	// If a colon is present, skip the colon and the type name.
	if p.curToken.Type == COLON {
		p.nextToken() // skip ':'
		// Optionally, capture the type if needed.
		if p.curToken.Type == IDENT {
			// p.curToken.Literal is the type name, e.g. "String"
			p.nextToken() // skip the type name
		}
	}
	return field
}

func (p *Parser) parseVariableDefinitions() []VariableDefinition {
	var vars []VariableDefinition
	p.nextToken() // Skip '('
	for p.curToken.Type != RPAREN && p.curToken.Type != EOF {
		if p.curToken.Type == DOLLAR {
			p.nextToken() // Skip '$'
			if p.curToken.Type != IDENT {
				return vars
			}
			varDef := VariableDefinition{}
			varDef.Variable = p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == COLON {
				p.nextToken()
				typeParsed := p.parseType()
				if typeParsed != nil {
					varDef.Type = *typeParsed
				}
			}
			vars = append(vars, varDef)
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // Skip ')'
	return vars
}

func (p *Parser) parseSelectionSet() *SelectionSet {
	ss := &SelectionSet{}
	p.nextToken() // skip '{'
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		sel := p.parseSelection()
		if sel != nil {
			ss.Selections = append(ss.Selections, sel)
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip '}'
	return ss
}

func (p *Parser) parseSelection() Selection {
	return p.parseField()
}

func (p *Parser) parseField() *Field {
	field := &Field{}
	if p.curToken.Type != IDENT {
		return nil
	}
	field.Name = p.curToken.Literal
	p.nextToken()
	if p.curToken.Type == LPAREN {
		field.Arguments = p.parseArguments()
	}
	if p.curToken.Type == LBRACE {
		field.SelectionSet = p.parseSelectionSet()
	}
	return field
}

func (p *Parser) parseArguments() []Argument {
	var args []Argument
	p.nextToken() // skip '('
	for p.curToken.Type != RPAREN && p.curToken.Type != EOF {
		arg := Argument{}
		if p.curToken.Type == IDENT {
			arg.Name = p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == COLON {
				p.nextToken()
				arg.Value = p.parseValue()
			}
			args = append(args, arg)
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip ')'
	return args
}

// parseObject parses a GraphQL object literal.
// It assumes the current token is the opening '{'.
func (p *Parser) parseObject() *Value {
	objFields := make(map[string]*Value)
	// Skip the '{'
	p.nextToken()
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		// Expect a field name (identifier) for the key.
		if p.curToken.Type != IDENT {
			// Error handling can be improved.
			return &Value{Kind: "Illegal", Literal: "expected object key"}
		}
		key := p.curToken.Literal
		p.nextToken()
		// Expect a colon.
		if p.curToken.Type != COLON {
			return &Value{Kind: "Illegal", Literal: "expected colon in object"}
		}
		p.nextToken() // skip colon
		// Parse the value recursively.
		value := p.parseValue()
		objFields[key] = value
		// If there's a comma, skip it.
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	// Skip the closing '}'
	p.nextToken()
	return &Value{
		Kind:         "Object",
		ObjectFields: objFields,
	}
}

func (p *Parser) parseArray() *Value {
	arr := []*Value{}
	p.nextToken() // skip '['
	for p.curToken.Type != RBRACKET && p.curToken.Type != EOF {
		val := p.parseValue()
		arr = append(arr, val)
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip ']'
	return &Value{Kind: "Array", List: arr}
}

// Update parseValue to handle objects.
func (p *Parser) parseValue() *Value {
	// If the current token indicates the start of an object literal.
	if p.curToken.Type == LBRACE {
		return p.parseObject()
	}
	if p.curToken.Type == LBRACKET {
		return p.parseArray()
	}

	val := &Value{}
	switch p.curToken.Type {
	case INT:
		val.Kind = "Int"
		val.Literal = p.curToken.Literal
		p.nextToken()
	case STRING:
		val.Kind = "String"
		val.Literal = p.curToken.Literal
		p.nextToken()
	case IDENT:
		// Handle booleans and enums.
		if p.curToken.Literal == "true" || p.curToken.Literal == "false" {
			val.Kind = "Boolean"
		} else {
			val.Kind = "Enum"
		}
		val.Literal = p.curToken.Literal
		p.nextToken()
	case DOLLAR:
		p.nextToken() // skip '$'
		if p.curToken.Type == IDENT {
			val.Kind = "Variable"
			val.Literal = p.curToken.Literal
			p.nextToken()
		} else {
			// No identifier after '$'; mark as a variable with an empty literal.
			val.Kind = "Variable"
			val.Literal = ""
		}

	default:
		val.Kind = "Illegal"
		val.Literal = p.curToken.Literal
		p.nextToken()
	}
	return val
}

func (p *Parser) parseType() *Type {
	var t Type
	if p.curToken.Type == LBRACKET {
		// This is a list type.
		p.nextToken()              // Skip '['
		innerType := p.parseType() // Recursively parse the inner type.
		t = Type{IsList: true, Elem: innerType}
		if p.curToken.Type != RBRACKET {
			// Handle error: expected closing bracket.
		}
		p.nextToken() // Skip ']'
		// Check for non-null on the list type.
		if p.curToken.Type == BANG {
			t.NonNull = true
			p.nextToken()
		}
		return &t
	} else if p.curToken.Type == IDENT {
		// Basic type.
		t = Type{Name: p.curToken.Literal}
		p.nextToken()
		// Check for non-null on the basic type.
		if p.curToken.Type == BANG {
			t.NonNull = true
			p.nextToken()
		}
		return &t
	}
	return nil
}
