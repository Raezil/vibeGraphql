package vibeGraphql

import (
	"testing"
)

func TestLexer_Numbers(t *testing.T) {
	input := "12345 67890"
	lexer := NewLexer(input)

	// First number.
	tok := lexer.NextToken()
	if tok.Type != INT {
		t.Fatalf("expected token type INT, got %s", tok.Type)
	}
	if tok.Literal != "12345" {
		t.Errorf("expected literal '12345', got %q", tok.Literal)
	}

	// Second number.
	tok = lexer.NextToken()
	if tok.Type != INT {
		t.Fatalf("expected token type INT, got %s", tok.Type)
	}
	if tok.Literal != "67890" {
		t.Errorf("expected literal '67890', got %q", tok.Literal)
	}

	// End of input.
	tok = lexer.NextToken()
	if tok.Type != EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestLexer_Strings(t *testing.T) {
	input := `"hello world" "another string"`
	lexer := NewLexer(input)

	// First string.
	tok := lexer.NextToken()
	if tok.Type != STRING {
		t.Fatalf("expected token type STRING, got %s", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected literal 'hello world', got %q", tok.Literal)
	}

	// Second string.
	tok = lexer.NextToken()
	if tok.Type != STRING {
		t.Fatalf("expected token type STRING, got %s", tok.Type)
	}
	if tok.Literal != "another string" {
		t.Errorf("expected literal 'another string', got %q", tok.Literal)
	}

	// End of input.
	tok = lexer.NextToken()
	if tok.Type != EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestLexer_PunctuationAndSymbols(t *testing.T) {
	// Test all single-character tokens.
	input := `= : , ; ( ) { } [ ] $ !`
	lexer := NewLexer(input)

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{ASSIGN, "="},
		{COLON, ":"},
		{COMMA, ","},
		{SEMICOLON, ";"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{DOLLAR, "$"},
		{BANG, "!"},
		{EOF, ""},
	}

	for i, tt := range tests {
		tok := lexer.NextToken()
		if tok.Type != tt.expectedType {
			t.Errorf("token %d: expected type %s, got %s", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("token %d: expected literal %q, got %q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_IdentifierWithUnderscore(t *testing.T) {
	input := "foo_bar baz123"
	lexer := NewLexer(input)

	tok := lexer.NextToken()
	if tok.Type != IDENT {
		t.Fatalf("expected token type IDENT, got %s", tok.Type)
	}
	if tok.Literal != "foo_bar" {
		t.Errorf("expected literal 'foo_bar', got %q", tok.Literal)
	}

	tok = lexer.NextToken()
	if tok.Type != IDENT {
		t.Fatalf("expected token type IDENT, got %s", tok.Type)
	}
	if tok.Literal != "baz123" {
		t.Errorf("expected literal 'baz123', got %q", tok.Literal)
	}

	tok = lexer.NextToken()
	if tok.Type != EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestLexer_IllegalCharacter(t *testing.T) {
	input := "@"
	lexer := NewLexer(input)

	tok := lexer.NextToken()
	if tok.Type != ILLEGAL {
		t.Fatalf("expected token type ILLEGAL, got %s", tok.Type)
	}
	if tok.Literal != "@" {
		t.Errorf("expected literal '@', got %q", tok.Literal)
	}

	tok = lexer.NextToken()
	if tok.Type != EOF {
		t.Errorf("expected token type EOF, got %s", tok.Type)
	}
}

func TestNextToken(t *testing.T) {
	input := `=+(){},;"hello" 123`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{ASSIGN, "="},
		{ILLEGAL, "+"}, // '+' is not defined so it's considered illegal.
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"}, // Even though there's no '}' in input, our input has no '}', so this is a placeholder.
		{COMMA, ","},
		{SEMICOLON, ";"},
		{STRING, "hello"},
		{INT, "123"},
		{EOF, ""},
	}

	l := NewLexer(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType || tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - token wrong. expected type=%q, literal=%q; got type=%q, literal=%q", i, tt.expectedType, tt.expectedLiteral, tok.Type, tok.Literal)
		}
	}
}
