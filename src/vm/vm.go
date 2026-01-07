package vm

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack        []object.Object
	stackPointer int
}

const STACK_SIZE = 2048

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:        make([]object.Object, STACK_SIZE),
		stackPointer: 0,
	}
}

func (vm *VM) Run() error {
	for instPointer := 0; instPointer < len(vm.instructions); instPointer++ {
		op := code.OpCode(vm.instructions[instPointer])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[instPointer+1:])
			instPointer += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: leftVal + rightVal}
			vm.push(total)

			fmt.Print(total.Value)
		}
	}

	return nil
}

func (vm *VM) StackTop() object.Object {
	if vm.stackPointer == 0 {
		return nil
	}

	return vm.stack[vm.stackPointer-1]
}

func (vm *VM) push(obj object.Object) error {
	if vm.stackPointer >= STACK_SIZE {
		return fmt.Errorf("Stack overflow")
	}

	vm.stack[vm.stackPointer] = obj
	vm.stackPointer++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.stackPointer-1]
	vm.stackPointer--
	return obj
}
