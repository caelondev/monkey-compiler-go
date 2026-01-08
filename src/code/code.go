package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte
type OpCode byte

const (
	OpConstant OpCode = iota
	OpTrue
	OpFalse
	OpNil

	OpAdd
	OpSubtract
	OpMultiply
	OpDivide
	OpExponent

	OpEqual
	OpNotEqual
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual

	OpNegate
	OpNot

	OpJump
	OpJumpNotTruthy

	OpPop
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpNil:           {"OpNil", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpNegate:        {"OpNegate", []int{}},
	OpNot:           {"OpNot", []int{}},
	OpGreater:       {"OpGreater", []int{}},
	OpGreaterEqual:  {"OpGreaterEqual", []int{}},
	OpLess:          {"OpLess", []int{}},
	OpLessEqual:     {"OpLessEqual", []int{}},
	OpSubtract:      {"OpSubtract", []int{}},
	OpMultiply:      {"OpMultiply", []int{}},
	OpDivide:        {"OpDivide", []int{}},
	OpExponent:      {"OpExponent", []int{}},
	OpPop:           {"OpPop", []int{}},
}

func Lookup(opcode OpCode) (*Definition, error) {
	def, ok := definitions[opcode]
	if !ok {
		return nil, fmt.Errorf("Unknown OpCode %d", opcode)
	}

	return def, nil
}

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0

	for i < len(ins) {
		def, err := Lookup(OpCode(ins[i]))
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(operands)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return fmt.Sprintf("%s\n", def.Name)
	case 1:
		return fmt.Sprintf("%s %d\n", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func Make(opcode OpCode, operands ...int) []byte {
	def, ok := definitions[opcode]
	if !ok {
		return []byte{}
	}

	instructionLength := 1

	for _, width := range def.OperandWidths {
		instructionLength += width
	}

	instruction := make([]byte, instructionLength) // Pre-allocate
	instruction[0] = byte(opcode)

	offset := 1

	for i, operand := range operands {
		width := def.OperandWidths[i]

		// Convert current width to byte
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(operand))
		}

		// Advance offset based on current width
		offset += width
	}

	return instruction
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
