package vm

import (
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/token"
)

func opcodeToOperator(opcode code.OpCode) token.TokenType {
	switch opcode {
	case code.OpAdd:
		return token.PLUS
	case code.OpSubtract:
		return token.MINUS
	case code.OpMultiply:
		return token.STAR
	case code.OpDivide:
		return token.SLASH
	case code.OpExponent:
		return token.CARET

	default:
		return token.ILLEGAL
	}
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
