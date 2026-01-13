package vm

import (
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

type Frame struct {
	closure            *object.Closure
	instructionPointer int
	basePointer        int
}

func NewFrame(closure *object.Closure, basePointer int) *Frame {
	return &Frame{
		closure:            closure,
		instructionPointer: -1,
		basePointer:        basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.closure.Fn.Instructions
}
