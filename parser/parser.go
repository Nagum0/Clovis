package parser

import (
	"clovis/lexer"
	"clovis/semantics"
	"fmt"
	"strconv"
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
	if p.matchAny(lexer.UINT_64, lexer.UINT_32, lexer.UINT_16, lexer.UINT_8, lexer.BOOL) {
		return p.parseVarDecl()
	} else if p.matchAny(lexer.STAR, lexer.IDENT, lexer.OPEN_PAREN) {
		return p.parseVarDefinition()
	} else if p.match(lexer.OPEN_CURLY) {
		return p.parseBlockStmt()
	} else if p.match(lexer.IF) {
		return p.parseIfStmt()
	} else if p.match(lexer.WHILE) {
		p.parseWhileStmt()
	} else if p.match(lexer.FOR) {
		p.parseForStmt()
	} else if p.match(lexer.ASSERT) {
		return p.parseAssert()
	} else {
		return p.parseExpressionStmt()
	}

	return nil, nil
}

// <varDecl> ::= <typeID> { "*" | "[" UINT_LIT "]" } IDENT ( ";" | "=" <expression> ";" )
func (p *Parser) parseVarDecl() (*VarDeclStmt, error) {
	decl := VarDeclStmt{}
	decl.Type = p.getType(p.consume().Type)

	for p.matchAny(lexer.STAR, lexer.OPEN_BRACKET) {
		if p.match(lexer.STAR) {
			decl.Type = semantics.Ptr{ ValueType: decl.Type }
			p.consume() // '*'
		} else if p.match(lexer.OPEN_BRACKET) {
			p.consume() // '['

			if !p.match(lexer.UINT_64_LIT) {
				return nil, NewParserError(
					p.peek(),
					fmt.Sprintf("Expected size specifier for array declaration but received '%v'", p.consume().Value),
				)
			}
			sizeToken := p.consume()
		
			if !p.match(lexer.CLOSE_BRACKET) {
				return nil, NewParserError(
					p.peek(),
					fmt.Sprintf("Expected ']' after array declaration but received '%v'", p.consume().Value),
				)
			}
			p.consume() // ']'
			
			arrLength, _ := strconv.Atoi(sizeToken.Value)
			decl.Type = semantics.Array{ Base: decl.Type, Length: arrLength }
		}
	}

	if !p.match(lexer.IDENT) {
		return nil, NewParserError(
			p.peek(),
			fmt.Sprintf("Expected an identifer after type defintion but received '%v'", p.consume().Value),
		)
	}
	decl.Ident = p.consume()

	if p.match(lexer.ASSIGN) {
		p.consume() // '='
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		decl.Right.SetVal(right)
	}

	if !p.match(lexer.SEMI) {
		return nil, NewParserError(
			p.peek(),
			fmt.Sprintf("Expected ';' after variable declaration but received '%v'", p.consume().Value),
		)
	}
	p.consume() // ';'

	return &decl, nil
}

// <varDefinition> ::= <lvalue> "=" <expression> ";"
func (p *Parser) parseVarDefinition() (Statement, error) {
	varDefStmt := VarDefinitionStmt{}

	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	varDefStmt.Left = left

	if !p.match(lexer.ASSIGN) {
		return nil, NewParserError(
			p.peek(),
			fmt.Sprintf("Expected '=' for assignment but received '%v'", p.peek().Value),
		)
	}
	varDefStmt.Op = p.consume()

	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	varDefStmt.Right = right

	if !p.match(lexer.SEMI) {
		return nil, NewParserError(
			p.peek(),
			fmt.Sprintf("Expected ';' at the end of statement but received '%v'", p.peek().Value),
		)
	}

	p.consume() // ';'
	
	return &varDefStmt, nil
}

// <blockStmt> ::= "{" <statements> "}"
func (p *Parser) parseBlockStmt() (Statement, error) {
	p.consume() // '{'

	stmts := []Statement{}

	for p.idx < len(p.tokens) && !p.isAtEnd() && !p.match(lexer.CLOSE_CURLY) {
		stmt, err := p.parseStatement()
		if err != nil {
			p.Errors = append(p.Errors, err)
			p.synchronize()
			continue
		}
		stmts = append(stmts, stmt)
	}

	blockStmt := BlockStmt{
		Statements: stmts,
	}
	
	if !p.match(lexer.CLOSE_CURLY) {
		err := NewParserError(
			p.peek(),
			fmt.Sprintf("Expected '}' but found '%v'", p.peek().Value),
		)
		return nil, err
	} else {
		p.consume() // '}'
	}

	return &blockStmt, nil
}

