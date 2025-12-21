package parser

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/semantics"
	"clovis/utils"
	"fmt"
	"regexp"
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
}

// Variable declaration statement.
type VarDeclStmt struct {
	Type   semantics.Type
	Ident  lexer.Token
	Right  utils.Optional[Expression]
	Symbol semantics.Symbol
}

func (stmt *VarDeclStmt) Semantics(s *semantics.SemanticChecker) error {
	if stmt.Right.HasVal() {
		right := stmt.Right.Value()
		if err := right.Semantics(s); err != nil {
			return err
		}

		if !stmt.Type.Equals(right.ExprType()) {
			return s.AddError(
				fmt.Sprintf(
					"Variable type %v and right side type %v do not match",
					stmt.Type.TypeID(),
					right.ExprType().TypeID(),
				),
				stmt.Ident,
			)
		}
	}

	if err := s.PushSymbol(stmt.Ident.Value, stmt.Type, stmt.Ident); err != nil {
		return err
	}
	stmt.Symbol, _ = s.TopSymbol()

	return nil
}

func (s VarDeclStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- VarDeclStmt -------------------------\n")
	fmt.Fprintf(
		e,
		"; type = %v ident = %v offset = %v size = %v\n",
		s.Type.TypeID(),
		s.Ident.Value,
		s.Symbol.Offset,
		s.Type.Size(),
	)

	size := s.Type.Size()
	fmt.Fprintf(e, "sub rsp, %v\n", size)

	if !s.Right.HasVal() {
		return
	}

	right := s.Right.Value()
	_, isArray := right.ExprType().(semantics.Array)
	if isArray {
		right.EmitCode(e)
		fmt.Fprintf(e, "mov rcx, %v\n", size)                    // Amount of bytes to move
		fmt.Fprintf(e, "mov rsi, rax\n")                         // rsi holds the source
		fmt.Fprintf(e, "lea rdi, [rbp - %v]\n", s.Symbol.Offset) // rdi holds the destination
		fmt.Fprintf(e, "rep movsb\n")
	} else {
		right.EmitCode(e)
		reg := s.Type.Register()
		asmSize := s.Type.ASMSize()
		fmt.Fprintf(e, "mov %v [rbp - %v], %v\n", asmSize, s.Symbol.Offset, reg)
	}
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

	_, isArray := stmt.Right.ExprType().(semantics.Array)
	if isArray {
		stmt.Right.EmitCode(e)
		size := stmt.Right.ExprType().Size()
		fmt.Fprintf(e, "mov rcx, %v\n", size)
		fmt.Fprintf(e, "mov rsi, rax\n")
		fmt.Fprintf(e, "pop rdi\n")
		fmt.Fprintf(e, "rep movsb\n")
	} else {
		stmt.Right.EmitCode(e)
		fmt.Fprintf(e, "pop rbx\n")
		fmt.Fprintf(
			e,
			"mov %v [rbx], %v\n",
			addr.ExprType().ASMSize(),
			addr.ExprType().Register(),
		)
	}
}

// A block statement holds a group of statements.
type BlockStmt struct {
	Statements []Statement
	// The size of the symbols declared inside this block.
	BlockSize int
}

func (stmt *BlockStmt) Semantics(s *semantics.SemanticChecker) error {
	s.PushBlock()

	for _, innerStmt := range stmt.Statements {
		// TODO: Implement better logging here
		innerStmt.Semantics(s)
	}

	stmt.BlockSize = s.PopBlock()
	return nil
}

func (stmt BlockStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- BlockStmt: Size = %v -------------------------\n", stmt.BlockSize)
	for _, innerStmt := range stmt.Statements {
		innerStmt.EmitCode(e)
	}

	fmt.Fprintf(e, "add rsp, %v\n", stmt.BlockSize)
}

