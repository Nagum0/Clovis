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
		s.Ident, s.Type.TypeID(), s.Offset, s.Size,
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
	s.blockIndexTable.Push(0) // global scope currently
	return &s
}

func (s SemanticChecker) String() string {
	return fmt.Sprintf(
		"SymbolTable:\n%v\nBlockIndexTable:\n%v\nnextAddr: %v\n",
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
		return s.AddError(
			fmt.Sprintf("Redeclaration of symbol '%v'", ident),
			token,
		)
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

// Pops a block off the symbol table.
// Returns the size of the popped block.
func (s *SemanticChecker) PopBlock() int {
	if s.blockIndexTable.Size == 0 {
		return 0
	}
	
	size := 0
	topBlockIndex, _ := s.blockIndexTable.Pop()
	symbolTableData := s.symbolTable.Data()
	for i := len(symbolTableData) - 1; i >= topBlockIndex; i-- {
		symbol, _ := s.symbolTable.Pop()
		size += symbol.Size
	}

	return size
}

func (s *SemanticChecker) GetSymbol(ident lexer.Token) (*Symbol, error) {
	symbolTableData := s.symbolTable.Data()
	for i := len(symbolTableData) - 1; i >= 0; i-- {
		symbol := symbolTableData[i]
		if symbol.Ident == ident.Value {
			return &symbol, nil
		}
	}

	return nil, s.AddError(
		fmt.Sprintf("Undeclared symbol '%v'", ident.Value),
		ident,
	)
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

func align16(x int) int {
    remainder := x % 16

    if remainder == 0 {
        return x
    }

    return x + (16 - remainder)
}
