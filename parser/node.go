package parser

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/semantics"
	"clovis/utils"
	"fmt"
	"strings"
)

func indentStr(n int) string {
	return strings.Repeat("  ", n)
}

// Represents a type for semantic analysis.
type Type string
const (
	UNKNOWN = "UNKNOWN"
	UINT = "UINT"
	BOOL = "BOOL"
)

// This interface represents a statement in the language
// and holds the needed functions for semantic analysis and code generation.
type Statement interface {
	// This checks whether a statement is semntically correct.
	// Also sets some extra information that is used by the emitter.
	Semantics(s *semantics.SemanticChecker) error

	// Using codegen.Emitter this emits the assembly code for the statement.
	EmitCode(e *codegen.Emitter) error

	// Pretty prints the statement.
	Print(indent int) string
}

// Variable declaration statement.
type VarDeclStmt struct {
	// The declared variables type.
	VarType Type
	// The variable's identifier.
	Ident	lexer.Token
	// An optional initializer value.
	// Can be any type of expression.
	Value	utils.Optional[Expression]
}

func (stmt *VarDeclStmt) Semantics(s *semantics.SemanticChecker) error {
	if stmt.Value.HasVal() {
		err := stmt.Value.Value().Semantics(s)
		if err != nil {
			// TODO: add semantic error
		}
	}

	if stmt.VarType != stmt.Value.Value().ExprType() {
		// TODO: add semantic error
	}

	// TODO: add to symbol table (check for redifinitions)

	return nil
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

// A variable definition statement.
type VarDefinitionStmt struct {
	// The variable's identifier.
	Ident lexer.Token
	// The value we want to set. Can be any expression.
	Value Expression
}

func (stmt *VarDefinitionStmt) Semantics(s *semantics.SemanticChecker) error {
	return nil
}

func (stmt VarDefinitionStmt) EmitCode(e *codegen.Emitter) error {
	return nil
}

func (stmt VarDefinitionStmt) Print(indent int) string {
	result := fmt.Sprintf("VarDefinitionStmt\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vIdent: %v\n", indentStr(indent + 1), stmt.Ident)
	result += fmt.Sprintf("%vValue: %v", indentStr(indent + 1), stmt.Value.Print(indent + 1))
	return fmt.Sprintf("%v%v\n}", indentStr(indent), result)
}

// This interface represents an expression in the language.
// and holds the needed functions for semantic analysis and code generation.
// All expressions must have a type that can be check with the ExprType() function.
type Expression interface {
	ExprType() Type
	// This checks whether the expression is semantically correct.
	// Also sets some extra information that is used by the emitter.
	Semantics(s *semantics.SemanticChecker) error

	// Using codegen.Emitter this emits the assembly code for the expression.
	EmitCode(e *codegen.Emitter) error

	// Pretty prints the expression.
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

func (expr *BinaryExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
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

// A unary expression holds a unary operator and a right value.
type UnaryExpression struct {
	Type  Type
	Op    lexer.TokenType
	Right Expression
}

func (exp UnaryExpression) ExprType() Type {
	return exp.Type
}

func (exp *UnaryExpression) Semantics(s *semantics.SemanticChecker) error {
	switch {
	case exp.Right.ExprType() == BOOL && exp.Op == lexer.NOT:
		fallthrough
	case exp.Right.ExprType() == UINT && exp.Op == lexer.MINUS:
		exp.Type = exp.Right.ExprType()
		break
	default:
		// TODO: add semantic error (incorrect operation)
	}

	return nil
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

func (exp *LiteralExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
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

func (exp *IdentExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
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

func (exp *GroupExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
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
