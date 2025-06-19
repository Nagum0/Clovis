package lexer

import (
	"fmt"
	"unicode"
)

type LexerError struct {
	line int
	col  int
	val  string
}

func NewLexerError(val string, line int, col int) *LexerError {
	return &LexerError{ val: val, line: line, col: col }
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("Unrecognized token '%v' at line %v at col %v", e.val, e.line, e.col)
}

type Lexer struct {
	Tokens []Token
	Errors []*LexerError
	input  string
	line   int
	col    int
	idx	   int
	buffer string
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		Tokens: []Token{},
		Errors: []*LexerError{},
		input: input,
		line: 1,
		col: 1,
		buffer: "",
	}
}

func (l *Lexer) Lex() error {
	for l.idx < len(l.input) {
		if l.peek() == '\n' {
			l.line++
			l.col = 1
			l.idx++
		} else if unicode.IsSpace(l.peek()) {
			l.col++
			l.idx++
		} else if l.peek() == ';' {
			l.consume()
			l.emitToken(SEMI, l.col - 1)
		} else if l.peek() == '(' {
			l.consume()
			l.emitToken(OPEN_PAREN, l.col - 1)
		} else if l.peek() == ')' {
			l.consume()
			l.emitToken(CLOSE_PAREN, l.col - 1)
		} else if l.peek() == '{' {
			l.consume()
			l.emitToken(OPEN_CURLY, l.col - 1)
		} else if l.peek() == '}' {
			l.consume()
			l.emitToken(CLOSE_CURLY, l.col - 1)
		} else if l.peek() == '=' {
			l.consume()
			if l.peek() == '=' {
				l.consume()
				l.emitToken(EQ, l.col - 2)
			} else {
				l.emitToken(ASSIGN, l.col - 1)
			}
		} else if l.peek() == '!' {
			l.consume()
			if l.peek() == '=' {
				l.consume()
				l.emitToken(NEQ, l.col - 2)
			} else {
				l.emitToken(NOT, l.col - 1)
			}
		} else if l.peek() == '<' {
			l.consume()
			if l.peek() == '=' {
				l.emitToken(LESS_EQ_THAN, l.col - 2)
			} else {
				l.emitToken(LESS_THAN, l.col - 1)
			}
		} else if l.peek() == '>' {
			l.consume()
			if l.peek() == '=' {
				l.emitToken(GREATER_EQ_THAN, l.col - 2)
			} else {
				l.emitToken(GREATER_THAN, l.col - 1)
			}
		} else if l.peek() == '+' {
			l.consume()
			l.emitToken(PLUS, l.col - 1)
		} else if l.peek() == '-' {
			l.consume()
			l.emitToken(MINUS, l.col - 1)
		} else if l.peek() == '*' {
			l.consume()
			l.emitToken(STAR, l.col - 1)
		} else if l.peek() == '/' {
			l.consume()
			l.emitToken(F_SLASH, l.col - 1)
		} else if unicode.IsDigit(l.peek()) {
			startCol := l.col
			l.consume()	

			for l.idx < len(l.input) && unicode.IsDigit(l.peek()) {
				l.consume()
			}

			l.emitToken(UINT_64_LIT, startCol)
		} else if unicode.IsLetter(l.peek()) || l.peek() == '_' {
			startCol := l.col
			l.consume()

			for l.idx < len(l.input) && 
			    (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
				l.consume()
			}

			if l.isKeyword(startCol) {
				continue
			} else {
				l.emitToken(IDENT, startCol)
			}
		}
	}

	l.emitToken(EOF, 0)

	if errsLen := len(l.Errors); errsLen != 0 {
		return l.Errors[errsLen - 1]
	} else {
		return nil
	}
}

func (l *Lexer) emitToken(tokenType TokenType, startCol int) {
	l.Tokens = append(l.Tokens, *NewToken(tokenType, l.buffer, l.line, startCol))
	l.buffer = ""
}

func (l *Lexer) isKeyword(startCol int) bool {
	switch l.buffer {
	case "if":
		l.emitToken(IF, startCol)
		return true
	case "else":
		l.emitToken(ELSE, startCol)
		return true
	case "while":
		l.emitToken(WHILE, startCol)
		return true
	case "for":
		l.emitToken(FOR, startCol)
		return true
	case "uint64":
		l.emitToken(UINT_64, startCol)
		return true
	case "uint32":
		l.emitToken(UINT_32, startCol)
		return true
	case "uint16":
		l.emitToken(UINT_16, startCol)
		return true
	case "uint8":
		l.emitToken(UINT_8, startCol)
		return true
	case "bool":
		l.emitToken(BOOL, startCol)
		return true
	case "true":
		l.emitToken(TRUE_LIT, startCol)
		return true
	case "false":
		l.emitToken(FALSE_LIT, startCol)
		return true
	case "assert":
		l.emitToken(ASSERT, startCol)
		return true
	}

	return false
}

func (l *Lexer) consume() {
	l.buffer += string(l.input[l.idx])
	l.col++
	l.idx++
}

func (l *Lexer) peek() rune {
	if l.idx == len(l.input) {
		return 0
	}

	return rune(l.input[l.idx])
}
