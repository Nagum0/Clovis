package semantics

import (
	"fmt"
	"strings"
)

// A unique type identifier represented as a string.
// Often times the name of the given type.
type TypeID string

const (
	UNDEFINED TypeID = "UNDEFINED"
	PTR       TypeID = "PTR"
	UINT_LIT  TypeID = "UINT_LIT"
	UINT64    TypeID = "UINT64"
	UINT32    TypeID = "UINT32"
	UINT16    TypeID = "UINT16"
	UINT8     TypeID = "UINT8"
	BOOL      TypeID = "BOOL"
)

// Any type implementing this interface can be used as a type in the compiler.
type Type interface {
	// Either TypeID or in the case of user defined types the type's name.
	TypeID() TypeID
	// Size of the type in bytes.
	Size() int
	// Which part of the rax register the type can be or is stored in.
	Register() string
	// The x86_64 nasm assembly size specifier.
	ASMSize() string
	// Return whether the current type and the other type are equal.
	Equals(other Type) bool
	// Checks whether a given binary operator can be used on the given type and
	// returns the result type after the operation.
	CanUseOperator(op string, operand Type) (bool, Type)
	// Checks whether a given unary operator can be used on the given type and
	// returns the result type after the operation.
	CanUseUnaryOperator(op string) (bool, Type)
}

// This type is used during parsing where the specific type cannot be deduced yet.
type Undefined struct{}

func (Undefined) TypeID() TypeID {
	return UNDEFINED
}

func (Undefined) Size() int {
	return 8
}

func (Undefined) Register() string {
	return "rax"
}

func (Undefined) ASMSize() string {
	return ""
}

func (Undefined) Equals(other Type) bool {
	return false
}

func (Undefined) CanUseOperator(op string, operand Type) (bool, Type) {
	return false, Undefined{}
}

func (Undefined) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// Represents a unsigned integer literal.
type UintLiteral struct{}

func (UintLiteral) TypeID() TypeID {
	return UINT_LIT
}

func (UintLiteral) Size() int {
	return 8
}

func (UintLiteral) Register() string {
	return "rax"
}

func (UintLiteral) ASMSize() string {
	return "QWORD"
}

func (u UintLiteral) Equals(other Type) bool {
	return other.TypeID() == UINT_LIT || other.TypeID() == UINT64
}

func (UintLiteral) CanUseOperator(op string, operand Type) (bool, Type) {
	if !IsNumber(operand) {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/", "=":
		return true, operand
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (UintLiteral) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// Unsigned 64 bit integer.
type Uint64 struct{}

func (Uint64) TypeID() TypeID {
	return UINT64
}

func (Uint64) Size() int {
	return 8
}

func (Uint64) Register() string {
	return "rax"
}

func (Uint64) ASMSize() string {
	return "QWORD"
}

func (u Uint64) Equals(other Type) bool {
	return other.TypeID() == UINT64 || other.TypeID() == UINT_LIT
}

func (Uint64) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT64 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/", "=":
		return true, Uint64{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (Uint64) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ValueType: Uint64{}}
	}

	return false, Undefined{}
}

// Unsigned 64 bit integer.
type Uint32 struct{}

func (Uint32) TypeID() TypeID {
	return UINT32
}

func (Uint32) Size() int {
	return 4
}

func (Uint32) Register() string {
	return "eax"
}

func (Uint32) ASMSize() string {
	return "DWORD"
}

func (u Uint32) Equals(other Type) bool {
	return other.TypeID() == UINT32 || other.TypeID() == UINT_LIT
}

func (Uint32) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT32 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/", "=":
		return true, Uint32{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (Uint32) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ValueType: Uint32{}}
	}

	return false, Undefined{}
}

// Unsigned 16 bit integer.
type Uint16 struct{}

func (Uint16) TypeID() TypeID {
	return UINT16
}

func (Uint16) Size() int {
	return 2
}

func (Uint16) Register() string {
	return "ax"
}

func (Uint16) ASMSize() string {
	return "WORD"
}

func (u Uint16) Equals(other Type) bool {
	return other.TypeID() == UINT16 || other.TypeID() == UINT_LIT
}

func (Uint16) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT16 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/", "=":
		return true, Uint16{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (Uint16) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ValueType: Uint16{}}
	}

	return false, Undefined{}
}

// Unsigned 8 bit integer.
type Uint8 struct{}

