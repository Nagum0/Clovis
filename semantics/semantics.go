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

type Symbol struct {
	Ident  string
	Type   Type
	Offset int
	Size   int
	Token  lexer.Token
}

// The SemanticChecker is used to analyze the statements and expressions
// to ensure their correctness.
type SemanticChecker struct {
	Errors			[]error
	symbolTable     utils.Stack[Symbol]
	blockIndexTable utils.Stack[int]
	nextAddr		int
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
		Offset: s.nextAddr,
		Size: symbolSize,
	}
	s.nextAddr += symbolSize
	s.symbolTable.Push(*symbol)

	return nil
}

func (s *SemanticChecker) PushBlock() {
	blockStartIndex := s.symbolTable.Size - 1
	s.blockIndexTable.Push(blockStartIndex)
}

func (s *SemanticChecker) PopBlock() {
	if s.blockIndexTable.Size == 0 {
		return
	}

	topBlockIndex, _ := s.blockIndexTable.Pop()
	symbolTableData := s.blockIndexTable.Data()
	for i := len(symbolTableData) - 1; i > topBlockIndex; i-- {
		s.symbolTable.Pop()
	}
}

func (s SemanticChecker) TopBlockHasSymbol(ident string) bool {
	topBlockIndex, err := s.blockIndexTable.Top()
	if err != nil {
		return false
	}
	
	symbolTableData := s.symbolTable.Data()
	for i := len(symbolTableData) - 1; i > topBlockIndex; i-- {
		symbol := symbolTableData[i]
		if symbol.Ident == ident {
			return true
		}
	}
	
	return false
}
