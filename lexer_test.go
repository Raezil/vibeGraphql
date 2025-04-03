package vibeGraphql

import "testing"

func TestLexer(t *testing.T) {
	input := `query { hello }`
	lexer := NewLexer(input)
	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	// Expected token types for the given input.
	expected := []TokenType{IDENT, LBRACE, IDENT, RBRACE, EOF}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("token %d: expected %s, got %s", i, exp, tokens[i].Type)
		}
	}
}