// If statement.
type IfStmt struct {
	// The if token. Used for error handling.
	IfToken   lexer.Token
	Condition Expression
	Stmt      Statement
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

// While statement.
type WhileStmt struct {
	// Needed for error handling
	WhileToken lexer.Token
	Condition  Expression
	Body       Statement
}

func (stmt *WhileStmt) Semantics(s *semantics.SemanticChecker) error {
	if err := stmt.Condition.Semantics(s); err != nil {
		return err
	}

	if err := stmt.Body.Semantics(s); err != nil {
		return err
	}

	if stmt.Condition.ExprType().TypeID() != semantics.BOOL {
		return s.AddError(
			fmt.Sprintf(
				"While statement condition must be of type BOOL received %v",
				stmt.Condition.ExprType().TypeID(),
			),
			stmt.WhileToken,
		)
	}

	return nil
}

func (stmt WhileStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- WhileStmt ------------------------- \n")
	loopLabel := e.NextLabel()
	endLabel := e.NextLabel()
	fmt.Fprintf(e, "%v:\n", loopLabel)
	stmt.Condition.EmitCode(e)
	fmt.Fprintf(e, "cmp al, 1\n")
	fmt.Fprintf(e, "jne %v\n", endLabel)
	stmt.Body.EmitCode(e)
	fmt.Fprintf(e, "jmp %v\n", loopLabel)
	fmt.Fprintf(e, "%v:\n", endLabel)
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

// A function declaration and definition statement.
type FuncDeclaration struct {
	Ident    lexer.Token
	FuncType semantics.Func
	Body     []Statement
	Params   []semantics.Symbol
}

func (stmt *FuncDeclaration) Semantics(s *semantics.SemanticChecker) error {
	if err := s.PushSymbol(stmt.Ident.Value, stmt.FuncType, stmt.Ident); err != nil {
		return err
	}

	s.PushStackFrame()
	s.PushBlock()

	for paramIdent, paramType := range stmt.FuncType.Params {
		if err := s.PushSymbol(paramIdent, paramType, stmt.Ident); err != nil {
			return s.AddError(
				fmt.Sprintf("Error while pushing param %v to symbol table: %v", paramIdent, err.Error()),
				stmt.Ident,
			)
		}
		topSymbol, _ := s.TopSymbol()
		stmt.Params = append(stmt.Params, topSymbol)
	}

	for _, stmt := range stmt.Body {
		if err := stmt.Semantics(s); err != nil {
			return err
		}
	}

	// Check if there is a return at the end of the function if the
	// return type is not Undefined
	if !stmt.FuncType.Return.Equals(semantics.Undefined{}) {
		err := s.AddError(
			fmt.Sprintf(
				"Function %v with return type %v missing a return statement",
				stmt.Ident.Value, stmt.FuncType.Return.TypeID(),
			),
			stmt.Ident,
		)

		if len(stmt.Body) == 0 {
			return err
		}

		if _, isReturn := stmt.Body[len(stmt.Body)-1].(*ReturnStmt); !isReturn {
			return err
		}
	}

	if err := s.PopStackFrame(); err != nil {
		return err
	}
	s.PopBlock()

	return nil
}

// TODO: Implement reading rest of the parameters from the stack
func (stmt FuncDeclaration) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(
		e, "; ------------------------- Func %v %v ------------------------- \n",
		stmt.Ident.Value, stmt.FuncType.TypeID(),
	)

	fmt.Fprintf(e, "%v:\n", stmt.Ident.Value)
	fmt.Fprintf(e, "push rbp\nmov rbp, rsp\n")

	paramRegisters := []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
	paramNumLimit := min(len(paramRegisters), len(stmt.Params))

	fullParamsSize := 0
	for _, paramSymbol := range stmt.Params[0:paramNumLimit] {
		fullParamsSize += paramSymbol.Size
	}
	fmt.Fprintf(e, "sub rsp, %v\n", fullParamsSize)

	for idx, paramSymbol := range stmt.Params[0:paramNumLimit] {
		fmt.Fprintf(
			e, "mov %v [rbp - %v], %v\n",
			paramSymbol.Type.ASMSize(),
			paramSymbol.Offset,
			transformParamRegister(paramSymbol.Type, paramRegisters[idx]),
		)
	}

	for _, bodyStmt := range stmt.Body {
		bodyStmt.EmitCode(e)
	}

	// If no return type emit return implicitly
	if stmt.FuncType.Return.Equals(semantics.Undefined{}) {
		fmt.Fprintf(e, "leave\nret\n")
	}

	fmt.Fprintf(e, "\n")
}

// Transform the 64 bit register name to the correct subregister specified by the type.
func transformParamRegister(t semantics.Type, paramRegister64 string) string {
	isRRegister, _ := regexp.MatchString("r(1[0-5]|8|9)", paramRegister64)

	switch t.(type) {
	case semantics.Ptr, semantics.Array, semantics.Func, semantics.Uint64, semantics.UintLiteral:
		return paramRegister64
	case semantics.Uint32:
		if isRRegister {
			return paramRegister64 + "d"
		}
		return strings.Replace(paramRegister64, "r", "e", 1)
	case semantics.Uint16:
		if isRRegister {
			return paramRegister64 + "w"
		}
		return paramRegister64[1:2]
	case semantics.Uint8, semantics.Bool:
		if isRRegister {
			return paramRegister64 + "b"
		}
		registerNameEnd := paramRegister64[1:2]
		registerNameEnd = strings.Replace(registerNameEnd, "x", "l", 1)
		if registerNameEnd[0] == 's' || registerNameEnd[0] == 'd' || registerNameEnd[0] == 'b' {
			return registerNameEnd + "l"
		}
	}

	return paramRegister64
}

// A return statement.
type ReturnStmt struct {
	ReturnToken lexer.Token
	ReturnExpr  *utils.Optional[Expression]
	ReturnType  semantics.Type
}

func (stmt *ReturnStmt) Semantics(s *semantics.SemanticChecker) error {
	if stmt.ReturnType.Equals(semantics.Undefined{}) {
		if stmt.ReturnExpr.HasVal() {
			return s.AddError(
				fmt.Sprintf(
					"Return should be empty for return type %v",
					semantics.UNDEFINED,
				),
				stmt.ReturnToken,
			)
		}
		return nil
	} else {
		var returnExpr Expression
		if !stmt.ReturnExpr.HasVal() {
			return s.AddError(
				fmt.Sprintf(
					"Expected return type %v but received %v",
					stmt.ReturnType.TypeID(), semantics.UNDEFINED,
				),
				stmt.ReturnToken,
			)
		} else {
			returnExpr = stmt.ReturnExpr.Value()
		}

		if err := returnExpr.Semantics(s); err != nil {
			return err
		}

		if !stmt.ReturnType.Equals(returnExpr.ExprType()) {
			return s.AddError(
				fmt.Sprintf(
					"Incorrect return type. Expected %v received %v",
					stmt.ReturnType.TypeID(),
					returnExpr.ExprType().TypeID(),
				),
				stmt.ReturnToken,
			)
		}

		return nil
	}
}

func (stmt ReturnStmt) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ------------------------- ReturnStmt ------------------------- \n")
	if stmt.ReturnExpr.HasVal() {
		stmt.ReturnExpr.Value().EmitCode(e)
		fmt.Fprintf(e, "leave\n")
		fmt.Fprintf(e, "ret\n")
	}
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
	Op    lexer.Token
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

func (BinaryExpression) IsAddressable() bool {
	return false
}

// A prefix expression holds a unary operator and a right value.
type PrefixExpression struct {
	Type        semantics.Type
	Op          lexer.Token
	Right       Expression
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

// A dereference expression.
// Example:
//
//		uint32 y = *x; // *x returns the value (rvalue) stored at the location where x is pointing to
//	 *x = 69; // Moves value 69 to the location x is pointing to (here *x returns an lvalue)
type DerefExpression struct {
	Type  semantics.Type
	Op    lexer.Token
	Right Expression
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

// A refence expression.
// Example:
//
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

	exp.Type = semantics.Ptr{ValueType: exp.Right.ExprType()}

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

// An array access expression.
// Example:
//
//	uint32[3] xs;
//	xs[1] = 2;
//	assert xs[1] == 2;
type ArrayAccessExpression struct {
	Type      semantics.Type
	Left      Expression
	IndexExpr Expression
	// For error handling.
	OpenBracket lexer.Token
}

func (exp ArrayAccessExpression) ExprType() semantics.Type {
	return exp.Type
}

func (exp *ArrayAccessExpression) Semantics(s *semantics.SemanticChecker) error {
	if err := exp.Left.Semantics(s); err != nil {
		return err
	}

	array, isArray := exp.Left.ExprType().(semantics.Array)
	if !isArray {
		return s.AddError(
			fmt.Sprintf(
				"'[]' operator can be only used on arrays but received %v",
				exp.Left.ExprType().TypeID(),
			),
			exp.OpenBracket,
		)
	}
	exp.Type = array.Base

	return nil
}

func (exp ArrayAccessExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ArrayAccessExpression rvalue type = %v\n", exp.Type)
	addrExp, _ := exp.Left.(AddressableExpression)
	addrExp.EmitAddressCode(e)
	fmt.Fprintf(e, "push rax\n")
	exp.IndexExpr.EmitCode(e)
	fmt.Fprintf(e, "mov rbx, %v\n", exp.Type.Size())
	fmt.Fprintf(e, "mul rbx\n")
	fmt.Fprintf(e, "pop rbx\n")
	fmt.Fprintf(e, "mov %v, %v [rbx + rax]\n", exp.Type.Register(), exp.Type.ASMSize())
}

func (exp ArrayAccessExpression) EmitAddressCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; ArrayAccessExpression lvalue type = %v\n", exp.Type)
	addrExp, _ := exp.Left.(AddressableExpression)
	addrExp.EmitAddressCode(e)
	fmt.Fprintf(e, "push rax\n")
	exp.IndexExpr.EmitCode(e)
	fmt.Fprintf(e, "mov rbx, %v\n", exp.Type.Size())
	fmt.Fprintf(e, "mul rbx\n")
	fmt.Fprintf(e, "pop rbx\n")
	fmt.Fprintf(e, "lea rax, [rbx + rax]\n")
}

func (ArrayAccessExpression) IsAddressable() bool {
	return true
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
		case "false":
			value = "0"
		}
	}

	fmt.Fprintf(e, "; LiteralExpression: type = %v value = %v\n", exp.Type.TypeID(), value)
	fmt.Fprintf(e, "mov rax, %v\n", value)
}

func (exp LiteralExpression) IsAddressable() bool {
	return false
}

// A identifier expression holds an identifier's token.
type IdentExpression struct {
	Type   semantics.Type
	Ident  lexer.Token
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
	exp.Symbol = *symbol
	exp.Type = symbol.Type

	return nil
}

func (exp IdentExpression) EmitCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; IdentExpression rvalue type = %v\n", exp.Type.TypeID())
	_, isArray := exp.Type.(semantics.Array)
	if isArray {
		fmt.Fprintf(e, "lea rax, [rbp - %v]\n", exp.Symbol.Offset)
	} else {
		fmt.Fprintf(
			e,
			"mov %v, %v [rbp - %v]\n",
			exp.Type.Register(),
			exp.Type.ASMSize(),
			exp.Symbol.Offset,
		)
	}
}

func (exp IdentExpression) EmitAddressCode(e *codegen.Emitter) {
	fmt.Fprintf(e, "; IdentExpression lvalue type = %v\n", exp.Type.TypeID())
	fmt.Fprintf(e, "lea rax, [rbp - %v]\n", exp.Symbol.Offset)
}

func (exp IdentExpression) IsAddressable() bool {
	return true
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
	return exp.Expr.IsAddressable()
}
