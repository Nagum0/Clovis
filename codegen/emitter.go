package codegen

import (
	"clovis/lexer"
	"fmt"
	"strings"
)

// Generates x86_64 assembly code.
type Emitter struct {
	Code       string
	LabelCount int
}

func NewEmitter() *Emitter {
	b := strings.Builder{}

	b.WriteString("section .text\n")
	b.WriteString("global _start\n")
	b.WriteString("global main\n\n")

	return &Emitter{
		Code: b.String(),
	}
}

func (e *Emitter) Write(p []byte) (n int, err error) {
	e.Code += string(p)
	return len(p), nil
}

func (e *Emitter) WriteString(code string) {
	e.Code += code
}

// Adds a exit syscall to the end of the code.
func (e *Emitter) End() {
	fmt.Fprintf(e, "\n; Emitter.End()\n")
	fmt.Fprintf(e, "_start:\n")
	fmt.Fprintf(e, "mov rbp, rsp\n")
	fmt.Fprintf(e, "sub rsp, 8\n")
	fmt.Fprintf(e, "call main\n")
	fmt.Fprintf(e, "add rsp, 8\n")
	fmt.Fprintf(e, "mov rdi, rax\n")
	fmt.Fprintf(e, "mov rax, 60\n")
	fmt.Fprintf(e, "syscall\n")
}

func ASMBinaryOp(op lexer.Token) string {
	switch op.Value {
	case "+":
		return "add"
	case "-":
		return "sub"
	case "*":
		return "mul"
	case "/":
		return "div"
	case "==":
		return "sete"
	case "!=":
		return "setne"
	case "<":
		return "setl"
	case "<=":
		return "setle"
	case ">":
		return "setg"
	case ">=":
		return "setge"
	}

	return ""
}

func (e *Emitter) NextLabel() string {
	e.LabelCount++
	return fmt.Sprintf(".L%02v", e.LabelCount)
}
