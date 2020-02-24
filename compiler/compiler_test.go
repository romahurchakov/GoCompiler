package compiler

import (
	"fmt"
	"testing"

	"gocompiler/ast"
	"gocompiler/ir"
	"gocompiler/lexer"
	"gocompiler/opcode"
	"gocompiler/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []opcode.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSub),
				opcode.Make(opcode.OpPop),
			},
		}, {
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpMul),
				opcode.Make(opcode.OpPop),
			},
		}, {
			input:             "1 / 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpDiv),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGreaterThan),
				opcode.Make(opcode.OpPop),
			},
		}, {
			input:             "1 < 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpGreaterThan),
				opcode.Make(opcode.OpPop),
			},
		}, {
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpEqual),
				opcode.Make(opcode.OpPop),
			},
		}, {
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpNotEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpNotEqual),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpMinus),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpFalse),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpTrue),
				opcode.Make(opcode.OpBang),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			if (true) { 10 }; 3333;
			`,
			expectedConstants: []interface{}{10, 3333},
			expectedInstructions: []opcode.Instructions{
				// 0000
				opcode.Make(opcode.OpTrue),
				// 0001
				opcode.Make(opcode.OpJumpNotTruthy, 10),
				// 0004
				opcode.Make(opcode.OpConstant, 0),
				// 0007
				opcode.Make(opcode.OpJump, 11),
				// 0010
				opcode.Make(opcode.OpNull),
				// 0011
				opcode.Make(opcode.OpPop),
				// 0012
				opcode.Make(opcode.OpConstant, 1),
				// 0015
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			if (true) { 10 } else { 20 }; 3333;
			`,
			expectedConstants: []interface{}{10, 20, 3333},
			expectedInstructions: []opcode.Instructions{
				// 0000
				opcode.Make(opcode.OpTrue),
				// 0001
				opcode.Make(opcode.OpJumpNotTruthy, 10),
				// 0004
				opcode.Make(opcode.OpConstant, 0),
				// 0007
				opcode.Make(opcode.OpJump, 13),
				// 0010
				opcode.Make(opcode.OpConstant, 1),
				// 0013
				opcode.Make(opcode.OpPop),
				// 0014
				opcode.Make(opcode.OpConstant, 2),
				// 0017
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let one = 1;
			let two = 2;
			`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpSetGlobal, 1),
			},
		},
		{
			input: `
			let one = 1;
			one;
			`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let one = 1;
			let two = one;
			two;
			`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpSetGlobal, 1),
				opcode.Make(opcode.OpGetGlobal, 1),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpArray, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpSub),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpMul),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpHash, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpHash, 6),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpConstant, 5),
				opcode.Make(opcode.OpMul),
				opcode.Make(opcode.OpHash, 4),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpArray, 3),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpConstant, 4),
				opcode.Make(opcode.OpAdd),
				opcode.Make(opcode.OpIndex),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpHash, 2),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpSub),
				opcode.Make(opcode.OpIndex),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `function() { return 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `function() { 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `function() { 1; 2 }`,
			expectedConstants: []interface{}{
				1,
				2,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.symbolTable

	compiler.emit(opcode.OpMul)

	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}

	compiler.emit(opcode.OpSub)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d", len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != opcode.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, opcode.OpAdd)
	}

	if compiler.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}

	if compiler.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(opcode.OpAdd)

	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != opcode.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, opcode.OpAdd)
	}

	previous := compiler.scopes[compiler.scopeIndex].previousInstruction
	if previous.Opcode != opcode.OpMul {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d", previous.Opcode, opcode.OpMul)
	}
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `function() { }`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpReturn),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 0, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `function() { 24 }();`,
			expectedConstants: []interface{}{
				24,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 1, 0),
				opcode.Make(opcode.OpCall, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let noArg = function() { 24 };
			noArg();
			`,
			expectedConstants: []interface{}{
				24,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 1, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpCall, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let oneArg = function(a) { };
			oneArg(24);
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpReturn),
				},
				24,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 0, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpCall, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let manyArg = function(a, b, c) { };
			manyArg(24, 25, 26);
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpReturn),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 0, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpCall, 3),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let oneArg = function(a) { a };
			oneArg(24);
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 0, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpCall, 1),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let manyArg = function(a, b, c) { a; b; c; };
			manyArg(24, 25, 26);
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpPop),
					opcode.Make(opcode.OpGetLocal, 2),
					opcode.Make(opcode.OpReturnValue),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 0, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpGetGlobal, 0),
				opcode.Make(opcode.OpConstant, 1),
				opcode.Make(opcode.OpConstant, 2),
				opcode.Make(opcode.OpConstant, 3),
				opcode.Make(opcode.OpCall, 3),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()

		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let num = 55;
			function() { num }
			`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetGlobal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpClosure, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			function() {
				let num = 55;
				num
			}
			`,
			expectedConstants: []interface{}{
				55,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			function() {
				let a = 55;
				let b = 77;
				a + b
			}
			`,
			expectedConstants: []interface{}{
				55,
				77,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 0),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpSetLocal, 1),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClosure(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			function(a) {
				function(b) {
					a + b
				}
			}
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpClosure, 0, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 1, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			function(a) {
				function(b) {
					function(c) {
						a + b + c
					}
				}
			}
			`,
			expectedConstants: []interface{}{
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetFree, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpClosure, 0, 2),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpClosure, 1, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpClosure, 2, 0),
				opcode.Make(opcode.OpPop),
			},
		},
		{
			input: `
			let global = 55;
		
			function() {
				let a = 66;
		
				function() {
					let b = 77;
		
					function() {
						let c = 88;
		
						global + a + b + c;
					}
				}
			}
			`,
			expectedConstants: []interface{}{
				55,
				66,
				77,
				88,
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 3),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpGetGlobal, 0),
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetFree, 1),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpAdd),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 2),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpGetFree, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpClosure, 4, 2),
					opcode.Make(opcode.OpReturnValue),
				},
				[]opcode.Instructions{
					opcode.Make(opcode.OpConstant, 1),
					opcode.Make(opcode.OpSetLocal, 0),
					opcode.Make(opcode.OpGetLocal, 0),
					opcode.Make(opcode.OpClosure, 5, 1),
					opcode.Make(opcode.OpReturnValue),
				},
			},
			expectedInstructions: []opcode.Instructions{
				opcode.Make(opcode.OpConstant, 0),
				opcode.Make(opcode.OpSetGlobal, 0),
				opcode.Make(opcode.OpClosure, 6, 0),
				opcode.Make(opcode.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func testInstructions(expected []opcode.Instructions, actual opcode.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%q", concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot=%q", i, concatted, actual)
		}
	}

	return nil
}

func concatInstructions(s []opcode.Instructions) opcode.Instructions {
	out := opcode.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

func testConstants(t *testing.T, expected []interface{}, actual []ir.Object) error {

	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d", len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		case []opcode.Instructions:
			switch actual := actual[i].(type) {
			case *ir.CompiledFunction:
				err := testInstructions(constant, actual.Instructions)
				if err != nil {
					return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
				}
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual ir.Object) error {
	result, ok := actual.(*ir.Integer)
	if !ok {
		return fmt.Errorf("ir is not Integer. got=%T (%+v)", result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual ir.Object) error {
	result, ok := actual.(*ir.String)
	if !ok {
		return fmt.Errorf("ir is not String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("ir has wrong value. got=%q, want=%q", result.Value, expected)
	}

	return nil
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
