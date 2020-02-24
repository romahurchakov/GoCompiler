package compiler

import (
	"fmt"
	"sort"

	"gocompiler/ast"
	"gocompiler/ir"
	"gocompiler/opcode"
)

type Compiler struct {
	constants   []ir.Object
	symbolTable *SymbolTable
	scopes      []CompilationScope
	scopeIndex  int
}

type Bytecode struct {
	Instructions opcode.Instructions
	Constants    []ir.Object
}

type EmittedInstruction struct {
	Opcode   opcode.Opcode
	Position int
}

type CompilationScope struct {
	instructions        opcode.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        opcode.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	return &Compiler{
		constants:   []ir.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(opcode.OpPop)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(opcode.OpGreaterThan)

			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(opcode.OpAdd)
		case "-":
			c.emit(opcode.OpSub)
		case "*":
			c.emit(opcode.OpMul)
		case "/":
			c.emit(opcode.OpDiv)
		case ">":
			c.emit(opcode.OpGreaterThan)
		case "==":
			c.emit(opcode.OpEqual)
		case "!=":
			c.emit(opcode.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(opcode.OpBang)
		case "-":
			c.emit(opcode.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPos := c.emit(opcode.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(opcode.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(opcode.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(opcode.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(opcode.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternative := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternative)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(opcode.OpIndex)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpCall, len(node.Arguments))
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(opcode.OpSetGlobal, symbol.Index)
		} else {
			c.emit(opcode.OpSetLocal, symbol.Index)
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(opcode.OpReturnValue)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)
	case *ast.IntegerLiteral:
		integer := &ir.Integer{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(opcode.OpTrue)
		} else {
			c.emit(opcode.OpFalse)
		}
	case *ast.StringLiteral:
		str := &ir.String{Value: node.Value}
		c.emit(opcode.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		keys := []ast.Expression{}

		for k := range node.Pairs {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(opcode.OpHash, len(node.Pairs)*2)
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(opcode.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(opcode.OpReturnValue) {
			c.emit(opcode.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledfunction := &ir.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		functionIndex := c.addConstant(compiledfunction)
		c.emit(opcode.OpClosure, functionIndex, len(freeSymbols))
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj ir.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op opcode.Opcode, operands ...int) int {
	ins := opcode.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op opcode.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == opcode.OpPop
}

func (c *Compiler) lastInstructionIs(op opcode.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) previousInstructionIs(op opcode.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].previousInstruction.Opcode == op
}

func (c *Compiler) removePreviousPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := append(old[:previous.Position], old[last.Position:]...)

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].previousInstruction = last
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := opcode.Opcode(c.currentInstructions()[opPos])
	newInstruction := opcode.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) currentInstructions() opcode.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        opcode.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() opcode.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, opcode.Make(opcode.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = opcode.OpReturnValue
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(opcode.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(opcode.OpGetLocal, s.Index)
	case FreeScope:
		c.emit(opcode.OpGetFree, s.Index)
	}
}
