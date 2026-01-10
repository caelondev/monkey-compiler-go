package vm

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/object"
)

func (vm *VM) executeNumericalBinop(fn func(left, right *object.Number)) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() != object.NUMBER_OBJECT || left.Type() != object.NUMBER_OBJECT {
		return fmt.Errorf("Cannot perform binary operation to unsupported type (%s and %s)", left.Type(), right.Type())
	}

	leftObj := left.(*object.Number)
	rightObj := right.(*object.Number)

	fn(leftObj, rightObj)
	return nil
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Nil:
		return false

	case *object.Boolean:
		return obj.Value

	case *object.Number:
		return obj.Value != 0

	case *object.NaN:
		return false

	default:
		return true
	}
}
