package lexer

import "fmt"

type TokenType string
const (
	EOF = "EOF"
	SEMI = "SEMI"
	
	IF = "IF"
	ELSE = "ELSE"
	WHILE = "WHILE"
	FOR = "FOR"
	UINT_64 = "UINT_64"
	UINT_32 = "UINT_32"
	UINT_16 = "UINT_16"
	UINT_8 = "UINT_8"
	BOOL = "BOOL"
	ASSERT = "ASSERT"

	UINT_64_LIT = "UINT_64_LIT"
	TRUE_LIT = "TRUE_LIT"
	FALSE_LIT = "FALSE_LIT"
	IDENT = "IDENT"

	OPEN_PAREN = "OPEN_PAREN"
	CLOSE_PAREN = "CLOSE_PAREN"
	OPEN_CURLY = "OPEN_CURLY"
	CLOSE_CURLY = "CLOSE_CURLY"
	OPEN_BRACKET = "OPEN_BRACKET"
	CLOSE_BRACKET = "CLOSE_BRACKET"
	EQ = "EQ"
	NEQ = "NEQ"
	LESS_THAN = "LESS_THAN"
	LESS_EQ_THAN = "LESS_EQ_THAN"
	GREATER_THAN = "GREATER_THAN"
	GREATER_EQ_THAN = "GREATER_EQ_THAN"
	NOT = "NOT"
	PLUS = "PLUS"
	PLUS_PLUS = "PLUS_PLUS"
	MINUS = "MINUS"
	MINUS_MINUS = "MINUS_MINUS"
	STAR = "STAR"
	F_SLASH = "F_SLASH"
	ASSIGN = "ASSIGN"
	AMPERSAND = "AMPERSAND"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func NewToken(tokenType TokenType, value string, line int, col int) *Token {
	return &Token{
		Type: tokenType,
		Value: value,
		Line: line,
		Col: col,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v %v Line: %v Col: %v", t.Type, t.Value, t.Line, t.Col)
}
