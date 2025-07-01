package semantics

import "fmt"

// A unique type identifier represented as a string.
// Often times the name of the given type.
type TypeID string
const (
	UNDEFINED TypeID = "UNDEFINED"
	PTR TypeID = "PTR"
	UINT_LIT TypeID = "UINT_LIT"
	UINT64 TypeID = "UINT64"
	UINT32 TypeID = "UINT32"
	UINT16 TypeID = "UINT16"
	UINT8 TypeID = "UINT8"
	BOOL TypeID = "BOOL"
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
type Undefined struct {}

func (_ Undefined) TypeID() TypeID {
	return UNDEFINED
}

func (_ Undefined) Size() int {
	return 8
}

func (_ Undefined) Register() string {
	return "rax"
}

func (_ Undefined) ASMSize() string {
	return ""
}

func (_ Undefined) Equals(other Type) bool {
	return false
}

func (_ Undefined) CanUseOperator(op string, operand Type) (bool, Type) {
	return false, Undefined{}
}

func (_ Undefined) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// Represents a unsigned integer literal.
type UintLiteral struct {}

func (_ UintLiteral) TypeID() TypeID {
	return UINT_LIT
}

func (_ UintLiteral) Size() int {
	return 8
}

func (_ UintLiteral) Register() string {
	return "rax"
}

func (_ UintLiteral) ASMSize() string {
	return "QWORD"
}

func (u UintLiteral) Equals(other Type) bool {
	return other.TypeID() == UINT_LIT || other.TypeID() == UINT64
}

func (_ UintLiteral) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ UintLiteral) CanUseUnaryOperator(op string) (bool, Type) {
	return false, Undefined{}
}

// Unsigned 64 bit integer.
type Uint64 struct {}

func (_ Uint64) TypeID() TypeID {
	return UINT64
}

func (_ Uint64) Size() int {
	return 8
}

func (_ Uint64) Register() string {
	return "rax"
}

func (_ Uint64) ASMSize() string {
	return "QWORD"
}

func (u Uint64) Equals(other Type) bool {
	return other.TypeID() == UINT64 || other.TypeID() == UINT_LIT
}

func (_ Uint64) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ Uint64) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ ValueType: Uint64{} }
	}

	return false, Undefined{}
}

// Unsigned 64 bit integer.
type Uint32 struct {}

func (_ Uint32) TypeID() TypeID {
	return UINT32
}

func (_ Uint32) Size() int {
	return 4
}

func (_ Uint32) Register() string {
	return "eax"
}

func (_ Uint32) ASMSize() string {
	return "DWORD"
}

func (u Uint32) Equals(other Type) bool {
	return other.TypeID() == UINT32 || other.TypeID() == UINT_LIT
}

func (_ Uint32) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ Uint32) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ ValueType: Uint32{} }
	}

	return false, Undefined{}
}

// Unsigned 16 bit integer.
type Uint16 struct {}

func (_ Uint16) TypeID() TypeID {
	return UINT16
}

func (_ Uint16) Size() int {
	return 2
}

func (_ Uint16) Register() string {
	return "ax"
}

func (_ Uint16) ASMSize() string {
	return "WORD"
}

func (u Uint16) Equals(other Type) bool {
	return other.TypeID() == UINT16 || other.TypeID() == UINT_LIT
}

func (_ Uint16) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ Uint16) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ ValueType: Uint16{} }
	}

	return false, Undefined{}
}

// Unsigned 8 bit integer.
type Uint8 struct {}

func (_ Uint8) TypeID() TypeID {
	return UINT8
}

func (_ Uint8) Size() int {
	return 1
}

func (_ Uint8) Register() string {
	return "al"
}

func (_ Uint8) ASMSize() string {
	return "BYTE"
}

func (u Uint8) Equals(other Type) bool {
	return other.TypeID() == UINT8 || other.TypeID() == UINT_LIT
}

func (_ Uint8) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ Uint8) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ ValueType: Uint8{} }
	}

	return false, Undefined{}
}

// A 1 byte boolean value.
type Bool struct {}

func (_ Bool) TypeID() TypeID {
	return BOOL
}

func (_ Bool) Size() int {
	return 1
}

func (_ Bool) Register() string {
	return "al"
}

func (_ Bool) ASMSize() string {
	return "BYTE"
}

func (u Bool) Equals(other Type) bool {
	return other.TypeID() == BOOL
}

func (_ Bool) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != BOOL {
		return false, Undefined{}
	}

	switch op {
	case "==", "<", ">", "<=", ">=", "!=", "=":
		return true, Bool{}
	}

	return false, Undefined{}
}

func (_ Bool) CanUseUnaryOperator(op string) (bool, Type) {
	if op == "&" {
		return true, Ptr{ ValueType: Bool{} }
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

func (_ Ptr) Size() int {
	return 8
}

func (_ Ptr) Register() string {
	return "rax"
}

func (_ Ptr) ASMSize() string {
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
		return true, Ptr{ ValueType: p }
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
	return TypeID(fmt.Sprintf("%v_ARRAY(%v)", a.Base.TypeID(), a.Length))
}

func (a Array) Size() int {
	return a.Length * a.Base.Size()
}

func (_ Array) Register() string {
	return "rax"
}

func (_ Array) ASMSize() string {
	return "QWORD"
}

func (a Array) Equals(other Type) bool {
	arrayType, isArray := other.(Array)
	return (a.TypeID() == other.TypeID()) || (isArray && a.Base.Equals(arrayType.Base))
}

func (a Array) CanUseOperator(op string, operand Type) (bool, Type) {
	return false, Undefined{}
}

func (a Array) CanUseUnaryOperator(op string) (bool, Type) {
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
