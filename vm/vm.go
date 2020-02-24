package vm

import (
	"fmt"

	"gocompiler/compiler"
	"gocompiler/ir"
	"gocompiler/opcode"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 2048

var True = &ir.Boolean{Value: true}
var False = &ir.Boolean{Value: false}
var Null = &ir.Null{}

type VM struct {
	constants   []ir.Object
	stack       []ir.Object
	sp          int // Always points to the next value. Top of stack is stack[sp-1]
	globals     []ir.Object
	frames      []*Frame
	framesIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	frames := make([]*Frame, MaxFrames)
	frames[0] = NewFrame(
		&ir.Closure{
			Function: &ir.CompiledFunction{
				Instructions: bytecode.Instructions,
			}}, 0)

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]ir.Object, StackSize),
		sp:          0,
		globals:     make([]ir.Object, GlobalSize),
		frames:      frames,
		framesIndex: 1,
	}
}

func (vm *VM) Run() error {
	var (
		ip  int
		ins opcode.Instructions
		op  opcode.Opcode
	)

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		op = opcode.Opcode(ins[ip])

		switch op {
		case opcode.OpConstant:
			constIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case opcode.OpPop:
			vm.pop()
		case opcode.OpAdd, opcode.OpSub, opcode.OpMul, opcode.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case opcode.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case opcode.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case opcode.OpEqual, opcode.OpNotEqual, opcode.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case opcode.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case opcode.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case opcode.OpJumpNotTruthy:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case opcode.OpJump:
			pos := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case opcode.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case opcode.OpSetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case opcode.OpGetGlobal:
			globalIndex := opcode.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case opcode.OpHash:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case opcode.OpArray:
			numElements := int(opcode.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case opcode.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case opcode.OpCall:
			numArgs := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case opcode.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case opcode.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case opcode.OpSetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()
		case opcode.OpGetLocal:
			localIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()

			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}
		case opcode.OpClosure:
			constIndex := opcode.ReadUint16(ins[ip+1:])
			numFree := opcode.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}

		case opcode.OpGetFree:
			freeIndex := opcode.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) StackTop() ir.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() ir.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) pop() ir.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) push(o ir.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) executeBinaryOperation(op opcode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == ir.IntegerObj && rightType == ir.IntegerObj:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == ir.StringObj && rightType == ir.StringObj:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op opcode.Opcode, left, right ir.Object) error {
	leftValue := left.(*ir.Integer).Value
	rightValue := right.(*ir.Integer).Value

	var result int64

	switch op {
	case opcode.OpAdd:
		result = leftValue + rightValue
	case opcode.OpSub:
		result = leftValue - rightValue
	case opcode.OpMul:
		result = leftValue * rightValue
	case opcode.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&ir.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op opcode.Opcode, left, right ir.Object) error {
	if op != opcode.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*ir.String).Value
	rightValue := right.(*ir.String).Value

	return vm.push(&ir.String{Value: leftValue + rightValue})
}

func (vm *VM) executeComparison(op opcode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == ir.IntegerObj || right.Type() == ir.IntegerObj {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d %s %s", op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op opcode.Opcode, left, right ir.Object) error {
	leftValue := left.(*ir.Integer).Value
	rightValue := right.(*ir.Integer).Value

	switch op {
	case opcode.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case opcode.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case opcode.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != ir.IntegerObj {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*ir.Integer).Value
	return vm.push(&ir.Integer{Value: -value})
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeIndexExpression(left, index ir.Object) error {
	switch {
	case left.Type() == ir.ArrayObj && index.Type() == ir.IntegerObj:
		return vm.executeArrayIndex(left, index)
	case left.Type() == ir.HashObj:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeHashIndex(hash, index ir.Object) error {
	hashObject := hash.(*ir.Hash)

	key, ok := index.(ir.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeArrayIndex(array, index ir.Object) error {
	arrayObject := array.(*ir.Array)
	i := index.(*ir.Integer).Value

	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) buildHash(startIndex, endIndex int) (ir.Object, error) {
	hashedPairs := make(map[ir.HashKey]ir.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := ir.HashPair{Key: key, Value: value}

		hashKey, ok := key.(ir.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &ir.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) buildArray(startIndex, endIndex int) ir.Object {
	elements := make([]ir.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &ir.Array{Elements: elements}
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *ir.Closure:
		return vm.callClosure(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*ir.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]ir.Object, numFree)

	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}

	vm.sp = vm.sp - numFree

	closure := &ir.Closure{Function: function, Free: free}
	return vm.push(closure)
}

func (vm *VM) callClosure(cl *ir.Closure, numArgs int) error {
	if numArgs != cl.Function.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Function.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Function.NumLocals

	return nil
}

func nativeBoolToBooleanObject(input bool) *ir.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func isTruthy(obj ir.Object) bool {
	switch obj := obj.(type) {
	case *ir.Null:
		return false
	case *ir.Boolean:
		return obj.Value
	default:
		return true
	}
}
