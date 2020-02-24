package vm

import (
	"fmt"
	"testing"

	"gocompiler/ast"
	"gocompiler/compiler"
	"gocompiler/ir"
	"gocompiler/lexer"
	"gocompiler/parser"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!(if (false) { 5; })", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 }", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", Null},
		{"if (false) { 10 }", Null},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVmTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVmTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[ir.HashKey]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[ir.HashKey]int64{
				(&ir.Integer{Value: 1}).HashKey(): 2,
				(&ir.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[ir.HashKey]int64{
				(&ir.Integer{Value: 2}).HashKey(): 4,
				(&ir.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", Null},
		{"[1, 2, 3][99]", Null},
		{"[1][-1]", Null},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", Null},
		{"{}[0]", Null},
	}

	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let returnsOne = function() { 1; };
			let returnsOneReturner = function() { returnsOne; };
			returnsOneReturner()();
			`,
			expected: 1,
		},
		{
			input: `
			let returnsOneReturner = function() {
				let returnsOne = function() { 1; };
				returnsOne;
			};
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let identity = function(a) { a; };
			identity(4);
			`,
			expected: 4,
		},
		{
			input: `
			let sum = function(a, b) { a + b; };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = function(a, b) { 
				let c = a + b;
				c;
			};
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = function(a, b) { 
				let c = a + b;
				c;
			};
			sum(1, 2) + sum(3, 4);
			`,
			expected: 10,
		},
		{
			input: `
			let sum = function(a, b) { 
				let c = a + b;
				c;
			};
			let outer = function() {
				sum(1, 2) + sum(3, 4);
			}
			outer();
			`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;
			
			let sum = function(a, b) { 
				let c = a + b;
				c + globalNum;
			};

			let outer = function() {
				sum(1, 2) + sum(3, 4) + globalNum;
			}
			outer() + globalNum;
			`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `function() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input:    `function(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input:    `function(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		comp := compiler.New()

		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())

		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}

		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err.Error())
		}
	}
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let newClosure = function(a) {
				function() { a; };
			};
			let closure = newClosure(99);
			closure();
			`,
			expected: 99,
		},
		{
			input: `
		let newAdder = function(a, b) {
			function(c) { a + b + c };
		};
		let adder = newAdder(1, 2);
		adder(8);
		`,
			expected: 11,
		},
		{
			input: `
		let newAdder = function(a, b) {
			let c = a + b;
			function(d) { c + d };
		};
		let adder = newAdder(1, 2);
		adder(8);
		`,
			expected: 11,
		},
		{
			input: `
		let newAdderOuter = function(a, b) {
			let c = a + b;
			function(d) {
				let e = d + c;
				function(f) { e + f; };
			};
		};
		let newAdderInner = newAdderOuter(1, 2)
		let adder = newAdderInner(3);
		adder(8);
		`,
			expected: 14,
		},
		{
			input: `
		let a = 1;
		let newAdderOuter = function(b) {
			function(c) {
				function(d) { a + b + c + d };
			};
		};
		let newAdderInner = newAdderOuter(2)
		let adder = newAdderInner(3);
		adder(8);
		`,
			expected: 14,
		},
		{
			input: `
		let newClosure = function(a, b) {
			let one = function() { a; };
			let two = function() { b; };
			function() { one() + two(); };
		};
		let closure = newClosure(9, 90);
		closure();
		`,
			expected: 99,
		},
	}

	runVmTests(t, tests)
}

func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
		let fibonacci = function(x) {
			if (x == 0) {
				return 0;
			} else {
				if (x == 1) {
					return 1;
				} else {
					fibonacci(x - 1) + fibonacci(x - 2);
				}
			}
		};
		fibonacci(15);
		`,
			expected: 610,
		},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()

		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual ir.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case []int:
		array, ok := actual.(*ir.Array)
		if !ok {
			t.Errorf("ir not Array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.Elements))
			return
		}

		for i, expectedElm := range expected {
			err := testIntegerObject(int64(expectedElm), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[ir.HashKey]int64:
		hash, ok := actual.(*ir.Hash)
		if !ok {
			t.Errorf("ir is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in pairs")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *ir.Null:
		if actual != Null {
			t.Errorf("ir is not Null: %T (%+v)", actual, actual)
		}
	case *ir.Error:
		errObj, ok := actual.(*ir.Error)
		if !ok {
			t.Errorf("ir is not Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errObj.Message)
		}
	}
}

func testIntegerObject(expected int64, actual ir.Object) error {
	result, ok := actual.(*ir.Integer)
	if !ok {
		return fmt.Errorf("ir is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("ir has wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}

func testBooleanObject(expected bool, actual ir.Object) error {
	result, ok := actual.(*ir.Boolean)
	if !ok {
		return fmt.Errorf("ir is not Bool. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("ir has wrong value. got=%t, want=%t",
			result.Value, expected)
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
