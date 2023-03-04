package vm

import (
	"fmt"
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParserProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	res, ok := actual.(*object.Integer)

	if !ok {
		return fmt.Errorf("object is not Integer.got=%T(%+v)", actual, actual)
	}
	if res.Value != expected {
		return fmt.Errorf("object val has wrong.got=%d,want=%d", res.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	res, ok := actual.(*object.String)

	if !ok {
		return fmt.Errorf("object is not Integer.got=%T(%+v)", actual, actual)
	}
	if res.Value != expected {
		return fmt.Errorf("object val has wrong.got=%q,want=%q", res.Value, expected)
	}

	return nil
}

type vmTestCase struct {
	input    string
	expected interface{} //栈顶元素
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"1 / 2", 0},
		{"2 + 2 * 2 / 2", 4},
		{"(1+2)*3", 9},
		{"-1 + 1", 0},
		{"-5", -5},
		{"-5 + (10 * 2) * (-1)", -25},
	}
	runVmTest(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 > 2", false},
		{"1 < 2", true},
		{"1 <= 2", true},
		{"1 < 1", false},
		{"1 > 1", false},
		{"2 >= 1", true},
		{"1 >= 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false ", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!1", false},
		{"!!true", true},
		{"!!false", false},
		{"!(if(false){5})", true},
		{"true && true", true},
		{"true && false", false},
		{"false || true", true},
	}
	runVmTest(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if(true){10};", 10},
		{"if(true){10}else{20};", 10},
		{"if(false){10}else{20};", 20},
		{"if(1){10};", 10},
		{"if(1<2){10};", 10},
		{"if(1<2){10}else{20};", 10},
		{"if(1>2){10}else{20};", 20},
		{"if(false){10}", Null},
		{"if((if(false){10})){ 10 }else{ 20 }", 20},
		// {"if(true){ };", Null},
	}

	runVmTest(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{"let one = 1;one;", 1},
		{"let one = 1;", 1}, //测试通过
		{"let one = 1;let two = 2;one+two", 3},
		{"let one = 1;let two= one + one;one+two;", 3},
		{"let one = 1;let two = one+1;two;", 2},
	}

	runVmTest(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon"+"key"`, "monkey"},
		{`"mon"+"key" + "ok"`, "monkeyok"},
	}

	runVmTest(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "[]",
			expected: []int{},
		},
		{
			input:    "[1, 2, 3]",
			expected: []int{1, 2, 3},
		},
		{
			input:    "[1+2,2-1,3*2]",
			expected: []int{3, 1, 6},
		},
	}

	runVmTest(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "{}",
			expected: map[object.HashKey]int64{},
		},
		{
			input: "{1:2, 3:4}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 3}).HashKey(): 4,
			},
		},
		{
			input: "{1+1:2, 3:4*3}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 2,
				(&object.Integer{Value: 3}).HashKey(): 12,
			},
		},
	}

	runVmTest(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1,2,3][1]", 2},
		{"[1,2,3][1+1]", 3},
		{"[[1,1,1]][0][0]", 1},
		{"[[1],[2,3]][1][0]", 2},
		{"[][0]", Null},
		{"[1,2][3]", Null},
		{"[1][-1]", Null},
		{"{1:1,2:2}[1]", 1},
		{"{1:1,2:2}[2]", 2},
		{"{1:1}[0]", Null},
		{"{}[0]", Null},
	}
	runVmTest(t, tests)
}

func TestCallFunctionWithoutArgs(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let six = fn(){ 3+3;};six();",
			expected: 6,
		},
		{
			input: `
				let one = fn(){ 1 };
				let two = fn(){ 2 };
				one() + two();
			`,
			expected: 3,
		},
		{
			input:    "let early = fn(){return 66;100;};early();",
			expected: 66,
		},
		{
			input:    "let early = fn(){return 66;return 100;};early();",
			expected: 66,
		},
		{
			input:    "let noReturn = fn(){};noReturn();",
			expected: Null,
		},
		{
			input: `let noReturn= fn(){ };
				let noReturnTwo = fn(){ noReturn(); };
				noReturnTwo();
				noReturn();
			`,
			expected: Null,
		},
		{
			input: `let noReturn= fn(){ };
				let noReturnTwo = fn(){ noReturn(); };
				noReturnTwo();
				noReturn();
			`,
			expected: Null,
		},
	}

	runVmTest(t, tests)
}

