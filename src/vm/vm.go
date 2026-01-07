package vm

import (
	"fmt"
	"math"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

const STACK_SIZE = 2048

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack        []object.Object
	stackPointer int
}

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

		case code.OpExponent:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: math.Pow(leftVal, rightVal)}
			vm.push(total)
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: leftVal + rightVal}
			vm.push(total)
		case code.OpSubtract:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: leftVal - rightVal}
			vm.push(total)
		case code.OpMultiply:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: leftVal * rightVal}
			vm.push(total)
		case code.OpDivide:
			right := vm.pop()
			left := vm.pop()
			rightVal := right.(*object.Number).Value
			leftVal := left.(*object.Number).Value

			total := &object.Number{Value: leftVal / rightVal}
			vm.push(total)

		case code.OpEqual, code.OpNotEqual, code.OpLess, code.OpLessEqual, code.OpGreater, code.OpGreaterEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpTrue:
			err := vm.push(object.TRUE)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(object.FALSE)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()
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

func (vm *VM) LastPoppedElement() object.Object {
	return vm.stack[vm.stackPointer]
}
