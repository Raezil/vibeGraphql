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
	}
	return doc
}

func (p *Parser) parseDefinition() Definition {
	if p.curToken.Literal == "query" ||
		p.curToken.Literal == "mutation" ||
		p.curToken.Literal == "subscription" {
		return p.parseOperationDefinition()
	} else if p.curToken.Type == LBRACE {
		// implicit query
		return p.parseOperationDefinition()
	}
	return nil
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
	p.nextToken() // skip '('
	for p.curToken.Type != RPAREN && p.curToken.Type != EOF {
		if p.curToken.Type == DOLLAR {
			p.nextToken() // skip '$'
			if p.curToken.Type != IDENT {
				return vars
			}
			varDef := VariableDefinition{}
			varDef.Variable = p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == COLON {
				p.nextToken()
				if p.curToken.Type == IDENT {
					varDef.Type.Name = p.curToken.Literal
					p.nextToken()
					if p.curToken.Type == BANG {
						varDef.Type.NonNull = true
						p.nextToken()
					}
				}
			}
			vars = append(vars, varDef)
		}
		if p.curToken.Type == COMMA {
			p.nextToken()
		}
	}
	p.nextToken() // skip ')'
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

// Update parseValue to handle objects.
func (p *Parser) parseValue() *Value {
	// If the current token indicates the start of an object literal.
	if p.curToken.Type == LBRACE {
		return p.parseObject()
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