// <ifStmt> ::= "if" <expression> <statement> ( "else" <statement> )
func (p *Parser) parseIfStmt() (Statement, error) {
	ifStmt := IfStmt{}
	ifStmt.IfToken = p.consume()
	
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	ifStmt.Condition = expr

	stmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	ifStmt.Stmt = stmt

	if p.match(lexer.ELSE) {
		p.consume() // 'else'
		elseStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		ifStmt.ElseStmt.SetVal(elseStmt)
	}
	
	return &ifStmt, nil
}

func (p *Parser) parseWhileStmt() {

}

func (p *Parser) parseForStmt() {

}

func (p *Parser) parseAssert() (Statement, error) {
	stmt := AssertStmt{}
	stmt.AssertToken = p.consume()

	expr, err := p.parseExpression()
	if err != nil {
		e := NewParserError(
			p.peek(),
			fmt.Sprintf("At assertion %v", err.Error()),
		)
		return nil, e
	}
	stmt.Expr = expr

	if !p.match(lexer.SEMI) {
		e := NewParserError(
			p.peek(),
			fmt.Sprintf("Expected ';' found %v", p.peek().Value),
		)
		return nil, e
	}
	p.consume() // ';'

	return &stmt, nil
}

func (p *Parser) parseExpressionStmt() (Statement, error) {
	exprStmt := ExpressionStmt{}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	exprStmt.Expr = expr

	if !p.match(lexer.SEMI) {
		e := NewParserError(
			p.peek(),
			fmt.Sprintf("Expected ';' found %v", p.peek().Value),
		)
		return nil, e
	}
	p.consume() // ';'

	return &exprStmt, nil
}

// <expression> ::= <equality>
func (p *Parser) parseExpression() (Expression, error) {
	return p.parseEquality()
}

// <equality> ::= <comparison> { ("==" | "!=") <comparison> }
func (p *Parser) parseEquality() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	
	if p.matchAny(lexer.EQ, lexer.NEQ) {
		op := p.consume()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Type: semantics.Undefined{},
			Left: left,
			Op: op,
			Right: right,
		}
	}

	return left, nil
}

// <comparison> ::= <term> { ("<" | "<=" | ">" | ">=") <term> }
func (p *Parser) parseComparison() (Expression, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.matchAny(lexer.LESS_THAN, lexer.LESS_EQ_THAN, lexer.GREATER_THAN, lexer.GREATER_EQ_THAN) {
		op := p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Type: semantics.Undefined{},
			Left: left,
			Op: op,
			Right: right,
		}
	}

	return left, nil
}

// <term> ::= <factor> { ("+" | "-") <factor> }
func (p *Parser) parseTerm() (Expression, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.matchAny(lexer.PLUS, lexer.MINUS) {
		op := p.consume()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Type: semantics.Undefined{},
			Left: left,
			Op: op,
			Right: right,
		}
	}

	return left, nil
}

// <factor> ::= <prefix> { ("*" | "/") <prefix> }
func (p *Parser) parseFactor() (Expression, error) {
	left, err := p.parsePrefix()
	if err != nil {
		return nil, err
	}

	for p.matchAny(lexer.STAR, lexer.F_SLASH) {
		op := p.consume()
		right, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpression{
			Type: semantics.Undefined{},
			Left: left,
			Op: op,
			Right: right,
		}
	}

	return left, nil
}

// <prefix> ::= ( "!" | "-" | "*" | "&" ) <prefix> | <postfix>
func (p *Parser) parsePrefix() (Expression, error) {
	if p.match(lexer.STAR) {
		derefExpr := DerefExpression{ Op: p.consume() }

		right, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}
		derefExpr.Right = right

		return &derefExpr, nil
	} else if p.match(lexer.AMPERSAND) {
		refExpr := ReferenceExpression{ Op: p.consume() }

		right, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}
		refExpr.Right = right

		return &refExpr, nil
	} else if p.matchAny(lexer.NOT, lexer.MINUS) {
		op := p.consume()
		right, err := p.parsePrefix()
		if err != nil {
			return nil, err
		}

		un := &PrefixExpression{
			Type: semantics.Undefined{},
			Op: op,
			Right: right,
		}

		return un, nil
	}

	return p.parsePostfix()
}

