package build

import (
	"bytes"
	"encoding/binary"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func writeUint32(buf *bytes.Buffer, v uint32) {
	_ = binary.Write(buf, binary.BigEndian, v)
}

func writeFloat64(buf *bytes.Buffer, v float64) {
	_ = binary.Write(buf, binary.BigEndian, v)
}

func writeString(buf *bytes.Buffer, v string) {
	writeUint32(buf, uint32(len(v)))
	buf.WriteString(v)
}

// Serialize constants
func serializeConstants(constants []object.Object) []byte {
	buf := new(bytes.Buffer)

	writeUint32(buf, uint32(len(constants))) // number of constants

	for _, c := range constants {
		switch obj := c.(type) {
		case *object.Number:
			buf.WriteByte(byte(code.CONSTANT_NUMBER))
			writeFloat64(buf, obj.Value)
		case *object.String:
			buf.WriteByte(byte(code.CONSTANT_STRING))
			writeString(buf, obj.Value)

		default:
			panic("unsupported constant type")
		}
	}

	return buf.Bytes()
}

func serializeInstructions(instructions []byte) []byte {
	buf := new(bytes.Buffer)
	writeUint32(buf, uint32(len(instructions)))
	buf.Write(instructions)
	return buf.Bytes()
}
