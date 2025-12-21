package lexer

import "fmt"

type TokenType string

const (
	EOF  TokenType = "EOF"
	SEMI TokenType = "SEMI"

	IF      TokenType = "IF"
	ELSE    TokenType = "ELSE"
	WHILE   TokenType = "WHILE"
	FOR     TokenType = "FOR"
	UINT_64 TokenType = "UINT_64"
	UINT_32 TokenType = "UINT_32"
	UINT_16 TokenType = "UINT_16"
	UINT_8  TokenType = "UINT_8"
	BOOL    TokenType = "BOOL"
	ASSERT  TokenType = "ASSERT"
	FUNC    TokenType = "FN"
	RETURN  TokenType = "RETURN"

	UINT_64_LIT TokenType = "UINT_64_LIT"
	TRUE_LIT    TokenType = "TRUE_LIT"
	FALSE_LIT   TokenType = "FALSE_LIT"
	IDENT       TokenType = "IDENT"

	OPEN_PAREN      TokenType = "OPEN_PAREN"
	CLOSE_PAREN     TokenType = "CLOSE_PAREN"
	OPEN_CURLY      TokenType = "OPEN_CURLY"
	CLOSE_CURLY     TokenType = "CLOSE_CURLY"
	OPEN_BRACKET    TokenType = "OPEN_BRACKET"
	CLOSE_BRACKET   TokenType = "CLOSE_BRACKET"
	EQ              TokenType = "EQ"
	NEQ             TokenType = "NEQ"
	LESS_THAN       TokenType = "LESS_THAN"
	LESS_EQ_THAN    TokenType = "LESS_EQ_THAN"
	GREATER_THAN    TokenType = "GREATER_THAN"
	GREATER_EQ_THAN TokenType = "GREATER_EQ_THAN"
	NOT             TokenType = "NOT"
	PLUS            TokenType = "PLUS"
	PLUS_PLUS       TokenType = "PLUS_PLUS"
	MINUS           TokenType = "MINUS"
	MINUS_MINUS     TokenType = "MINUS_MINUS"
	STAR            TokenType = "STAR"
	F_SLASH         TokenType = "F_SLASH"
	ASSIGN          TokenType = "ASSIGN"
	AMPERSAND       TokenType = "AMPERSAND"
	ARROW           TokenType = "ARROW"
	COMMA           TokenType = "COMMA"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func NewToken(tokenType TokenType, value string, line int, col int) *Token {
	return &Token{
		Type:  tokenType,
		Value: value,
		Line:  line,
		Col:   col,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("%v %v Line: %v Col: %v", t.Type, t.Value, t.Line, t.Col)
}
