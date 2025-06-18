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
	fmt.Fprintf(
		e,
		"; VarDeclStmt: %v type = %v\n",
		s.Ident.Value, s.VarType,
	)
	fmt.Fprintf(e, "sub rsp, %v\n", s.VarType.Size())
	
	if s.Value.HasVal() {
		s.Value.Value().EmitCode(e)
		fmt.Fprintf(e, "mov %v [rbp - %v], %v\n", s.VarType.ASMSize(), s.Symbol.Offset, s.VarType.ASMExprReg())
	}

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
	Ident  lexer.Token
	// The value we want to set. Can be any expression.
	Value  Expression
	// Used by the code emitter to get the symbol data
	Symbol semantics.Symbol
}

func (stmt *VarDefinitionStmt) Semantics(s *semantics.SemanticChecker) error {
	if err := stmt.Value.Semantics(s); err != nil {
		return err
	}

	symbol, err := s.GetSymbol(stmt.Ident)
	if err != nil {
		return err
	}

	if symbol.Type != stmt.Value.ExprType() {
		return s.AddError(
			fmt.Sprintf(
				"Cannot set variable '%v' of type %v to type %v", 
				stmt.Ident.Value, 
				symbol.Type, 
				stmt.Value.ExprType(),
			),
			stmt.Ident,
		)
	}

	stmt.Symbol = *symbol

	return nil
}

func (stmt VarDefinitionStmt) EmitCode(e *codegen.Emitter) error {
	stmt.Value.EmitCode(e)
	fmt.Fprintf(
		e, 
		"; VarDefinitionStmt: %v type = %v offset = %v\n",
		stmt.Ident.Value, stmt.Symbol.Type, stmt.Symbol.Offset,
	)
	fmt.Fprintf(
		e, 
		"mov %v [rbp - %v], %v\n",
		stmt.Symbol.Type.ASMSize(), stmt.Symbol.Offset, stmt.Symbol.Type.ASMExprReg(),
	)

	return nil
}

func (stmt VarDefinitionStmt) Print(indent int) string {
	result := fmt.Sprintf("VarDefinitionStmt\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vIdent: %v\n", indentStr(indent + 1), stmt.Ident)
	result += fmt.Sprintf("%vValue: %v", indentStr(indent + 1), stmt.Value.Print(indent + 1))
	return fmt.Sprintf("%v%v\n}", indentStr(indent), result)
}

// A block statement holds a group of statements.
type BlockStmt struct {
	Statements []Statement
	// The size of the symbols declared inside this block.
	BlockSize  int
}

func (stmt *BlockStmt) Semantics(s *semantics.SemanticChecker) error {
	s.PushBlock()
	
	for _, innerStmt := range stmt.Statements {
		innerStmt.Semantics(s)
	}

	stmt.BlockSize = s.PopBlock()
	return nil
}

func (stmt BlockStmt) EmitCode(e *codegen.Emitter) error {
	fmt.Fprintf(e, "; BlockStmt: Size = %v", stmt.BlockSize)
	for _, innerStmt := range stmt.Statements {
		innerStmt.EmitCode(e)
	}

	fmt.Fprintf(e, "add rsp, %v\n", stmt.BlockSize)

	return nil
}

// TODO: Add symbol data info for block statement printing
// FIXME: Fix BlockStmt pretty printing.
func (stmt BlockStmt) Print(indent int) string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "\n%vBlockStmt\n%v{\n", indentStr(indent), indentStr(indent))
	fmt.Fprintf(&b, "%vBlockSize: %v\n", indentStr(indent + 1), stmt.BlockSize)
	fmt.Fprintf(&b, "%vStatements: \n", indentStr(indent + 1))

	for _, s := range stmt.Statements {
		fmt.Fprintf(&b, "%v,", s.Print(indent + 1))
	}

	fmt.Fprintf(&b, "\n%v}", indentStr(indent))
	return b.String()
}

// If statement.
type IfStmt struct {
	// The if token. Used for error handling.
	IfToken   lexer.Token
	Condition Expression
	Stmt	  Statement
	ElseStmt  utils.Optional[Statement]
}

func (stmt *IfStmt) Semantics(s *semantics.SemanticChecker) error {
	if err := stmt.Condition.Semantics(s); err != nil {
		return err
	}

	if stmt.Condition.ExprType() != semantics.BOOL {
		return s.AddError(
			fmt.Sprintf(
				"If statement condition must be of type BOOL received %v",
				stmt.Condition.ExprType(),
			),
			stmt.IfToken,
		)
	}

	if err := stmt.Stmt.Semantics(s); err != nil {
		return err
	}

	if !stmt.ElseStmt.HasVal() {
		return nil
	}

	if err := stmt.ElseStmt.Value().Semantics(s); err != nil {
		return err
	}

	return nil
}

func (stmt IfStmt) EmitCode(e *codegen.Emitter) error {
	fmt.Fprintf(e, "; IfStmt\n")
	stmt.Condition.EmitCode(e)
	fmt.Fprintf(e, "cmp al, 1\n")
	falseLabel := e.NextLabel()
	fmt.Fprintf(e, "jne %v\n", falseLabel)
	stmt.Stmt.EmitCode(e)
	fmt.Fprintf(e, "%v:\n", falseLabel)
	
	if stmt.ElseStmt.HasVal() {
		stmt.ElseStmt.Value().EmitCode(e)
	}

	return nil
}

