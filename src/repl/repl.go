package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/lexer"
	"github.com/caelondev/monkey-compiler-go/src/parser"
	"github.com/caelondev/monkey-compiler-go/src/run"
	"github.com/caelondev/monkey-compiler-go/src/vm"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	var allLines []string

	for {
		fmt.Printf(">> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		allLines = append(allLines, line)

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			run.PrintParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compiler::Error: %s\n", err)
			continue
		}

		comp.Disassemble()

		vm := vm.New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			fmt.Fprintf(out, "VM::Error: %s\n", err)
			continue
		}

		stackTop := vm.LastPoppedElement()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")
	}
}
