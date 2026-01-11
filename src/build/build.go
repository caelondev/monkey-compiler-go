package build

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/lexer"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/parser"
	"github.com/caelondev/monkey-compiler-go/src/vm"
)

const MAGIC = "MCGO"
const VERSION = 1

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

	// Convert bytecode to raw bytes
	encodedBytes := EncodeBytecode(bytecode.Constants, bytecode.Instructions)

	outputPath := FormatFileName(path)
	WriteByteToFile(outputPath, encodedBytes)

	fmt.Println("Build successful")
}

func EncodeBytecode(constants []object.Object, instructions []byte) []byte {
	buf := new(bytes.Buffer)

	buf.Write([]byte(MAGIC))
	buf.WriteByte(VERSION)
	buf.Write(serializeConstants(constants))
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

	compiler.DisassembleBytecode(bytecode)

	return nil
}
