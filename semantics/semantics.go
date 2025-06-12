package semantics

import (
	"clovis/lexer"
	"clovis/utils"
	"fmt"
)

type SemanticError struct {
	msg   string
	token lexer.Token
}

func NewSemanticError(msg string, token lexer.Token) *SemanticError {
	return &SemanticError{
		msg: msg,
		token: token,
	}
}

func (s *SemanticError) Error() string {
	return fmt.Sprintf("Semantic error at line %v at col %v\n\t%v", s.token.Line, s.token.Col, s.msg)
}

// Represents a type for semantic analysis.
type Type string
const (
	UNKNOWN = "UNKNOWN"
	UINT = "UINT"
	BOOL = "BOOL"
)

func (t Type) Size() int {
	switch t {
	case BOOL:
		return 1
	case UINT:
		return 8
	}

	return 0
}

// Returns the size specifier of the type in x86_64 asm.
func (t Type) ASMSize() string {
	switch t {
	case BOOL:
		return "BYTE"
	case UINT:
		return "QWORD"
	}

	return ""
}

// Returns which subregister of rax the type's value can be found in.
func (t Type) ASMExprReg() string {
	switch t {
	case BOOL:
		return "al"
	case UINT:
		return "rax"
	}

	return "rax"
}

type Symbol struct {
	Ident  string
	Type   Type
	Offset int
	Size   int
	Token  lexer.Token
}

func (s Symbol) String() string {
	return fmt.Sprintf(
		"ident: %v, type: %v, stack_offset: %v, size: %v",
		s.Ident, s.Type, s.Offset, s.Size,
	)
}

// The SemanticChecker is used to analyze the statements and expressions
// to ensure their correctness.
type SemanticChecker struct {
	Errors			[]error
	symbolTable     utils.Stack[Symbol]
	blockIndexTable utils.Stack[int]
	nextAddr		int
}

func NewSemanticChecker() *SemanticChecker {
	s := SemanticChecker{}
	s.PushBlock()
	return &s
}

func (s SemanticChecker) String() string {
	return fmt.Sprintf(
		"SymbolTable:\n%v\nBlokcIndexTable:\n%v\nnextAddr: %v\n",
		s.symbolTable,
		s.blockIndexTable,
		s.nextAddr,
	)
}

func (s *SemanticChecker) AddError(msg string, token lexer.Token) error {
	err := NewSemanticError(msg, token)
	s.Errors = append(s.Errors, err)
	return err
}

func (s *SemanticChecker) PushSymbol(ident string, symbolType Type, token lexer.Token) error {
	if s.TopBlockHasSymbol(ident) {
		err := NewSemanticError(
			fmt.Sprintf("Redeclaration of symbol '%v'", ident),
			token,
		)
		s.Errors = append(s.Errors, err)
		return err
	}

	symbolSize := symbolType.Size()
	symbol := &Symbol{
		Ident: ident,
		Type: symbolType,
		Token: token,
		Offset: s.nextAddr + symbolSize,
		Size: symbolSize,
	}
	s.nextAddr += symbolSize
	s.symbolTable.Push(*symbol)

	return nil
}

func (s *SemanticChecker) TopSymbol() (Symbol, error) {
	return s.symbolTable.Top()
}


func (s *SemanticChecker) PushBlock() {
	blockStartIndex := s.symbolTable.Size
	s.blockIndexTable.Push(blockStartIndex)
}

func (s *SemanticChecker) PopBlock() {
	if s.blockIndexTable.Size == 0 {
		return
	}

	topBlockIndex, _ := s.blockIndexTable.Pop()
	symbolTableData := s.blockIndexTable.Data()
	for i := len(symbolTableData) - 1; i >= topBlockIndex; i-- {
		s.symbolTable.Pop()
	}
}

func (s *SemanticChecker) GetSymbol(ident lexer.Token) (*Symbol, error) {
	symbolTableData := s.symbolTable.Data()
	for i := len(symbolTableData) - 1; i >= 0; i-- {
		symbol := symbolTableData[i]
		if symbol.Ident == ident.Value {
			return &symbol, nil
		}
	}
	
	err := NewSemanticError(
		fmt.Sprintf("Undeclared symbol '%v'", ident.Value),
		ident,
	)
	s.Errors = append(s.Errors, err)

	return nil, err
}

func (s SemanticChecker) TopBlockHasSymbol(ident string) bool {
	topBlockIndex, err := s.blockIndexTable.Top()
	if err != nil {
		return false
	}
	
	symbolTableData := s.symbolTable.Data()
	for i := len(symbolTableData) - 1; i >= topBlockIndex; i-- {
		symbol := symbolTableData[i]
		if symbol.Ident == ident {
			return true
		}
	}
	
	return false
}

// Returns the semantic type of an operator.
func OperatorType(op lexer.Token) Type {
	switch op.Value {
	case "+":
		fallthrough
	case "-":
		fallthrough
	case "*":
		fallthrough
	case "/":
		return UINT
	case "==":
		fallthrough
	case "<":
		fallthrough
	case "<=":
		fallthrough
	case ">":
		fallthrough
	case ">=":
		return BOOL
	}


	return UNKNOWN
}

func align16(x int) int {
    remainder := x % 16

    if remainder == 0 {
        return x
    }

    return x + (16 - remainder)
}

