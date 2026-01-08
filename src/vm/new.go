package vm

import (
	"os"

	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:        make([]object.Object, STACK_SIZE),
		globals:      make([]object.Object, GLOBAL_SIZE),
		stackPointer: 0,
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, global []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = global
	return vm
}

func NewFromFile(path string) (*VM, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bytecode, err := DecodeBytecode(data)
	if err != nil {
		return nil, err
	}

	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, STACK_SIZE),
		globals:      make([]object.Object, GLOBAL_SIZE),
		stackPointer: 0,
	}, nil
}

