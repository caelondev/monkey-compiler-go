package compiler

import (
	"fmt"
	"sort"

	"github.com/caelondev/monkey-compiler-go/src/ast"
	"github.com/caelondev/monkey-compiler-go/src/code"
	"github.com/caelondev/monkey-compiler-go/src/object"
	"github.com/caelondev/monkey-compiler-go/src/token"
)

type EmittedInstruction struct {
	OpCode   code.OpCode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions // []byte
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants   []object.Object
	scopes      []CompilationScope
	scopeIndex  int
	symbolTable *SymbolTable

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions // []byte
	Constants    []object.Object
}

func New() *Compiler {
	globalScope := CompilationScope{
		instructions:        make(code.Instructions, 0),
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	for i, fn := range object.NativeFunctions {
		symbolTable.DefineNative(i, fn.Name)
	}

	return &Compiler{
		constants:   make([]object.Object, 0),
		scopes:      []CompilationScope{globalScope},
		scopeIndex:  0,
		symbolTable: symbolTable,

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

		switch node.Expression.(type) {
		case *ast.AssignmentExpression:
		case *ast.IndexAssignmentExpression:
		default:
			c.emit(code.OpPop)
		}

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

		c.enterBlockScope()
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		c.leaveBlockScope()
		if node.Alternative == nil {
			// Reassign jump pos to the end of if stmt address
			posAfterConsequence := len(c.currentInstructions())
			c.changeOperand(jumpNotTruthyPos, posAfterConsequence)
		} else {
			// Emit with some bogus value
			jumpPos := c.emit(code.OpJump, 9999)
			c.enterBlockScope()
			posAfterConsequence := len(c.currentInstructions())
			c.changeOperand(jumpNotTruthyPos, posAfterConsequence)

			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			c.leaveBlockScope()

			posAfterAlternative := len(c.currentInstructions())
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

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit with bogus value / placeholder ---
		jumpPos := c.emit(code.OpJump, 9999)
		posAfterConsequence := len(c.currentInstructions())

		// Set end of jumpNotTruthyPos to "jump pos"
		// But we're not directly using jumpPos
		c.changeOperand(jumpNotTruthyPos, posAfterConsequence)

		err = c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		posAfterAlternative := len(c.currentInstructions())
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

			symbol, err := c.symbolTable.Define(name.Value)
			if err != nil {
				return err
			}

			c.emitSetToScope(symbol)
		}

	case *ast.Identifier:
		symbol, err := c.symbolTable.Resolve(node.Value)
		if err != nil {
			return err
		}

		c.emitGetToScope(symbol)

	case *ast.AssignmentExpression:
		err := c.Compile(node.NewValue)
		if err != nil {
			return err
		}

		assignee, ok := node.Assignee.(*ast.Identifier)
		if !ok {
			return fmt.Errorf("Cannot ra-assign invalid left-hand expression")
		}

		symbol, err := c.symbolTable.Reassign(assignee.Value)
		if err != nil {
			return err
		}

		c.emitSetToScope(symbol)

	case *ast.BatchAssignmentStatement:
		for _, assignee := range node.Assignees {
			err := c.Compile(node.NewValue)
			if err != nil {
				return err
			}

			symbol, err := c.symbolTable.Reassign(assignee.Value)
			if err != nil {
				return err
			}

			c.emitSetToScope(symbol)
		}

	case *ast.IndexAssignmentExpression:
		err := c.Compile(node.Target)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		err = c.Compile(node.NewValue)
		if err != nil {
			return err
		}

		c.emit(code.OpSetIndex)

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

	case *ast.HashLiteral:
		var keys []ast.Expression
		for key := range node.Pairs {
			keys = append(keys, key)
		}

		// NOTE: This line ensures the keys are sorted
		// since Go maps are unordered
		// Implementing this on other language mighy not require
		// the sorting below
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			// Compile key
			err := c.Compile(key)
			if err != nil {
				return err
			}

			// Compile value
			err = c.Compile(node.Pairs[key])
			if err != nil {
				return err
			}
		}

		// *2 to include key ---
		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		err := c.Compile(node.Target)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Implicit nil return
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpNil)
			c.emit(code.OpReturnValue)
		}

		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{Instructions: instructions, NumLocals: numLocals, NumParameters: len(node.Parameters)}
		c.emit(code.OpConstant, c.addConstant(compiledFn))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

	case *ast.FunctionDeclarationStatement:
		// Define function
		symbol, err := c.symbolTable.Define(node.Name.Value)
		if err != nil {
			return err
		}

		c.enterScope()

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Implicit nil return
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpNil)
			c.emit(code.OpReturnValue)
		}

		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		compiledFn := &object.CompiledFunction{Instructions: instructions, NumLocals: numLocals, NumParameters: len(node.Parameters)}
		c.emit(code.OpConstant, c.addConstant(compiledFn))

		c.emitSetToScope(symbol)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	default:
		return fmt.Errorf("Unknown AST node: '%s' (%T)", node.String(), node)
	}

	return nil
}

func (c *Compiler) emit(opcode code.OpCode, operands ...int) int {
	instruction := code.Make(opcode, operands...)
	position := c.addInstruction(instruction)

	c.setLastInstruction(opcode, position)
	return position
}

func (c *Compiler) setLastInstruction(opcode code.OpCode, position int) {
	// Shifts instructions
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{OpCode: opcode, Position: position}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	// Get opcode on given position
	opcode := code.OpCode(c.scopes[c.scopeIndex].instructions[opPos])

	// Attach/Replace an operand to the opcode
	newInstruction := code.Make(opcode, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(position int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		// Replaces all instruction bytes in the given offset
		c.scopes[c.scopeIndex].instructions[position+i] = newInstruction[i]
	}
}

func (c *Compiler) lastInstructionIs(opcode code.OpCode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.OpCode == opcode
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction
	old := c.currentInstructions()
	newIns := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = newIns
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1 // Return the object "Address"
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) addInstruction(ins []byte) int {
	insPos := len(c.currentInstructions())
	newInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = newInstructions

	return insPos // Return instruction "address"
}

func (c *Compiler) emitGetToScope(symbol Symbol) {
	switch symbol.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, symbol.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, symbol.Index)
	case NativeScope:
		c.emit(code.OpGetNative, symbol.Index)
	}
}

func (c *Compiler) emitSetToScope(symbol Symbol) {
	if symbol.Scope == GlobalScope {
		c.emit(code.OpSetGlobal, symbol.Index)
	} else {
		c.emit(code.OpSetLocal, symbol.Index)
	}
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        make(code.Instructions, 0),
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.enterBlockScope()
}

func (c *Compiler) enterBlockScope() {
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	inst := c.currentInstructions()

	// Delete scope
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.leaveBlockScope()
	return inst
}

func (c *Compiler) leaveBlockScope() {
	c.symbolTable = c.symbolTable.Outer
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}
