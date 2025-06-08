package main

import (
	"clovis/codegen"
	"clovis/lexer"
	"clovis/parser"
	"clovis/semantics"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	errOccured := false

	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Expected source file.")
		os.Exit(1)
	}
	
	// -- INPUT
	input, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// -- LEXING
	lexer := lexer.NewLexer(string(input))
	err = lexer.Lex()
	if err != nil {
		for _, err = range lexer.Errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
	
	// -- PARSING
	parser := parser.NewParser(lexer.Tokens)
	err = parser.Parse()
	if err != nil {
		errOccured = true
		for _, err = range parser.Errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	parserLogFile, err := os.Create("plog.txt")
	defer parserLogFile.Close()

	for _, stmt := range parser.Stmts {
		parserLogFile.WriteString(fmt.Sprintf("%v\n\n", stmt.Print(0)))
	}

	// -- SEMANTIC ANALYSIS
	semantics := semantics.NewSemanticChecker()

	semanticsLogFile, err := os.Create("slog.txt")
	defer semanticsLogFile.Close()

	for _, stmt := range parser.Stmts {
		err := stmt.Semantics(semantics)
		if err != nil {
			errOccured = true
			semanticsLogFile.WriteString(err.Error())
			fmt.Println(err.Error())
		} else {
			log := fmt.Sprintf(
				"-----------------------------------------------------\n\n%v\n\n%v\n-----------------------------------------------------\n", 
				stmt.Print(0), 
				semantics,
			)
			semanticsLogFile.WriteString(log)
			fmt.Println(log)
		}
	}

	if errOccured {
		os.Exit(1)
	}

	// -- CODE GENERATION
	emitter := codegen.NewEmitter()
	for _, stmt := range parser.Stmts {
		stmt.EmitCode(emitter)
	}
	emitter.End()

	// -- COMPILE & LINK
	asmFile, err := os.Create("out.asm")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer asmFile.Close()
	asmFile.WriteString(emitter.Code)

	nasmCmd := exec.Command("nasm", "-f", "elf64", "out.asm", "-o", "out.o")
	if err := nasmCmd.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	ldCmd := exec.Command("ld", "out.o", "-o", "out")
	if err := ldCmd.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