// <postfix> ::= <primary> { ( "++" | "--" | <arrayAccess> }
func (p *Parser) parsePostfix() (Expression, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	
	// TODO: ++ and -- postfix operators
	for p.match(lexer.OPEN_BRACKET) {
		expr, err := p.parseArrayAccess()
		if err != nil {
			return nil, err
		}
		arrayAccessExpr := expr.(*ArrayAccessExpression)
		arrayAccessExpr.Left = left
		left = arrayAccessExpr
	}

	return left, nil
}

// <primary> ::= <literal> | ident | "(" <expression> ")" 
func (p *Parser) parsePrimary() (Expression, error) {
	if p.matchAny(lexer.UINT_64_LIT, lexer.TRUE_LIT, lexer.FALSE_LIT) {
		litExpr := &LiteralExpression{
			Type: p.getType(p.peek().Type),
			Value: p.consume(),
		}
		return litExpr, nil
	} else if p.match(lexer.IDENT) {
		identExpr := &IdentExpression{
			Type: semantics.Undefined{},
			Ident: p.consume(),
		}
		return identExpr, nil
	} else if p.match(lexer.OPEN_PAREN) {
		return p.parseGroupExpr()
	} else {
		err := NewParserError(
			p.peek(),
			"Invalid expression",
		)
		return nil, err
	}
}

// <arrayAccess> := "[" <expression> "]"
func (p *Parser) parseArrayAccess() (Expression, error) {
	arrayAccessExpr := ArrayAccessExpression{ 
		Type: semantics.Undefined{},
		OpenBracket: p.consume(),
	}

	indexExpr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	arrayAccessExpr.IndexExpr = indexExpr

	if !p.match(lexer.CLOSE_BRACKET) {
		return nil, NewParserError(
			p.peek(),
			fmt.Sprintf("Expected ']' after array access but received '%v'", p.peek().Value),
		)
	}
	p.consume() // ']'

	return &arrayAccessExpr, nil
}

// <groupExpr> ::= "(" <expression> ")"
func (p *Parser) parseGroupExpr() (Expression, error) {
	groupExpr := &GroupExpression{
		Type: semantics.Undefined{},
	}

	p.consume()

    expr, err := p.parseExpression()
    if err != nil {
    	return nil, err
    }

	groupExpr.Expr = expr
    
    if !p.match(lexer.CLOSE_PAREN) {
    	err = NewParserError(
        	p.consume(),
        	"Expected ')' after group expression",
    	)

    	return nil, err
    }

    p.consume()

    return groupExpr, nil
}

func (p *Parser) consume() lexer.Token {
	t := p.tokens[p.idx]
	p.idx++
	return t
}

func (p *Parser) peek() lexer.Token {
	t := p.tokens[p.idx]
	return t
}

func (p *Parser) match(tokenType lexer.TokenType) bool {
	return p.tokens[p.idx].Type == tokenType
}

func (p *Parser) matchAny(tokenTypes ...lexer.TokenType) bool {
	for _, t := range tokenTypes {
		if p.tokens[p.idx].Type == t {
			return true
		}
	}

	return false
}

func (p *Parser) isAtEnd() bool {
	return p.tokens[p.idx].Type == lexer.EOF
}

func (p *Parser) synchronize() {
	for !p.isAtEnd() && p.matchAny(lexer.SEMI, lexer.CLOSE_CURLY, lexer.EOF) {
		p.consume()
	}

	p.consume()
}

func (p *Parser) getType(tokenType lexer.TokenType) semantics.Type {
	switch tokenType {
	case lexer.UINT_8:
		return semantics.Uint8{}
	case lexer.UINT_16:
		return semantics.Uint16{}
	case lexer.UINT_32:
		return semantics.Uint32{}
	case lexer.UINT_64:
		return semantics.Uint64{}
	case lexer.UINT_64_LIT:
		return semantics.UintLiteral{}
	case lexer.BOOL:
		fallthrough
	case lexer.TRUE_LIT:
		fallthrough
	case lexer.FALSE_LIT:
		return semantics.Bool{}
	}

	return semantics.Undefined{}
}
