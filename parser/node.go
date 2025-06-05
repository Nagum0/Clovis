package parser

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/utils"
	"fmt"
)

type Type string
const (
	UNKNOWN = "UNKNOWN"
	UINT = "UINT"
	BOOL = "BOOL"
)

// Anything that implements this interface can be used as a statement in the language.
type Statement interface {
	EmitCode(e *codegen.Emitter) error
}

// Variable declaration statement.
// <varDecl> ::= ( "uint" | "bool" ) ident ";" | ( "uint" | "bool" ) ident "=" <expression> ";"
type VarDeclStmt struct {
	VarType lexer.TokenType
	Ident	lexer.Token
	Value	utils.Optional[Expression]
}

func (s VarDeclStmt) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (s VarDeclStmt) String() string {
	return fmt.Sprintf("VarDeclStmt:\n  %v  %v  %v", s.VarType, s.Ident.Value, s.Value)
}

// Anything that implements this interface is an expression in the language.
type Expression interface {
	ExprType() Type
	EmitCode(e *codegen.Emitter) error
}

// A binary expression holds a left value and a right value and an operator.
type BinaryExpression struct {
	Type  Type
	Left  Expression
	Op	  lexer.TokenType
	Right Expression
}

func (exp BinaryExpression) ExprType() Type {
	return exp.Type
}

func (exp BinaryExpression) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (exp BinaryExpression) String() string {
	return fmt.Sprintf("BinaryExpression:  %v  %v  %v", exp.Left, exp.Op, exp.Right)
}

// A unary expression holds a right value and an operator.
type UnaryExpression struct {
	Type  Type
	Op    lexer.TokenType
	Right Expression
}

func (exp UnaryExpression) ExprType() Type {
	return exp.Type
}

func (exp UnaryExpression) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (exp UnaryExpression) String() string {
	return fmt.Sprintf("UnaryExpression:  %v  %v", exp.Op, exp.Right)
}

// A literal expression holds a literal.
type LiteralExpression struct {
	Type  Type
	Value lexer.Token
}

func (exp LiteralExpression) ExprType() Type {
	return exp.Type
}

func (exp LiteralExpression) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (exp LiteralExpression) String() string {
	return fmt.Sprintf("LiteralExpression:  %v", exp.Value)
}

// A identifier expression holds an identifier's token.
type IdentExpression struct {
	Type  Type
	Ident lexer.Token
}

func (exp IdentExpression) ExprType() Type {
	return exp.Type
}

func (exp IdentExpression) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (exp IdentExpression) String() string {
	return fmt.Sprintf("IdentExpression:  %v", exp.Ident)
}

// A group expression holds an internal expression.
type GroupExpression struct {
	Type Type
	Expr Expression
}

func (exp GroupExpression) ExprType() Type {
	return exp.Type
}

func (exp GroupExpression) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (exp GroupExpression) String() string {
	return fmt.Sprintf("GroupExpression:  %v", exp.Expr)
}
