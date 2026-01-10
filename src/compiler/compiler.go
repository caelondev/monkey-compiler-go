package compiler

import (
	"fmt"

	"github.com/caelondev/monkey-compiler-go/src/ast"
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/token"
)

type EmittedInstruction struct {
	OpCode   code.OpCode
	Position int
}

type Compiler struct {
	instructions code.Instructions // []byte
	constants    []object.Object
	symbolTable  *SymbolTable

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions // []byte
	Constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: make(code.Instructions, 0),
		constants:    make([]object.Object, 0),
		symbolTable:  NewSymbolTable(),

		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
}

func NewWithState(table *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = table
	compiler.constants = constants
	return compiler
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

		c.emit(code.OpPop)

	case *ast.BooleanExpression:
		var bool code.OpCode

		if node.Value {
			bool = code.OpTrue
		} else {
			bool = code.OpFalse
		}

		c.emit(bool)

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.NilLiteral:
		c.emit(code.OpNil)

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
		case token.CARET:
			c.emit(code.OpExponent)
		case token.PLUS:
			c.emit(code.OpAdd)
		case token.MINUS:
			c.emit(code.OpSubtract)
		case token.STAR:
			c.emit(code.OpMultiply)
		case token.SLASH:
			c.emit(code.OpDivide)

		case token.EQUAL:
			c.emit(code.OpEqual)
		case token.NOT_EQUAL:
			c.emit(code.OpNotEqual)
		case token.LESS:
			c.emit(code.OpLess)
		case token.GREATER:
			c.emit(code.OpGreater)
		case token.LESS_EQUAL:
			c.emit(code.OpLessEqual)
		case token.GREATER_EQUAL:
			c.emit(code.OpGreaterEqual)

		default:
			return fmt.Errorf("Unknown binary operator token: '%s'", node.Operator.Type)
		}

	case *ast.NumberLiteral:
		num := &object.Number{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(num))

	case *ast.AbsoluteExpression:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		c.emit(code.OpAbsolute)

	case *ast.UnaryExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.NOT:
			c.emit(code.OpNot)
		case token.MINUS:
			c.emit(code.OpNegate)
		default:
			return fmt.Errorf("Unknown unary operator token: '%s'", node.Operator.Type)
		}

	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.IfStatement:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit with some bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		if node.Alternative == nil {
			// Reassign jump pos to the end of if stmt address
			posAfterConsequence := len(c.instructions)
			c.changeOperand(jumpNotTruthyPos, posAfterConsequence)
		} else {
			// Emit with some bogus value
			jumpPos := c.emit(code.OpJump, 9999)
			posAfterConsequence := len(c.instructions)
			c.changeOperand(jumpNotTruthyPos, posAfterConsequence)

			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIsPop() {
				c.removeLastPop()
			}

			posAfterAlternative := len(c.instructions)
			c.changeOperand(jumpPos, posAfterAlternative)
		}

	case *ast.TernaryExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit with bogus value / placeholder ---
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		// Emit with bogus value / placeholder ---
		jumpPos := c.emit(code.OpJump, 9999)
		posAfterConsequence := len(c.instructions)

		// Set end of jumpNotTruthyPos to "jump pos"
		// But we're not directly using jumpPos
		c.changeOperand(jumpNotTruthyPos, posAfterConsequence)

		err = c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

		posAfterAlternative := len(c.instructions)
		c.changeOperand(jumpPos, posAfterAlternative)

	case *ast.VarStatement:
		for _, name := range node.Names {
			// TODO: This operation is lowkey expensive
			// maybe optimize this? this probably doesnt affect the
			// runtime that much, but its better to point this one out
			// ... Maybe recompiling is the only option???
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}

			symbol, error := c.symbolTable.Define(name.Value)
			if error {
				return fmt.Errorf("Cannot redeclare already existing variable '%s'", name.Value)
			}

			c.emit(code.OpSetGlobal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, exists := c.symbolTable.Resolve(node.Value)
		if !exists {
			return fmt.Errorf("Cannot resolve variable '%s'", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)

	case *ast.IndexSliceExpression:
		err := c.Compile(node.Target)
		if err != nil {
			return err
		}

		if node.Start != nil {
			err = c.Compile(node.Start)
			if err != nil {
				return err
			}
		} else {
			c.emit(code.OpNil)
		}

		if node.End != nil {
			err = c.Compile(node.End)
			if err != nil {
				return err
			}
		} else {
			c.emit(code.OpNil)
		}

		c.emit(code.OpSlice)

	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			err := c.Compile(element)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	default:
		return fmt.Errorf("Unknown AST node: '%s' (%T)", node.String(), node)
	}

	return nil
}

func (c *Compiler) Disassemble() {
	bytecode := c.Bytecode()

	fmt.Println("\n== Disassembler ==")
	instructions := bytecode.Instructions
	i := 0
	for i < len(instructions) {
		op := code.OpCode(instructions[i])
		def, err := code.Lookup(op)
		if err != nil {
			fmt.Printf("Unknown opcode %d\n", op)
			i++
			continue
		}
		fmt.Printf("%04d %s", i, def.Name)
		i++
		operands := make([]int, len(def.OperandWidths))
		for j, width := range def.OperandWidths {
			switch width {
			case 2:
				operands[j] = int(code.ReadUint16(instructions[i:]))
				i += 2
			default:
				panic("Unsupported operand width")
			}
		}
		if len(operands) > 0 {
			fmt.Printf(" %v", operands)
		}
		fmt.Println()
	}

	fmt.Println("\nConstants:")
	for idx, c := range bytecode.Constants {
		switch v := c.(type) {
		case *object.Number:
			fmt.Printf("%d: %g\n", idx, v.Value)
		case *object.String:
			fmt.Printf("%d: \"%s\"\n", idx, v.Value)

		default:
			fmt.Printf("%d: unknown constant type %T\n", idx, c)
		}
	}
	fmt.Println()
}

func (c *Compiler) emit(opcode code.OpCode, operands ...int) int {
	instruction := code.Make(opcode, operands...)
	position := c.addInstruction(instruction)

	c.setLastInstruction(opcode, position)
	return position
}

func (c *Compiler) setLastInstruction(opcode code.OpCode, position int) {
	// Shifts instructions
	previous := c.lastInstruction
	last := EmittedInstruction{OpCode: opcode, Position: position}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	// Get opcode on given position
	opcode := code.OpCode(c.instructions[opPos])

	// Attach an operand to the opcode
	newInstruction := code.Make(opcode, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		// Replaces all instruction bytes in the given offset
		c.instructions[position+i] = newInstruction[i]
	}
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.OpCode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	// resets the instructions up until the last instruction position
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1 // Return the object "Address"
}

func (c *Compiler) addInstruction(ins []byte) int {
	insPos := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return insPos // Return instruction "address"
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
