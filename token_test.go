package vibeGraphql

import "testing"

func TestTokenTypes(t *testing.T) {
	// Check that constant values are set as expected.
	if ASSIGN != "=" {
		t.Errorf("expected ASSIGN to be '=', got %s", ASSIGN)
	}
	if COLON != ":" {
		t.Errorf("expected COLON to be ':', got %s", COLON)
	}
	if COMMA != "," {
		t.Errorf("expected COMMA to be ',', got %s", COMMA)
	}
	if SEMICOLON != ";" {
		t.Errorf("expected SEMICOLON to be ';', got %s", SEMICOLON)
	}
	if LPAREN != "(" {
		t.Errorf("expected LPAREN to be '(', got %s", LPAREN)
	}
	if RPAREN != ")" {
		t.Errorf("expected RPAREN to be ')', got %s", RPAREN)
	}
	if LBRACE != "{" {
		t.Errorf("expected LBRACE to be '{', got %s", LBRACE)
	}
	if RBRACE != "}" {
		t.Errorf("expected RBRACE to be '}', got %s", RBRACE)
	}
	if LBRACKET != "[" {
		t.Errorf("expected LBRACKET to be '[', got %s", LBRACKET)
	}
	if RBRACKET != "]" {
		t.Errorf("expected RBRACKET to be ']', got %s", RBRACKET)
	}
	if DOLLAR != "$" {
		t.Errorf("expected DOLLAR to be '$', got %s", DOLLAR)
	}
	if BANG != "!" {
		t.Errorf("expected BANG to be '!', got %s", BANG)
	}
}
