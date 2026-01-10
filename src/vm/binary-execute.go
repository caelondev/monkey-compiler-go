package vm

import (
	"fmt"
	"math"
	"strings"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func (vm *VM) executeBinop(opcode code.OpCode) error {
	r := vm.peekStackAddr(0)
	l := vm.peekStackAddr(1)

	switch {
	case l.Type() == object.NUMBER_OBJECT && r.Type() == object.NUMBER_OBJECT:
		right := vm.pop()
		left := vm.pop()
		return vm.executeNumericBinop(left, right, opcode)

	case l.Type() == object.STRING_OBJECT && r.Type() == object.STRING_OBJECT:
		right := vm.pop().(*object.String)
		left := vm.pop().(*object.String)

		// concatenate
		if opcode == code.OpAdd {
			return vm.push(&object.String{Value: left.Value + right.Value})
		}

	case l.Type() == object.STRING_OBJECT && r.Type() == object.NUMBER_OBJECT:
		right, _ := vm.pop().(*object.Number)
		left, _ := vm.pop().(*object.String)

		// repeat
		if opcode == code.OpMultiply {
			return vm.push(&object.String{Value: strings.Repeat(left.Value, int(right.Value))})
		}

	default:
		return fmt.Errorf("Cannot perform binary operation to unsupported type (%s and %s)", l.Type(), r.Type())
	}

	return fmt.Errorf("Invalid operator type '%s' for operand type %s and %s", opcodeToOperator(opcode), l.Type(), r.Type())
}

func (vm *VM) executeNumericBinop(left, right object.Object, opcode code.OpCode) error {
	l := left.(*object.Number).Value
	r := right.(*object.Number).Value

	var err error

	switch opcode {
	case code.OpAdd:
		err = vm.push(&object.Number{Value: l + r})
	case code.OpSubtract:
		err = vm.push(&object.Number{Value: l - r})
	case code.OpMultiply:
		err = vm.push(&object.Number{Value: l * r})
	case code.OpDivide:
		err = vm.push(&object.Number{Value: l / r})
	case code.OpExponent:
		err = vm.push(&object.Number{Value: math.Pow(l, r)})
	}

	if err != nil {
		return err
	}
	return nil
}
