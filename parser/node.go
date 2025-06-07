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
	VarType semantics.Type
	// The variable's identifier.
	Ident	lexer.Token
	// An optional initializer value.
	// Can be any type of expression.
	Value	utils.Optional[Expression]
	// Used by the code emitter to get the symbol data
	Symbol  semantics.Symbol
}

func (stmt *VarDeclStmt) Semantics(s *semantics.SemanticChecker) error {
	if stmt.Value.HasVal() {
		if err := stmt.Value.Value().Semantics(s); err != nil {
			return err
		}

		if stmt.VarType != stmt.Value.Value().ExprType() {
			return s.AddError(
				fmt.Sprintf(
					"Declared type %v and initialized value type %v do not match",
					stmt.VarType,
					stmt.Value.Value().ExprType(),
				),
				stmt.Ident,
			)
		}
	}

	if err := s.PushSymbol(stmt.Ident.Value, stmt.VarType, stmt.Ident); err != nil {
		return err
	}

	stmt.Symbol, _ = s.TopSymbol()

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
	ExprType() semantics.Type
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
	Type  semantics.Type
	Left  Expression
	Op	  lexer.Token
	Right Expression
}

func (exp BinaryExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *BinaryExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Left.Semantics(s); err != nil {
		return err
	}

	if err := exp.Right.Semantics(s); err != nil {
		return err
	}

	if exp.Left.ExprType() != exp.Right.ExprType() {
		return s.AddError(
			fmt.Sprintf(
				"Cannot use operator '%v' between types %v and %v", 
				exp.Op.Value,
				exp.Left.ExprType(), 
				exp.Right.ExprType(),
			),
			exp.Op,
		)
	} else {
		exp.Type = exp.Left.ExprType()
	}

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
	Type  semantics.Type
	Op    lexer.Token
	Right Expression
}

func (exp UnaryExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *UnaryExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Right.Semantics(s); err != nil {
		return err
	}
	
	// Check correct usage of operation
	switch {
	case exp.Right.ExprType() == semantics.BOOL && exp.Op.Type == lexer.NOT:
		fallthrough
	case exp.Right.ExprType() == semantics.UINT && exp.Op.Type == lexer.MINUS:
		exp.Type = exp.Right.ExprType()
		break
	default:
		return s.AddError(
			fmt.Sprintf("Cannot use operator '%v' on type %v", exp.Op.Value, exp.Right.ExprType()),
			exp.Op,
		)
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
	Type  semantics.Type
	Value lexer.Token
}

func (exp LiteralExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *LiteralExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil // No semantics needed
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
	Type   semantics.Type
	Ident  lexer.Token
	// Used during code generation for symbol data
	Symbol semantics.Symbol
}

func (exp IdentExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *IdentExpression) Semantics(s *semantics.SemanticChecker) error {
	symbol, err := s.GetSymbol(exp.Ident)
	if err != nil {
		return err
	}

	exp.Type = symbol.Type
	exp.Symbol = *symbol

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
	Type semantics.Type
	Expr Expression
}

func (exp GroupExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *GroupExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Expr.Semantics(s); err != nil {
		return err
	} else {
		exp.Type = exp.Expr.ExprType()
	}

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
