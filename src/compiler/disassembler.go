package compiler

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func DisassembleBytecode(bytecode *Bytecode) {
	fmt.Println("\n== Disassembler ==")
	DisassembleInstruction(bytecode.Instructions)
	fmt.Println()
	DisassembleConstants(bytecode.Constants)
	fmt.Println()
}

func DisassembleInstruction(instructions code.Instructions) {
	i := 0
	for i < len(instructions) {
		op := code.OpCode(instructions[i])
		def, err := code.Lookup(op)
		if err != nil {
			fmt.Printf("Unknown opcode %d\n", op)
			i++
			continue
		}
		fmt.Printf("%04d %s", i, def.Name)
		i++
		operands := make([]int, len(def.OperandWidths))
		for j, width := range def.OperandWidths {
			switch width {
			case 2:
				operands[j] = int(code.ReadUint16(instructions[i:]))
				i += 2
			default:
				panic("Unsupported operand width")
			}
		}
		if len(operands) > 0 {
			fmt.Printf(" %v", operands)
		}

		fmt.Println()
	}

}

func DisassembleConstants(constants []object.Object) {
	fmt.Println("Constants:")

	if len(constants) == 0 {
		fmt.Printf("No constants found")
		return
	}

	for idx, cn := range constants {
		switch v := cn.(type) {
		case *object.Number:
			fmt.Printf("%d: %g\n", idx, v.Value)
		case *object.String:
			fmt.Printf("%d: \"%s\"\n", idx, v.Value)
		case *object.CompiledFunction:
			fmt.Println()
			fmt.Printf("%d: --- Start Compile Function ---\n", idx)
			DisassembleInstruction(v.Instructions)
			fmt.Printf("%d: --- End Compile Function ---", idx)
			fmt.Println()

		default:
			fmt.Printf("%d: unknown constant type %T\n", idx, cn)
		}
	}
}
