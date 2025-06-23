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
	EmitCode(e *codegen.Emitter)
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
	// Used by the code emitter to get the symbol data.
	Symbol  semantics.Symbol
}

func (stmt *VarDeclStmt) Semantics(s *semantics.SemanticChecker) error {
	if stmt.Value.HasVal() {
		if err := stmt.Value.Value().Semantics(s); err != nil {
			return err
		}
		
		// Check type correctness
		if !(semantics.IsNumber(stmt.VarType) && semantics.IsNumber(stmt.Value.Value().ExprType())) && 
		   !(stmt.VarType.TypeID() == stmt.Value.Value().ExprType().TypeID()) {
			return s.AddError(
				fmt.Sprintf(
					"Declared type %v and initialized value type %v do not match",
					stmt.VarType.TypeID(),
					stmt.Value.Value().ExprType().TypeID(),
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

func (s VarDeclStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(
		e,
		"; ------------------------- VarDeclStmt: %v type = %v -------------------------\n",
		s.Ident.Value, s.VarType.TypeID(),
	)
	fmt.Fprintf(e, "sub rsp, %v\n", s.VarType.Size())
	
	if s.Value.HasVal() {
		s.Value.Value().EmitCode(e)
		fmt.Fprintf(e, "mov %v [rbp - %v], %v\n", s.VarType.ASMSize(), s.Symbol.Offset, s.VarType.Register())
	}
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
	Left  Expression
	Op    lexer.Token
	Right Expression
}

func (stmt *VarDefinitionStmt) Semantics(s *semantics.SemanticChecker) error {
	if err := stmt.Left.Semantics(s); err != nil {
		return err
	}

	if err := stmt.Right.Semantics(s); err != nil {
		return err
	}

	_, isAddr := stmt.Left.(AddressableExpression)
	if !isAddr {
		return s.AddError(
			"Left side of assignment only accepts addressable expressions",
			stmt.Op,
		)
	}

	if l, _ := stmt.Left.ExprType().CanUseOperator("=", stmt.Right.ExprType()); !l {
		return s.AddError(
			fmt.Sprintf(
				"Cannot assign type %v to address with type %v", 
				stmt.Right.ExprType().TypeID(),
				stmt.Left.ExprType().TypeID(),
			),
			stmt.Op,
		)
	}
	
	return nil
}

func (stmt VarDefinitionStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- VarDefinitionStmt -------------------------\n")
	addr, _ := stmt.Left.(AddressableExpression)
	addr.EmitAddressCode(e)
	fmt.Fprintf(e, "push rax\n")
	stmt.Right.EmitCode(e)
	fmt.Fprintf(e, "pop rbx\n")
	fmt.Fprintf(
		e, 
		"mov %v [rbx], %v\n",
		addr.ExprType().ASMSize(),
		addr.ExprType().Register(),
	)
}

func (stmt VarDefinitionStmt) Print(indent int) string {
	return ""
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

func (stmt BlockStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- BlockStmt: Size = %v -------------------------", stmt.BlockSize)
	for _, innerStmt := range stmt.Statements {
		innerStmt.EmitCode(e)
	}

	fmt.Fprintf(e, "add rsp, %v\n", stmt.BlockSize)
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

	if stmt.Condition.ExprType().TypeID() != semantics.BOOL {
		return s.AddError(
			fmt.Sprintf(
				"If statement condition must be of type BOOL received %v",
				stmt.Condition.ExprType().TypeID(),
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

func (stmt IfStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- IfStmt ------------------------- \n")
	stmt.Condition.EmitCode(e)
	fmt.Fprintf(e, "cmp al, 1\n")
	falseLabel := e.NextLabel()
	fmt.Fprintf(e, "jne %v\n", falseLabel)
	stmt.Stmt.EmitCode(e)
	endLabel := e.NextLabel()
	fmt.Fprintf(e, "jmp %v\n", endLabel)
	fmt.Fprintf(e, "%v:\n", falseLabel)
	
	if stmt.ElseStmt.HasVal() {
		stmt.ElseStmt.Value().EmitCode(e)
	}

	fmt.Fprintf(e, "%v:\n", endLabel)
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

	if stmt.Expr.ExprType().TypeID() != semantics.BOOL {
		return s.AddError(
			"Assert statement expects a boolean expression",
			stmt.AssertToken,
		)
	}

	return nil
}

func (stmt AssertStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- AssertStmt ------------------------- \n")
	stmt.Expr.EmitCode(e)
	fmt.Fprintf(e, "cmp al, 1\n")
	endLabel := e.NextLabel()
	fmt.Fprintf(e, "je %v\n", endLabel)
	fmt.Fprintf(e, "mov rax, 60\nmov rdi, 1\nsyscall\n")
	fmt.Fprintf(e, "%v:\n", endLabel)
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

func (stmt ExpressionStmt) EmitCode(e *codegen.Emitter) {
	stmt.Expr.EmitCode(e)
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
	EmitCode(e *codegen.Emitter)
	// Returns whether the expression is addressable.
	IsAddressable() bool
	// Pretty prints the expression.
	Print(indent int) string
}

// An AddressableExpression implements everything that a Expression implements
// it just represents expressions that have a memory location or point to one.
type AddressableExpression interface {
	Expression
	// Moves the address of the expression into the rax register.
	EmitAddressCode(e *codegen.Emitter)
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
		
	l, t := exp.Left.ExprType().CanUseOperator(exp.Op.Value, exp.Right.ExprType())
	if !l {
		return s.AddError(
			fmt.Sprintf(
				"Cannot use operator '%v' between types %v and %v", 
				exp.Op.Value,
				exp.Left.ExprType().TypeID(), 
				exp.Right.ExprType().TypeID(),
			),
			exp.Op,
		)
	}
	exp.Type = t
	
	return nil
}

// Binary expressions are evaluated in the rax register.
func (exp BinaryExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; BinaryExpression: type = %v op = %v\n", exp.Type.TypeID(), exp.Op.Value)
	exp.Right.EmitCode(e)
	e.WriteString("push rax\n")
	exp.Left.EmitCode(e)
	e.WriteString("pop rbx\n")
	
	// TODO: Clean up the binary operation logic
	binOp := codegen.ASMBinaryOp(exp.Op)
	if binOp == "add" || binOp == "sub" {
		fmt.Fprintf(e, "%v rax, rbx\n", binOp)
	} else if binOp == "mul" || binOp == "div" {
		fmt.Fprintf(e, "%v rbx\n", binOp)
	} else if exp.Type.TypeID() == semantics.BOOL {
		fmt.Fprintf(e, "cmp rax, rbx\n")
		fmt.Fprintf(e, "%v al\n", binOp)
	}
}

func (_ BinaryExpression) IsAddressable() bool {
	return false
}

func (exp BinaryExpression) Print(indent int) string {
	result := fmt.Sprintf("BinaryExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%v\n", exp.Left.Print(indent + 1))
	result += fmt.Sprintf("%vOp: %v\n", indentStr(indent + 1), exp.Op)
	result += fmt.Sprintf("%v", exp.Right.Print(indent + 1))
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}

// A prefix expression holds a unary operator and a right value.
type PrefixExpression struct {
	Type        semantics.Type
	Op    	    lexer.Token
	Right 	    Expression
	Addressable bool
}

func (exp PrefixExpression) ExprType() semantics.Type {
	return exp.Type
}

// TODO: PrefixExpression.Semantics for "-" and "!"
func (exp *PrefixExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
}

// TODO: PrefixExpression.EmitCode for "-" and "!"
func (exp PrefixExpression) EmitCode(e *codegen.Emitter) {

}

func (exp PrefixExpression) IsAddressable() bool {
	return exp.Addressable
}

func (exp PrefixExpression) Print(indent int) string {
	result := fmt.Sprintf("PrefixExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += fmt.Sprintf("%vOp: %v\n", indentStr(indent + 1), exp.Op)
	result += exp.Right.Print(indent + 1)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}

// TODO: PostfixExpression
// A postfix expression holds a unary operator and a left value.
type PostfixExpression struct {
	Type        semantics.Type
	Left        Expression
	Op          lexer.Token
	Addressable bool
}

func (exp PostfixExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *PostfixExpression) Semantics(s *semantics.SemanticChecker) error {
	return nil
}

func (exp PostfixExpression) EmitCode(e *codegen.Emitter) {
	
}

func (exp PostfixExpression) IsAddressable() bool {
	return exp.Addressable
}

func (exp PostfixExpression) Print(indent int) string {
	return ""
}

// A dereference expression.
// Example:
//	uint32 y = *x; // *x returns the value (rvalue) stored at the location where x is pointing to
//  *x = 69; // Moves value 69 to the location x is pointing to (here *x returns an lvalue)
type DerefExpression struct {
	Type   semantics.Type
	Op     lexer.Token
	Right  Expression
}

func (exp DerefExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *DerefExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Right.Semantics(s); err != nil {
		return err
	}
	
	ptr, isPtr := exp.Right.ExprType().(semantics.Ptr)
	if !isPtr {
		return s.AddError(
			 fmt.Sprintf(
				 "'*' dereference operator expected a PTR not %v", 
				 exp.Right.ExprType().TypeID(),
			 ),
			 exp.Op,
		)
	}
	exp.Type = ptr.ValueType
	
	return nil
}

func (exp DerefExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; DerefExpression rvalue type = %v\n", exp.Type.TypeID())
	exp.Right.EmitCode(e)
	fmt.Fprintf(e, "mov %v, %v [rax]\n", exp.Type.Register(), exp.Type.ASMSize())
}

func (exp DerefExpression) EmitAddressCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; DerefExpression lvalue type = %v\n", exp.Type.TypeID())
	exp.Right.EmitCode(e)
}

func (exp DerefExpression) IsAddressable() bool {
	return true
}

func (exp DerefExpression) Print(indent int) string {
	return ""
}

// A refence expression.
// Example:
//	uint32* xPtr = &x; // &x returns the address of x
type ReferenceExpression struct {
	Type  semantics.Type
	Op    lexer.Token
	Right Expression
}

func (exp ReferenceExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *ReferenceExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Right.Semantics(s); err != nil {
		return err
	}

	if !exp.Right.IsAddressable() {
		return s.AddError(
			"Expected an addressable expression",
			exp.Op,
		)
	}

	exp.Type = semantics.Ptr{ ValueType: exp.Right.ExprType() }

	return nil
}

func (exp ReferenceExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ReferenceExpression type = %v\n", exp.Type.TypeID())
	addr, _ := exp.Right.(AddressableExpression)
	addr.EmitAddressCode(e)
}

func (exp ReferenceExpression) EmitAddressCode(e *codegen.Emitter) {
	// rax already holds the address
	panic("&&x or &(&x) is incorrect usage")
}

func (exp ReferenceExpression) IsAddressable() bool {
	return true
}

func (exp ReferenceExpression) Print(indent int) string {
	return ""
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
func (exp LiteralExpression) EmitCode(e *codegen.Emitter) {
	value := exp.Value.Value
	if exp.Type.TypeID() == semantics.BOOL {
		switch exp.Value.Value {
		case "true":
			value = "1"
			break
		case "false":
			value = "0"
			break
		}
	}
	
	fmt.Fprintf(e, "; LiteralExpression: type = %v value = %v\n", exp.Type.TypeID(), value)
	fmt.Fprintf(e, "mov rax, %v\n", value)
}

func (exp LiteralExpression) IsAddressable() bool {
	return false
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
	// Used during code generation for symbol data.
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

func (exp IdentExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(
		e,
		"; IdentExpression %v type = %v offset = %v\n",
		exp.Ident.Value, exp.Type.TypeID(), exp.Symbol.Offset,
	)
	fmt.Fprintf(
		e,
		"xor rax, rax\nmov %v, %v [rbp - %v]\n",
		exp.Type.Register(), exp.Type.ASMSize(), exp.Symbol.Offset,
	)
}

func (exp IdentExpression) EmitAddressCode(e *codegen.Emitter) {
	fmt.Fprintf(
		e, 
		"; IdentExpression %v lea\n",
		exp.Ident.Value,
	)
	fmt.Fprintf(e, "lea rax, [rbp - %v]\n", exp.Symbol.Offset)
}

func (exp IdentExpression) IsAddressable() bool {
	return true
}

func (exp IdentExpression) Print(indent int) string {
	result := fmt.Sprintf("IdentExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type.TypeID())
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

func (exp GroupExpression) EmitCode(e *codegen.Emitter) {
	exp.Expr.EmitCode(e)
}

func (exp GroupExpression) EmitAddressCode(e *codegen.Emitter) {
	addr, isAddr := exp.Expr.(AddressableExpression)
	if isAddr {
		addr.EmitAddressCode(e)
	}
}

func (exp GroupExpression) IsAddressable() bool {
	return exp.IsAddressable()
}

func (exp GroupExpression) Print(indent int) string {
	result := fmt.Sprintf("GroupExpression\n%v{\n", indentStr(indent))
	result += fmt.Sprintf("%vType: %v\n", indentStr(indent + 1), exp.Type)
	result += exp.Expr.Print(indent + 1)
	return fmt.Sprintf("%v%v\n%v}", indentStr(indent), result, indentStr(indent))
}
