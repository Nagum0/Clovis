package parser

import (
	"clovis/lexer"
	"fmt"
)

type ParserError struct {
	token lexer.Token
	msg   string
}

func NewParserError(token lexer.Token, msg string) *ParserError {
	return &ParserError{
		token: token,
		msg: msg,
	}
}

func (e *ParserError) Error() string {
	return fmt.Sprintf("Error at line %v at column %v at token %v\n\t%v", e.token.Line, e.token.Col, e.token.Type, e.msg)
}

type Parser struct {
	Stmts  []Statement
	Errors []error
	tokens []lexer.Token
	idx    int
}

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		Stmts: []Statement{},
		Errors: []error{},
		tokens: tokens,
		idx: 0,
	}
}

func (p *Parser) Parse() error {
	p.parseProgram()

	if errLen := len(p.Errors); errLen != 0 {
		return p.Errors[errLen - 1]
	}

	return nil
}

func (p *Parser) parseProgram() {
	p.Stmts = p.parseStatements()
}

func (p *Parser) parseStatements() []Statement {
	stmts := []Statement{}

	for p.idx < len(p.tokens) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			p.Errors = append(p.Errors, err)
			p.synchronize()
			continue
		}
		stmts = append(stmts, stmt)
	}

	return stmts
}

func (p *Parser) parseStatement() (Statement, error) {
	if p.match(lexer.UINT) || p.match(lexer.BOOL) {
		return p.parseVarDecl()
	} else if p.match(lexer.IDENT) {
		p.parseVarDefinition()
	} else if p.match(lexer.OPEN_CURLY) {
		p.parseBlockStmt()
	} else if p.match(lexer.IF) {
		p.parseIfStmt()
	} else if p.match(lexer.WHILE) {
		p.parseWhileStmt()
	} else if p.match(lexer.FOR) {
		p.parseForStmt()
	}

	return nil, nil
}

// <varDecl> ::= ( "uint" | "bool" ) ident ";" | 
//               ( "uint" | "bool" ) ident "=" <expression> ";"
func (p *Parser) parseVarDecl() (*VarDeclStmt, error) {
	varDeclStmt := &VarDeclStmt{}
	varTypeToken := p.consume()
	varDeclStmt.VarType = varTypeToken.Type

	if p.match(lexer.IDENT) {
		varDeclStmt.Ident = p.consume()
	} else {
		err := NewParserError(
			p.consume(),
			fmt.Sprintf("Expected and identifier after %v", varTypeToken.Type),
		)
		return nil, err
	}

	if p.match(lexer.ASSIGN) {
		p.consume()
		valExpr := p.parseExpression()
		varDeclStmt.Value.SetVal(valExpr)
	} else if p.match(lexer.SEMI) {
		p.consume()
	} 

	if varDeclStmt.Value.HasVal() && !p.match(lexer.SEMI) {
		err := NewParserError(
			p.consume(),
			"Expected a ';' after variable declaration",
		)
		return nil, err
	} else {
		p.consume()
	}

	return varDeclStmt, nil
}

func (p *Parser) parseVarDefinition() {

}

func (p *Parser) parseBlockStmt() {

}

func (p *Parser) parseIfStmt() {

}

func (p *Parser) parseWhileStmt() {

}

func (p *Parser) parseForStmt() {

}

func (p *Parser) parseExpression() Expression {
	p.consume()
	return nil
}

func (p *Parser) consume() lexer.Token {
	t := p.tokens[p.idx]
	p.idx++
	return t
}

func (p *Parser) match(tokenType lexer.TokenType) bool {
	return p.tokens[p.idx].Type == tokenType
}

func (p *Parser) isAtEnd() bool {
	return p.tokens[p.idx].Type == lexer.EOF
}

func (p *Parser) synchronize() {
	for !p.isAtEnd() && p.tokens[p.idx].Type != lexer.SEMI {
		p.consume()
	}
	p.consume()
}
