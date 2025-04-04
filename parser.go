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
	// Assume current token is "type".
	p.nextToken() // Skip the "type" keyword.
	if p.curToken.Type != IDENT {
		// Error handling: expected type name.
		return nil
	}
	typeName := p.curToken.Literal
	p.nextToken() // Move past the type name.

	// Expect an opening brace.
	if p.curToken.Type != LBRACE {
		// Error handling: expected '{'.
		return nil
	}
	p.nextToken() // Skip '{'.

	var fields []*Field
	// Parse all fields until we reach the closing brace.
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		field := p.parseField()
		if field != nil {
			fields = append(fields, field)
		}
		// Optionally skip commas if used.
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // Skip '}'.

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
	// Instead of skipping "type" definitions, parse them.
	if p.curToken.Literal == "type" {
		return p.parseTypeDefinition()
	}
	// If the token isn't recognized, advance and return nil.
	p.nextToken()
	return nil
}

// skipTypeDefinition advances the parser past a type definition block.
func (p *Parser) skipTypeDefinition() {
	// Skip the "type" keyword.
	p.nextToken()
	// Skip the type name (if present).
	if p.curToken.Type == IDENT {
		p.nextToken()
	}
	// If there's a block starting with '{', skip its entirety.
	if p.curToken.Type == LBRACE {
		p.skipBlock()
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
