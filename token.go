package vibeGraphql

// ----------------------
// Token Definitions
// ----------------------
type TokenType string

const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers and literals
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	STRING TokenType = "STRING"

	// Symbols
	ASSIGN    TokenType = "="
	COLON     TokenType = ":"
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
	LBRACKET  TokenType = "["
	RBRACKET  TokenType = "]"

	// GraphQL extras
	DOLLAR TokenType = "$"
	BANG   TokenType = "!"
)

type Token struct {
	Type    TokenType
	Literal string
}
