package semantics

// A unique type identifier represented as a string.
// Often times the name of the given type.
type TypeID string
const (
	UNDEFINED TypeID = "UNDEFINED"
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
	// Checks whether a given operator can be used on the given type and 
	// returns the result type after the operation.
	CanUseOperator(op string, operand Type) (bool, Type)
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

func (_ Undefined) CanUseOperator(op string, operand Type) (bool, Type) {
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

func (_ UintLiteral) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT64 {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/":
		return true, UintLiteral{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

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

func (_ Uint64) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT64 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/":
		return true, Uint64{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
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

func (_ Uint32) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT32 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/":
		return true, Uint32{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
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

func (_ Uint16) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT16 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/":
		return true, Uint16{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
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

func (_ Uint8) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != UINT8 && operand.TypeID() != UINT_LIT {
		return false, Undefined{}
	}

	switch op {
	case "+", "-", "*", "/":
		return true, Uint8{}
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}

// A 1 byte boolean value.
type Bool struct {}

func (_ Bool) Name() string {
	return "bool"
}

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

func (_ Bool) CanUseOperator(op string, operand Type) (bool, Type) {
	if operand.TypeID() != BOOL {
		return false, Undefined{}
	}

	switch op {
	case "==", "<", ">", "<=", ">=", "!=":
		return true, Bool{}
	}

	return false, Undefined{}
}
