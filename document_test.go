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
