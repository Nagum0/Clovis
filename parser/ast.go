package parser

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/utils"
)

type Type string
const (
	UINT = "UINT"
	BOOL = "BOOL"
)

type Statement interface {
	EmitCode(e *codegen.Emitter) error
}

type VarDeclStmt struct {
	VarType lexer.TokenType
	Ident	lexer.Token
	Value	utils.Optional[Expression]
}

func (s *VarDeclStmt) EmitCode(e *codegen.Emitter) error {
	return nil
}

type Expression interface {
	EmitCode(e *codegen.Emitter) error
}