func TestCallFunction2s(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				let one = fn(){ 1 };
				let returnOne = fn(){ one; };
				returnOne()();
			`,
			expected: 1,
		},
		{
			input: `
				let o = fn(){ "one"; };
				let oo = fn(){ o; };
				let ooo = fn(){ oo; };
				ooo()()();
			`,
			expected: "one",
		},
		{
			input:    "fn(){ 1 }();",
			expected: 1,
		},
		{
			input: `
				let returnFn = fn(){
					let one = fn(){ 1 };
					one
				};
				returnFn()();
			`,
			expected: 1,
		},
	}

	runVmTest(t, tests)
}

func TestCallFunctionWithoutBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    "let six = fn(){ let one= 1; one + 5};six();",
			expected: 6,
		},
		{
			input: `
				let oneAndTwo = fn(){ let one=1;let two=2; one+two; };
				oneAndTwo();
			`,
			expected: 3,
		},
		{
			input: `
				let oneAndTwo = fn(){ let one=1;let two=2; one + two; };
				let thAndFour = fn(){ let th=3;let four=4; th + four; };
				oneAndTwo()+thAndFour();
			`,
			expected: 10,
		},
		{
			input: `
				let globalTen = 10;
				let oneAndTwoAndTen = fn(){ let one=1;let two=2; one + two + globalTen; };
				oneAndTwoAndTen();
			`,
			expected: 13,
		},
		{
			input: `
				let globalTen = fn(){ 10 };
				let oneAndTwoAndTen = fn(){ let one=1;let two=2; one + two + globalTen() + globalTen(); };
				oneAndTwoAndTen();
			`,
			expected: 23,
		},
		{
			input: `
				let global = 20;
				let minusOne = fn(){ let one=1; global - one;};
				let minusTwo = fn(){ let two=2; global - two;};
				minusOne() + minusTwo();
			`,
			expected: 37,
		},
	}

	runVmTest(t, tests)
}

func TestCallingFunctionWithBindingArgument(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				let four = fn(a){ a };
				four(4);
			`,
			expected: 4,
		},
		{
			input: `
				let sum = fn(a,b){ a+b };
				sum(3,3);
			`,
			expected: 6,
		},
		{
			input: `
				let sum = fn(a,b){let c= 4; a+b+c };
				let t = sum(3,3);
				sum(5,t);
			`,
			expected: 19,
		},
		{
			input: `
				let f = fn(){ 5 };
				let sum = fn(a,b,f){let c= 4; a+b+c+f(); };
				let t = sum(3,3,f);
				sum(5,t,f);
			`,
			expected: 29,
		},
		// {
		// 	input: `
		// 		let wrong = fn(){ 1 };
		// 		wrong(1);
		// 	`,
		// 	expected: "wrong number of parameter. got=1,want=0",
		// },
	}
	runVmTest(t, tests)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len([])`, 0},
		{
			`len(1)`,
			&object.Error{
				Message: "argument to `len` not supported, got INTEGER",
			},
		},
		{`last([1,2,3])`, 3},
		{"first([1,2,3])", 1},
		{"push([], 4);", []int{4}},
		{`push(1, 1)`,
			&object.Error{
				Message: "argument to `push` must be an array, got INTEGER",
			},
		},
		{`rest([1,2])`, []int{2}},
	}

	runVmTest(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
				let newClosure = fn(a){ fn() { a; }};
				let closure = newClosure(66);
				closure();
			`,
			expected: 66,
		},
		{
			input: `
				let newClosure = fn(a,b){ fn(c) { a+b+c; }};
				let closure = newClosure(1,1);
				closure(1);
			`,
			expected: 3,
		},
		{
			input: `
				let newClosure = fn(a,b){
					let c = a + b;
					if( c < 3) {
						fn(d) { c + d }						
					}else {
						fn(d) { c + d + 2}
					}
				};
				let closure = newClosure(1,2);
				closure(2);
			`,
			expected: 7,
		},
		{
			input: `
				let newClosure = fn(a,b){ 
					let c = a+b;
					fn(d) { 
						let e = c + d;
						fn(f){ e + f;}
					}
				};
				let closure = newClosure(1,1);
				closure(1)(1);
			`,
			expected: 4,
		},
		{
			input: ` 
 			let newClosure = fn(a, b) { 
 				let one = fn() { a; }; 
 				let two = fn() { b; }; 
 				fn() { one() + two(); }; 
 			}; 
 			let closure = newClosure(9, 9); 
 			closure(); 
 			`,
			expected: 18,
		},
	}

	runVmTest(t, tests)
}

