package vm

import (
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

type Frame struct {
	fn                 *object.CompiledFunction
	instructionPointer int
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{
		fn:                 fn,
		instructionPointer: -1,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
