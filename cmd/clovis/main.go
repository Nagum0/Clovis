package main

import (
	"clovis/lexer"
	"clovis/parser"
	"clovis/semantics"
	"fmt"
	"os"
)

func main() {
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
		for _, err = range parser.Errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	logFile, err := os.Create("log.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer logFile.Close()

	// for _, stmt := range parser.Stmts {
	// 	logFile.WriteString(fmt.Sprintf("%v\n\n", stmt.Print(0)))
	// }

	// -- SEMANTIC ANALYSIS
	semantics := semantics.SemanticChecker{}

	for _, stmt := range parser.Stmts {
		stmt.Semantics(&semantics)
	}
}
