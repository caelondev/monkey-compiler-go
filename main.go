package main

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/lexer"
	"github.com/caelondev/monkey-compiler-go/src/parser"
	"github.com/caelondev/monkey-compiler-go/src/vm"
)

func main() {
	l := lexer.New("1+1")
	p := parser.New(l)
	program := p.ParseProgram()

	c := compiler.New()
	c.Compile(program)

	vm := vm.New(c.Bytecode())

	fmt.Printf("%v", c.Bytecode().Instructions)
	fmt.Printf("%s\n", vm.Run())
}
