package parser

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/utils"
	"fmt"
	"strings"
)

func indentStr(n int) string {
	return strings.Repeat("  ", n)
}

type Type string
const (
	UNKNOWN = "UNKNOWN"
	UINT = "UINT"
	BOOL = "BOOL"
)

// Anything that implements this interface can be used as a statement in the language.
type Statement interface {
	EmitCode(e *codegen.Emitter) error
	Print(indent int) string
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

func (s VarDeclStmt) Print(indent int) string {
	result := fmt.Sprintf("%vVarDeclStmt\n%v{\n", indentStr(indent), indentStr(indent))
	result += fmt.Sprintf("%vVarType: %v\n", indentStr(indent + 1), s.VarType)
	result += fmt.Sprintf("%vIdent: %v\n", indentStr(indent + 1), s.Ident.Value)

	if s.Value.HasVal() {
		result += fmt.Sprintf("%v", s.Value.Value().Print(indent + 1))
	}

	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}

// Anything that implements this interface is an expression in the language.
type Expression interface {
	ExprType() Type
	EmitCode(e *codegen.Emitter) error
	Print(indent int) string
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

func (exp BinaryExpression) Print(indent int) string {
	result := fmt.Sprintf("BinaryExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%v\n", exp.Left.Print(indent + 1))
	result += fmt.Sprintf("%vOp: %v\n", indentStr(indent + 1), exp.Op)
	result += fmt.Sprintf("%v", exp.Right.Print(indent + 1))
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
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

func (exp UnaryExpression) Print(indent int) string {
	result := fmt.Sprintf("UnaryExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%vOp: %v\n", indentStr(indent + 1), exp.Op)
	result += exp.Right.Print(indent + 1)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
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

func (exp LiteralExpression) Print(indent int) string {
	result := fmt.Sprintf("LiteralExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%vValue: %v", indentStr(indent + 1), exp.Value)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
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

func (exp IdentExpression) Print(indent int) string {
	result := fmt.Sprintf("IdentExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%vValue: %v", indentStr(indent + 1), exp.Ident.Value)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
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

func (exp GroupExpression) Print(indent int) string {
	result := fmt.Sprintf("GroupExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += exp.Expr.Print(indent + 1)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}
