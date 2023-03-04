package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5+5+5+5-10", 10},
		{"2*2*2*2*2", 32},
		{"-50+100+(-50)", 0},
		{"5*2+10", 20},
		{"20+2*(-10)", 0},
		{"50/2*2+10", 60},
		{"2*(5+10)", 30},
		{"3*(3*3)", 27},
		{"(5+10*2+15/3)*2-10", 50},
		{"-6/-2-1", 2},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 <= 2", true},
		{"1 > 2", false},
		{"1 >=2", false},
		{"1 < 1", false},
		{"1>1", false},
		{"1==1", true},
		{"1!=1", false},
		{"1==2", false},
		{"1!=2", true},
		{"true == false", false},
		{"true == true", true},
		{"true != true", false},
		{"1<2 != true", false},
		{"1<2 == true", true},
		{"true && true", true},
		{"false || true", true},
		{"false && true", false},
		{"false && false", false},
		{"1 < 2 && 2 >= 1", true},
		{"1 > 2 || 2 >= 1", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if(true){10}", 10},
		{"if(false){10}", nil},
		{"if(1){10}", 10},
		{"if(1<2){10}", 10},
		{"if(1>2){10}", nil},
		{"if(1<2){10}else{10}", 10},
		{"if(1>2){10}else{20}", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"return 10", 10},
		{"return 10;9;", 10},
		{"return 2 *5;9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if(10 > 1){if(10>1){return 10;} return 1;}", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, int64(tt.expected.(int)))
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + false; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true;", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5;true + false;5;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if(10 > 1){ true + false;}", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if(10 > 1){if(10>1){return true + false;} return 1;}", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if(10 > 1){if(10>true){return 10;} return 1;}", "type mismatch: INTEGER > BOOLEAN"},
		{"foobar;", "identifier not found: foobar"},
		{`"hello" - "world"`, "unknown operator: STRING - STRING"},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)
		err, ok := evaluated.(*object.Error)

		if !ok {
			t.Errorf("no error object returned in %d row. got=%T (%+v)", i+1, evaluated, evaluated)
		}

		if err.Message != tt.expectedMessage {
			t.Errorf("wrong error message in %d row. expected=%q, got=%q", i+1, tt.expectedMessage, err.Message)
		}
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5;a;", 5},
		{"let a = 5 * 5;a;", 25},
		{"let a = 5;let b = a;b;", 5},
		{"let a = 1;let b = a;let c= a+b+1;c;", 3},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := `fn(x) {x+2;}`

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameter. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not '%q'. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let ident = fn(x){x;};ident(5);", 5},
		{"let ident=fn(x){return x;};ident(5);", 5},
		{"let double=fn(x){x*2;};double(5);", 10},
		{"let add=fn(x,y){x+y;};add(5,5);", 10},
		{"let add=fn(x,y){x+y;};add(5+5,add(5,5));", 20},
		{"fn(x){x;}(5);", 5},
		{"let a = fn(x){return x;}(5);a;", 5},
		{"let a=5;let b=fn(x){ x + a;};b(2)", 7},
		{"let fib=fn(n){if(n < 3){return 1;}else{return fib(n-1)+fib(n-2);}};fib(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
		let newAddr = fn(x){
			fn(y){ x + y };
		};
		let addTwo = newAddr(2);
		addTwo(2);
	`
	testIntegerObject(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "hello world!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"hello" + " " + "world!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "hello world!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one","two")`, "wrong number of arguments. got 2, want 1"},
		{"first(1)", "argument to `first` must be an array, got INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			err, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object isn't Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if err.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, err.Message)
			}
		}
	}
}

func TestArrayLiteral(t *testing.T) {
	input := "[1,2*2,3+3]"

	evaluated := testEval(input)
	array, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object isn't Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(array.ELements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(array.ELements))
	}

	testIntegerObject(t, array.ELements[0], 1)
	testIntegerObject(t, array.ELements[1], 4)
	testIntegerObject(t, array.ELements[2], 6)
}

func TestArrayIndexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1,2,3][0]", 1},
		{"let i = 0;[1][i];", 1},
		{"[1,2,3][1+1];", 3},
		{"let arr = [1,2,3];arr[0]+arr[1]+arr[2];", 6},
		{"[1,2,3][3]", nil},
		{"[1,2,3][-1]", nil},
	}

	for _, tt := range tests {
		eval := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, eval, int64(integer))
		} else {
			testNullObject(t, eval)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
		{
			"one": 10-9,
			two: 1+1,
			"thr" + "ee": 3,
			4:4,
			true: 5,
			false: 6,
		}
	`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of paris. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs.")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo":5}["foo"]`, 5},
		{`{"foo":5}["bar"]`, nil},
		{`let key = "foo";{"foo":5}[key]`, 5},
		{`{}["foo"]`, nil},
		{`{5:5}[5]`, 5},
		{`{true:5}[true]`, 5},
		{`{false:5}[false]`, 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if !ok {
			testNullObject(t, evaluated)
		} else {
			testIntegerObject(t, evaluated, int64(integer))
		}
	}
}

func TestErrorHanding(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			`{"name":"monkey"}[fn(x){x}]`,
			"unusable as hash key: FUNCTION",
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		err, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("object isn't Error. got=%T (%+v)", evaluated, evaluated)
		}
		if err.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, err.Message)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not null. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParserProgram()
	return Eval(program, object.NewEnvironment())
}

func testIntegerObject(t *testing.T, evaluated object.Object, expected int64) bool {
	result, ok := evaluated.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", evaluated, evaluated)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d,want=%d", result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, evaluated object.Object, expected bool) bool {
	result, ok := evaluated.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", evaluated, evaluated)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t,want=%t", result.Value, expected)
		return false
	}
	return true
}
