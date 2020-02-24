package vm

import (
	"gocompiler/ir"
	"gocompiler/opcode"
)

type Frame struct {
	cl          *ir.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *ir.Closure, basePointer int) *Frame {
	return &Frame{
		ip:          -1,
		cl:          cl,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() opcode.Instructions {
	return f.cl.Function.Instructions
}