func (Uint8) TypeID() TypeID {
	return UINT8
}

func (Uint8) Size() int {
	return 1
}

func (Uint8) Register() string {
	return "al"
}

func (Uint8) ASMSize() string {
	return "BYTE"
}

func (u Uint8) Equals(other Type) bool {
	return other.TypeID() == UINT8 || other.TypeID() == UINT_LIT
}

func (Uint8) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT8 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/", "=":
		return true, Uint8{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (Uint8) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ValueType: Uint8{}}
	}

	return false, Undefined{}
}

// A 1 byte boolean value.
type Bool struct{}

func (Bool) TypeID() TypeID {
	return BOOL
}

func (Bool) Size() int {
	return 1
}

func (Bool) Register() string {
	return "al"
}

func (Bool) ASMSize() string {
	return "BYTE"
}

func (u Bool) Equals(other Type) bool {
	return other.TypeID() == BOOL
}

func (Bool) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != BOOL {
		return false, Undefined{}
	}

	switch op {
	case "==", "<", ">", "<=", ">=", "!=", "=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (Bool) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ValueType: Bool{}}
	}

	return false, Undefined{}
}

// A pointer to some other type of data.
type Ptr struct {
	ValueType Type
}

func (p Ptr) TypeID() TypeID {
	return TypeID(fmt.Sprintf("%v_PTR", p.ValueType.TypeID()))
}

func (Ptr) Size() int {
	return 8
}

func (Ptr) Register() string {
	return "rax"
}

func (Ptr) ASMSize() string {
	return "QWORD"
}

func (p Ptr) Equals(other Type) bool {
	return p.TypeID() == other.TypeID()
}

func (p Ptr) CanUseOperator(op string, operand Type) (bool, Type) {
	if p.TypeID() != operand.TypeID() {
		return false, Undefined{}
	}

	if op == "=" {
		return true, p
	}

	return false, Undefined{}
}

func (p Ptr) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "*" {
		return true, p.ValueType
	}

	if op == "&" {
		return true, Ptr{ValueType: p}
	}

	return false, Undefined{}
}

// An array of elements.
// When referring to arrays we treat them as addresses of their first element.
type Array struct {
	Base   Type
	Length int
}

func (a Array) TypeID() TypeID {
	return TypeID(fmt.Sprintf("%v_ARRAY[%v]", a.Base.TypeID(), a.Length))
}

func (a Array) Size() int {
	return a.Length * a.Base.Size()
}

func (Array) Register() string {
	return "rax"
}

func (Array) ASMSize() string {
	return "QWORD"
}

func (a Array) Equals(other Type) bool {
	arrayType, isArray := other.(Array)
	return (a.TypeID() == other.TypeID()) || (isArray && a.Base.Equals(arrayType.Base))
}

func (a Array) CanUseOperator(op string, operand Type) (bool, Type) {
	if !a.Equals(operand) {
		return false, Undefined{}
	}

	if op == "=" {
		return true, operand
	}

	return false, Undefined{}
}

func (a Array) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// This type represents a function with it's parameters and return type.
type Func struct {
	Params map[string]Type
	Return Type
}

func (f Func) TypeID() TypeID {
	builder := strings.Builder{}

	builder.WriteString("FUNC_")
	for _, param := range f.Params {
		fmt.Fprintf(&builder, "%v_", param.TypeID())
	}
	fmt.Fprintf(&builder, "%v", f.Return.TypeID())

	return TypeID(builder.String())
}

func (Func) Size() int {
	return 8
}

func (Func) Register() string {
	return "rax"
}

func (Func) ASMSize() string {
	return "QWORD"
}

func (f Func) Equals(other Type) bool {
	otherFunc, l := other.(Func)
	if !l {
		return false
	}

	return f.TypeID() == otherFunc.TypeID()
}

func (f Func) CanUseOperator(op string, operand Type) (bool, Type) {
	return false, Undefined{}
}

func (f Func) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// ---------------------------------------------------
//                  HELPER FUNCTIONS
// ---------------------------------------------------

func IsNumber(t Type) bool {
	switch t.TypeID() {
	case UINT64:
		fallthrough
	case UINT32:
		fallthrough
	case UINT16:
		fallthrough
	case UINT8:
		fallthrough
	case UINT_LIT:
		return true
	}

	return false
}
