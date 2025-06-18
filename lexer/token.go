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
	UINT = "UINT"
	BOOL = "BOOL"
	ASSERT = "ASSERT"

	UINT_LIT = "UINT_LIT"
	TRUE_LIT = "TRUE_LIT"
	FALSE_LIT = "FALSE_LIT"
	IDENT = "IDENT"

	OPEN_PAREN = "OPEN_PAREN"
	CLOSE_PAREN = "CLOSE_PAREN"
	OPEN_CURLY = "OPEN_CURLY"
	CLOSE_CURLY = "CLOSE_CURLY"
	EQ = "EQ"
	NEQ = "NEQ"
	LESS_THAN = "LESS_THAN"
	LESS_EQ_THAN = "LESS_EQ_THAN"
	GREATER_THAN = "GREATER_THAN"
	GREATER_EQ_THAN = "GREATER_EQ_THAN"
	NOT = "NOT"
	PLUS = "PLUS"
	MINUS = "MINUS"
	STAR = "STAR"
	F_SLASH = "F_SLASH"
	ASSIGN = "ASSIGN"
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
