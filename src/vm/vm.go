package vm

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

const STACK_SIZE = 10000
const GLOBAL_SIZE = 65536
const FRAME_SIZE = 10000 // max call depth ---

type VM struct {
	constants []object.Object

	frames       []*Frame
	stack        []object.Object
	globals      []object.Object
	stackPointer int
	frameIndex   int
}

func (vm *VM) Run() error {
	var instPointer int
	var inst code.Instructions
	var op code.OpCode

	for vm.currentFrame().instructionPointer < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().instructionPointer++

		instPointer = vm.currentFrame().instructionPointer
		inst = vm.currentFrame().Instructions()
		op = code.OpCode(inst[instPointer])

		switch op {
		case code.OpConstant:
			// OpConstant <constIndex:2>
			constIndex := code.ReadUint16(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		// NOTE: We could one-line the cases here ---
		// but for readability, i guess this is fine ---
		case code.OpExponent:
			// OpExponent
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpAdd:
			// OpAdd
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpSubtract:
			// OpSubtract
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpMultiply:
			// OpMultiply
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpDivide:
			// OpDivide
			err := vm.executeBinop(op)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpLess, code.OpLessEqual, code.OpGreater, code.OpGreaterEqual:
			// Op[code]
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpTrue:
			// OpTrue
			err := vm.push(object.TRUE)
			if err != nil {
				return err
			}

		case code.OpFalse:
			// OpFalse
			err := vm.push(object.FALSE)
			if err != nil {
				return err
			}

		case code.OpNegate:
			// OpNegate
			prev := vm.peekStackAddr(0)
			num, ok := prev.(*object.Number)
			if !ok {
				return fmt.Errorf("Cannot negate non-numeric value type '%s'\n", prev.Type())
			}

			vm.stack[vm.stackPointer-1] = &object.Number{Value: -num.Value}

		case code.OpAbsolute:
			// OpAbsolute
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
			// OpNot
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
			// OpJump <jumpPos:2>
			pos := int(code.ReadUint16(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer = pos - 1
		case code.OpJumpNotTruthy:
			// OpJumpNotTruthy <jumpPos:2>
			pos := int(code.ReadUint16(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().instructionPointer = pos - 1
			}

		case code.OpNil:
			// OpNil
			err := vm.push(object.NIL)
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			// OpSetGlobal <globalIndex:2>
			globalIndex := code.ReadUint16(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 2

			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			// OpGetGlobal <globalIndex:2>
			globalIndex := code.ReadUint16(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 2

			value := vm.globals[globalIndex]
			if value == nil {
				return fmt.Errorf("Cannot access uninitialized global identifier")
			}

			err := vm.push(value)
			if err != nil {
				return err
			}

		case code.OpGetFree:
			// OpGetFree <freeIndex:1>
			freeIndex := int(code.ReadUint8(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer += 1

			cl := vm.currentFrame().closure
			err := vm.push(cl.Free[freeIndex])
			if err != nil {
				return err
			}

		case code.OpGetNative:
			// OpGetNative <nativeIndex:2>
			nativeIndex := code.ReadUint8(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 1

			definition := object.NativeFunctions[nativeIndex]

			err := vm.push(definition.NativeFn)
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			// OpSetLocal <localIndex:2>
			localIndex := code.ReadUint16(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 2

			// Set local index relative to the
			// base pointer
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()

		case code.OpGetLocal:
			// OpGetLocal <localIndex:2>
			localIndex := code.ReadUint16(inst[instPointer+1:])
			vm.currentFrame().instructionPointer += 2

			frame := vm.currentFrame()
			value := vm.stack[frame.basePointer+int(localIndex)]

			if value == nil {
				return fmt.Errorf("Cannot access uninitialized local identifier")
			}

			err := vm.push(value)
			if err != nil {
				return err
			}

		case code.OpSetIndex:
			// OpSetIndex
			newVal := vm.pop()
			index := vm.pop()
			target := vm.pop()

			err := vm.executeIndexAssignment(target, index, newVal)
			if err != nil {
				return err
			}

		case code.OpSlice:
			// OpSlice
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
			// OpArray <arrayLength:2>
			arrayLength := int(code.ReadUint16(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer += 2 // Advance past array length

			elements := make([]object.Object, arrayLength)

			sp := vm.stackPointer
			for i := 0; i < arrayLength; i++ {
				elements[i] = vm.stack[sp-arrayLength+i]
			}

			// adjust stack pointer
			vm.drop(arrayLength)

			err := vm.push(&object.Array{Elements: elements})
			if err != nil {
				return err
			}

		case code.OpHash:
			// OpSize <hashSize:2>
			hashSize := int(code.ReadUint16(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer += 2 // Advance past hash size

			hashPairs := make(map[object.HashKey]object.HashPair)

			for i := vm.stackPointer - hashSize; i < vm.stackPointer; i += 2 {
				key := vm.stack[i]
				value := vm.stack[i+1]

				pair := object.HashPair{Key: key, Value: value}
				hashKey, hashable := key.(object.Hashable)
				if !hashable {
					return fmt.Errorf("Cannot use key type of '%s' in a hash map", key.Type())
				}

				hashPairs[hashKey.HashKey()] = pair
			}

			vm.drop(hashSize)
			vm.push(&object.Hash{Pairs: hashPairs})

		case code.OpIndex:
			// OpIndex
			index := vm.pop()
			target := vm.pop()

			err := vm.executeIndexExpression(target, index)
			if err != nil {
				return err
			}

		case code.OpCurrentClosure:
			closure := vm.currentFrame().closure
			err := vm.push(closure)
			if err != nil {
				return err
			}

		case code.OpClosure:
			// OpClosure <functionIndex:2> <numFree:1>
			functionIndex := int(code.ReadUint16(inst[instPointer+1:]))
			numFree := int(code.ReadUint8(inst[instPointer+3:]))
			vm.currentFrame().instructionPointer += 3

			err := vm.pushClosure(functionIndex, numFree)
			if err != nil {
				return err
			}

		case code.OpCall:
			// OpCall <numArgs:2>
			numArgs := int(code.ReadUint8(inst[instPointer+1:]))
			vm.currentFrame().instructionPointer += 1

			err := vm.executeCallExpression(numArgs)
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			// OpReturnValue
			returnValue := vm.pop()

			// This exits the function frame
			frame := vm.popFrame()

			// Reset stack pointer
			vm.stackPointer = frame.basePointer - 1 // -1 for function clearing
			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.drop(1)
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

func (vm *VM) drop(n int) {
	vm.stackPointer -= n
}

func (vm *VM) LastPoppedElement() object.Object {
	return vm.stack[vm.stackPointer]
}

func (vm *VM) GetStack() []object.Object {
	return vm.stack
}

func (vm *VM) GetStackPointer() int {
	return vm.stackPointer
}

func (vm *VM) peekStackAddr(n int) object.Object {
	return vm.stack[vm.stackPointer-n-1]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(frame *Frame) error {
	vm.frames[vm.frameIndex] = frame
	vm.frameIndex++
	return nil
}

func (vm *VM) pushClosure(functionIndex, numFree int) error {
	constant := vm.constants[functionIndex]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("Cannot push constant type '%s' into the stack as a Closure", constant.Type())
	}

	free := make([]object.Object, numFree)

	for i := range free {
		free[i] = vm.stack[vm.stackPointer-numFree+i]
	}

	vm.drop(numFree)

	closure := &object.Closure{Fn: fn, Free: free}
	return vm.push(closure)
}

func (vm *VM) popFrame() *Frame {
	frame := vm.frames[vm.frameIndex-1]
	vm.frameIndex--
	return frame
}
