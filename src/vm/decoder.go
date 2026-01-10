package vm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/compiler"
	"github.com/caelondev/monkey-compiler-go/src/object"
)

func readUint32(buf *bytes.Reader) (uint32, error) {
	var val uint32
	err := binary.Read(buf, binary.BigEndian, &val)
	return val, err
}

func readString(buf *bytes.Reader) (string, error) {
	// Get string length
	length, err := readUint32(buf)
	if err != nil {
		return "", err
	}

	// Allocate bytes based on string length
	strBytes := make([]byte, length)

	// Set allocated bytes to char bytes
	if _, err := buf.Read(strBytes); err != nil {
		return "", err
	}

	return string(strBytes), nil
}

func readFloat64(buf *bytes.Reader) (float64, error) {
	var val float64
	err := binary.Read(buf, binary.BigEndian, &val)
	return val, err
}

func DecodeBytecode(data []byte) (*compiler.Bytecode, error) {
	buf := bytes.NewReader(data)

	// Check magic
	magic := make([]byte, 4)
	if _, err := buf.Read(magic); err != nil {
		return nil, err
	}
	if string(magic) != "MCGO" {
		return nil, fmt.Errorf("invalid bytecode file: wrong magic")
	}

	// Check version
	versionByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	if versionByte != 1 {
		return nil, fmt.Errorf("unsupported bytecode version: %d", versionByte)
	}

	// Read constants
	constCount, err := readUint32(buf)
	if err != nil {
		return nil, err
	}

	constants := make([]object.Object, 0, constCount)

	for i := uint32(0); i < constCount; i++ {
		tag, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		switch tag {
		case byte(code.CONSTANT_NUMBER):
			num, err := readFloat64(buf)
			if err != nil {
				return nil, err
			}
			constants = append(constants, &object.Number{Value: num})
		case byte(code.CONSTANT_STRING):
			str, err := readString(buf)
			if err != nil {
				return nil, err
			}
			constants = append(constants, &object.String{Value: str})

		default:
			return nil, fmt.Errorf("unknown constant tag: %d", tag)
		}
	}

	instLen, err := readUint32(buf)
	if err != nil {
		return nil, err
	}

	instructions := make([]byte, instLen)
	if _, err := buf.Read(instructions); err != nil {
		return nil, err
	}

	return &compiler.Bytecode{
		Constants:    constants,
		Instructions: instructions,
	}, nil
}
