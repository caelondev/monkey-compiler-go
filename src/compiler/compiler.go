package compiler

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/ast"
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/token"
)

type Compiler struct {
	instructions code.Instructions // []byte
	constants    []object.Object
}

type Bytecode struct {
	Instructions code.Instructions // []byte
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: make(code.Instructions, 0),
		constants:    make([]object.Object, 0),
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

	case *ast.BinaryExpression:
		leftErr := c.Compile(node.Left)
		if leftErr != nil {
			return leftErr
		}

		rightErr := c.Compile(node.Right)
		if rightErr != nil {
			return rightErr
		}

		switch node.Operator.Type {
		case token.PLUS:
			c.emit(code.OpAdd)
		default:
			return fmt.Errorf("Unknown binary operator token: '%s'", node.Operator.Type)
		}

	case *ast.NumberLiteral:
		num := &object.Number{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(num))
	}

	return nil // Default
}

func (c *Compiler) emit(opcode code.OpCode, operands ...int) int {
	instruction := code.Make(opcode, operands...)
	position := c.addInstruction(instruction)
	return position
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1 // Return the object "Address"
}

func (c *Compiler) addInstruction(ins []byte) int {
	insPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return insPos
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
