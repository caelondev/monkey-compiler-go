package build

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/lexer"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/parser"
	"github.com/caelondev/monkey-compiler-go/src/vm"
)

func BuildFile(path string) {
	input, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	l := lexer.New(string(input))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		return
	}

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		panic(err)
	}

	bytecode := comp.Bytecode()

	encodedBytes := EncodeBytecode(bytecode.Constants, bytecode.Instructions)

	// Use FormatFileName to get the output path
	outputPath := FormatFileName(path)
	WriteByteToFile(outputPath, encodedBytes)

	fmt.Println("Build successful")
}

func EncodeBytecode(constants []object.Object, instructions []byte) []byte {
	buf := new(bytes.Buffer)

	// magic
	buf.Write([]byte("MCGO"))

	// version
	buf.WriteByte(1)

	// constants
	buf.Write(serializeConstants(constants))

	// instructions
	buf.Write(serializeInstructions(instructions))

	return buf.Bytes()
}

func WriteByteToFile(filename string, bytecode []byte) error {
	// 0644 is a unix permission value
	return os.WriteFile(filename, bytecode, 0644)
}

func FormatFileName(filename string) string {
	name := filepath.Base(filename)
	parts := strings.Split(name, ".")

	parts = parts[:len(parts)-1]

	joined := strings.Join(parts, ".")
	joined += ".mnc"
	return joined
}

func DisassembleFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	bytecode, err := vm.DecodeBytecode(data)
	if err != nil {
		return err
	}

	fmt.Println("== Disassembler ==")
	instructions := bytecode.Instructions
	i := 0
	for i < len(instructions) {
		op := instructions[i]
		def, err := code.Lookup(code.OpCode(op))
		if err != nil {
			fmt.Printf("%04d UNKNOWN OPCODE %d\n", i, op)
			i++
			continue
		}

		fmt.Printf("%04d %s", i, def.Name)
		i++

		// read operands
		for _, width := range def.OperandWidths {
			if width == 2 {
				val := int(instructions[i])<<8 | int(instructions[i+1])
				fmt.Printf(" [%d]", val)
				i += 2
			} else {
				fmt.Printf(" [unsupported operand width %d]", width)
			}
		}
		fmt.Println()
	}

	// print constants
	fmt.Println("\nConstants:")
	for idx, c := range bytecode.Constants {
		switch v := c.(type) {
		case *object.Number:
			fmt.Printf("%d: %f\n", idx, v.Value)
		case *object.Boolean:
			fmt.Printf("%d: %v\n", idx, v.Value)
		case *object.Nil:
			fmt.Printf("%d: nil\n", idx)
		default:
			fmt.Printf("%d: unknown constant type %T\n", idx, c)
		}
	}

	return nil
}
