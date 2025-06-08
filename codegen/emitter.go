package codegen

import "strings"

// Generates x86_64 assembly code.
type Emitter struct {
	Code string
}

func NewEmitter() *Emitter {
	b := strings.Builder{}
	b.WriteString("section .text\n")
	b.WriteString("global _start\n\n")
	b.WriteString("_start:\n")
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
	b := strings.Builder{}
	b.WriteString("\n; Emitter.End()\n")
	b.WriteString("mov rax, 60\n")
	b.WriteString("mov rdi, 0\n")
	b.WriteString("syscall\n")
	e.Code += b.String()
}