func (stmt IfStmt) Print(indent int) string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "\n%vIfStmt\n%v{", indentStr(indent), indentStr(indent))
	fmt.Fprintf(&b, "%v\n", stmt.Condition.Print(indent + 1))
	fmt.Fprintf(&b, "%v\n", stmt.Stmt.Print(indent + 1))
	fmt.Fprintf(&b, "\n%v}", indentStr(indent))

	return b.String()
}

// Assert statement.
type AssertStmt struct {
	// The assert token. Used for error handling information.
	AssertToken lexer.Token
	Expr        Expression
}

func (stmt *AssertStmt) Semantics(s *semantics.SemanticChecker) error {
	if err := stmt.Expr.Semantics(s); err != nil {
		return err
	}

	if stmt.Expr.ExprType() != semantics.BOOL {
		return s.AddError(
			"Assert statement expects a boolean expression",
			stmt.AssertToken,
		)
	}

	return nil
}

func (stmt AssertStmt) EmitCode(e *codegen.Emitter) error {
	fmt.Fprintf(e, "; AssertStmt\n")
	stmt.Expr.EmitCode(e)
	fmt.Fprintf(e, "cmp al, 1\n")
	endLabel := e.NextLabel()
	fmt.Fprintf(e, "je %v\n", endLabel)
	fmt.Fprintf(e, "mov rax, 60\nmov rdi, 1\nsyscall\n")
	fmt.Fprintf(e, "%v:\n", endLabel)
	return nil
}

func (stmt AssertStmt) Print(indent int) string {
	b := strings.Builder{}
	
	fmt.Fprintf(&b, "\n%vAssertStmt\n%v{\n", indentStr(indent), indentStr(indent))
	fmt.Fprintf(&b, "%v\n", stmt.Expr.Print(indent + 1))
	fmt.Fprintf(&b, "\n%v}", indentStr(indent))

	return b.String()
}

// Expression statement.
type ExpressionStmt struct {
	Expr Expression
}

func (stmt *ExpressionStmt) Semantics(s *semantics.SemanticChecker) error {
	return stmt.Expr.Semantics(s)
}

func (stmt ExpressionStmt) EmitCode(e *codegen.Emitter) error {
	return stmt.Expr.EmitCode(e)
}

func (stmt ExpressionStmt) Print(indent int) string {
	b := strings.Builder{}

	fmt.Fprintf(&b, "\n%vExpressionStmt\n%v{\n", indentStr(indent), indentStr(indent))
	fmt.Fprintf(&b, "%v\n", stmt.Expr.Print(indent + 1))
	fmt.Fprintf(&b, "\n%v}", indentStr(indent))

	return b.String()
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
	// All expressions are evaluated in the rax register.
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
	}
	
	leftType := exp.Left.ExprType()
	opType := semantics.OperatorType(exp.Op)

	if (leftType == semantics.BOOL && opType == semantics.UINT) {
		return s.AddError(
			fmt.Sprintf(
				"Cannot use operator '%v' on type %v", 
				exp.Op.Value,
				leftType, 
			),
			exp.Op,
		)
	} else {
		exp.Type = semantics.OperatorType(exp.Op)
	}

	return nil
}

// Binary expressions are evaluated in the rax register.
func (exp BinaryExpression) EmitCode(e *codegen.Emitter) error {
	fmt.Fprintf(e, "; BinaryExpression: type = %v op = %v\n", exp.Type, exp.Op.Value)
	exp.Right.EmitCode(e)
	e.WriteString("push rax\n")
	exp.Left.EmitCode(e)
	e.WriteString("pop rbx\n")

	binOp := codegen.ASMBinaryOp(exp.Op)
	if binOp == "add" || binOp == "sub" {
		fmt.Fprintf(e, "%v rax, rbx\n", binOp)
	} else if binOp == "mul" || binOp == "div" {
		fmt.Fprintf(e, "%v rbx\n", binOp)
	} else if exp.Type == semantics.BOOL {
		fmt.Fprintf(e, "cmp rax, rbx\n")
		fmt.Fprintf(e, "%v al\n", binOp)
	}

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

// TODO: UnaryExpression.EmitCode
// Unary expressions are evaluated in the rax register.
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

// Literal expressions are evaluated in the rax register.
func (exp LiteralExpression) EmitCode(e *codegen.Emitter) error {
	value := exp.Value.Value
	if exp.Type == semantics.BOOL {
		switch exp.Value.Value {
		case "true":
			value = "1"
			break
		case "false":
			value = "0"
			break
		}
	}
	
	fmt.Fprintf(e, "; LiteralExpression: type = %v value = %v\n", exp.Type, value)
	fmt.Fprintf(e, "mov rax, %v\n", value)

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

// Identifier expressions are evaluated in the rax register.
func (exp IdentExpression) EmitCode(e *codegen.Emitter) error {
	fmt.Fprintf(
		e,
		"; IdentExpression: %v type = %v offset = %v\n",
		exp.Ident.Value, exp.Type, exp.Symbol.Offset,
	)
	fmt.Fprintf(
		e,
		"xor rax, rax\nmov %v, %v [rbp - %v]\n",
		exp.Type.ASMExprReg(), exp.Type.ASMSize(), exp.Symbol.Offset,
	)

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
	return exp.Expr.EmitCode(e)
}

func (exp GroupExpression) Print(indent int) string {
	result := fmt.Sprintf("GroupExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += exp.Expr.Print(indent + 1)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}
