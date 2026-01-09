package vm

import (
	"fmt"
	"math"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

const STACK_SIZE = 2048
const GLOBAL_SIZE = 65536

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack        []object.Object
	globals      []object.Object
	stackPointer int
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

		case code.OpNegate:
			prev := vm.stack[vm.stackPointer-1]
			num, ok := prev.(*object.Number)
			if !ok {
				return fmt.Errorf("Cannot negate non-numeric value type '%s'\n", prev.Type())
			}

			vm.stack[vm.stackPointer-1] = &object.Number{Value: -num.Value}

		case code.OpAbsolute:
			prev := vm.stack[vm.stackPointer-1]
			num, ok := prev.(*object.Number)
			if !ok {
				return fmt.Errorf("Cannot take the absolute value of a non-numeric value type '%s'\n", prev.Type())
			}

			// Avoid allocation
			if num.Value < 0 {
				// NOTE: This is a trick, since calling math.Abs() is expensive
				vm.stack[vm.stackPointer-1] = &object.Number{Value: -num.Value}
			}

		case code.OpNot:
			prev := vm.stack[vm.stackPointer-1]
			var boolObj *object.Boolean

			// Flip value
			if isTruthy(prev) {
				boolObj = object.FALSE
			} else {
				boolObj = object.TRUE
			}

			vm.stack[vm.stackPointer-1] = boolObj

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[instPointer+1:]))
			instPointer = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[instPointer+1:]))
			instPointer += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				instPointer = pos - 1
			}

		case code.OpNil:
			err := vm.push(object.NIL)
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[instPointer+1:])
			instPointer += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[instPointer+1:])
			instPointer += 2

			err := vm.push(vm.globals[globalIndex])
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
