package vm

import (
	"fmt"

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

		// NOTE: We could one-line the cases here ---
		// but for readability, i guess this is fine ---
		case code.OpExponent:
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpAdd:
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpSubtract:
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpMultiply:
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpDivide:
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

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
			prev := vm.peekStackAddr(0)
			num, ok := prev.(*object.Number)
			if !ok {
				return fmt.Errorf("Cannot negate non-numeric value type '%s'\n", prev.Type())
			}

			vm.stack[vm.stackPointer-1] = &object.Number{Value: -num.Value}

		case code.OpAbsolute:
			prev := vm.peekStackAddr(0)
			num, ok := prev.(*object.Number)
			if !ok {
				return fmt.Errorf("Cannot take the absolute value of a non-numeric value type '%s'\n", prev.Type())
			}

			if num.Value < 0 {
				// NOTE: This is a trick, since calling math.Abs() is expensive/slower
				vm.stack[vm.stackPointer-1] = &object.Number{Value: -num.Value}
			}

		case code.OpNot:
			prev := vm.peekStackAddr(0)
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

		case code.OpSlice:
			end := vm.pop()
			start := vm.pop()
			target := vm.pop()

			if start.Type() == object.NIL_OBJECT {
				start = &object.Number{Value: 0}
			}
			if end.Type() == object.NIL_OBJECT {
				end = &object.Number{Value: float64(len(target.Inspect()) - 2)}
				// NOTE: -2 is for the "" trim
			}

			if start.Type() != object.NUMBER_OBJECT || end.Type() != object.NUMBER_OBJECT {
				return fmt.Errorf("Cannot slice expression with invalid index slicing types ('%s' and '%s')", start.Type(), end.Type())
			}

			endVal := int(end.(*object.Number).Value)
			startVal := int(start.(*object.Number).Value)

			// Check over/under slice
			// NOTE: -2 is for the "" trim
			if endVal > len(target.Inspect())-2 || startVal < 0 {
				return fmt.Errorf("Cannot slice: index out of bounds [%d:%d] (length %d)", startVal, endVal, len(target.Inspect()))
			}

			switch target.Type() {
			case object.STRING_OBJECT:
				targetVal := target.(*object.String).Value
				slicedStr := targetVal[startVal:endVal]
				err := vm.push(&object.String{Value: slicedStr})
				if err != nil {
					return err
				}
			}

		case code.OpArray:
			arrayLength := code.ReadUint16(vm.instructions[instPointer+1:])
			instPointer += 2 // Skip length bytes

			elements := make([]object.Object, arrayLength)
			for i := int(arrayLength) - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}

			err := vm.push(&object.Array{Elements: elements})
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

func (vm *VM) GetStackPointer() int {
	return vm.stackPointer
}

func (vm *VM) peekStackAddr(n int) object.Object {
	return vm.stack[vm.stackPointer-n-1]
}
