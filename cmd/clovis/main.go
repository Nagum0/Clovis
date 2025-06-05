package main

import (
	"clovis/lexer"
	"clovis/parser"
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Expected source file.")
		os.Exit(1)
	}
	
	input, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	lexer := lexer.NewLexer(string(input))
	err = lexer.Lex()
	if err != nil {
		for _, err = range lexer.Errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
	
	parser := parser.NewParser(lexer.Tokens)
	err = parser.Parse()
	if err != nil {
		for _, err = range parser.Errors {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	fmt.Printf("%v %v\n", len(parser.Stmts), len(parser.Errors))

	for _, stmt := range parser.Stmts {
		fmt.Println(stmt)
	}
}