// recursive
func TestRecursiveFibonacci(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					0
				} else {
					countDown(x - 1)
				}
			};
			countDown(1);
			`,
			expected: 0,
		},
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					0
				} else {
					countDown(x - 1)
				}
			};
			let wrapper = fn(){
				countDown(1);				
			}
			wrapper();
			`,
			expected: 0,
		},
		{
			input: `
				let wrapper = fn(){
					let countDown = fn(x) {
						if (x == 0) {
							0
						} else {
							countDown(x - 1)
						}
					};
					countDown(1)		
				}
				wrapper();
			`,
			expected: 0,
		},
		{
			input: `
			let fib = fn(x) {
				if (x == 0) {
					0
				}else {
					if (x == 1) {
						1
					}else{
						fib(x-1) + fib(x-2)						
					}
				}
			};
			fib(3);
			`,
			expected: 2,
		},
	}
	runVmTest(t, tests)
}

func TestWhileStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `let foo = 0; while(foo < 2) { let a = 0; let foo = a + 1;} foo;`,
			expected: 2,
		},
	}
	runVmTest(t, tests)
}

func TestAssignExpression(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `let foo = 0; foo = 2; foo;`,
			expected: 2,
		},
		{
			input:    `let a = 0; let fun = fn(){ a = 3;} fun(); a;`,
			expected: 3,
		},
		{
			input: `
				let a = 0; 
				let fun = fn(c){ 
					let bar = fn(c) { 
						a = c;
					};
					return bar;
				};
				fun(3)(4);
				a;
			`,
			expected: 3,
		},
	}
	runVmTest(t, tests)
}

func TestForStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `let b = 0; for(let a = 0; a < 3; a = a + 1) { puts(a); b = a; } b;`,
			expected: 2,
		},
	}
	runVmTest(t, tests)
}

func runVmTest(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		compiler := compiler.New()
		err := compiler.Compile(program)

		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		for i, constant := range compiler.ByteCode().Constants {
			fmt.Printf("Constant %d %p (%T): \n", i, constant, constant)
			switch constant := constant.(type) {
			case *object.CompiledFunction:
				fmt.Printf("Instructions: \n%s", constant.Instructions)
			case *object.Integer:
				fmt.Printf("Value: %d\n", constant.Value)
			}
			fmt.Println()
		}

		vm := New(compiler.ByteCode())
		err = vm.Run()

		if err != nil {
			t.Fatalf("vm run failed: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(
	t *testing.T,
	expected interface{},
	actual object.Object,
) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case *object.Null:
		if actual != Null {
			t.Errorf("object is not Null: %T(%+v)", actual, actual)
		}
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object is not Array: %T(%+v)", actual, actual)
			return
		}
		if len(array.ELements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.ELements))
		}
		for i, expected := range expected {
			err := testIntegerObject(int64(expected), array.ELements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T(%+v)", actual, actual)
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(hash.Pairs), len(expected))
		}

		for key, val := range expected {
			pair, ok := hash.Pairs[key]
			if !ok {
				t.Errorf("no pair for get key in Pairs")
				return
			}
			err := testIntegerObject(val, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is't Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expect=%q,got=%q", expected.Message, errObj.Message)
		}
	}
}

func testBooleanObject(expected bool, actual object.Object) error {
	res, ok := actual.(*object.Boolean)

	if !ok {
		return fmt.Errorf("object is not Boolean.got=%T(%+v)", actual, actual)
	}
	if res.Value != expected {
		return fmt.Errorf("object val has wrong.got=%t,want=%t", res.Value, expected)
	}

	return nil
}
