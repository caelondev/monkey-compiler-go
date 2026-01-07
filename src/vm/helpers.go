package vm

import "github.com/caelondev/monkey-compiler-go/src/object"

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
