package vm

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func (vm *VM) executeComparison(op code.OpCode) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() == object.NUMBER_OBJECT && left.Type() == object.NUMBER_OBJECT {
		return vm.executeNumberComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right))
	}

	return fmt.Errorf("Unknown comparison operator: '%d'\n", op)
}

func (vm *VM) executeNumberComparison(op code.OpCode, left, right object.Object) error {
	l := left.(*object.Number).Value
	r := right.(*object.Number).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(l == r))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(l != r))
	case code.OpLess:
		return vm.push(nativeBoolToBooleanObject(l < r))
	case code.OpGreater:
		return vm.push(nativeBoolToBooleanObject(l > r))
	case code.OpLessEqual:
		return vm.push(nativeBoolToBooleanObject(l <= r))
	case code.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(l >= r))
	}

	return fmt.Errorf("Unknown comparison operator: '%d'\n", op)
}

func nativeBoolToBooleanObject(b bool) *object.Boolean {
	if b {
		return object.TRUE
	}

	return object.FALSE
}
