package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
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

type compilerTestCase struct {
	input               string
	expectedConstants   []interface{}
	expectedInstruction []code.Instruction
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1;2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0), //0, 1是1，2在常量池中的index
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				//code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 * 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 / 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []interface{}{1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTest(t, tests)
}
func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 > 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []interface{}{2, 1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTest(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if(true){ 10 }; 3333;",
			expectedConstants: []interface{}{10, 3333},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue),
				code.Make(code.OpJumpNotTruthy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "if(true){10}else{20}; 3333;",
			expectedConstants: []interface{}{10, 20, 3333},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpTrue), // 0
				code.Make(code.OpJumpNotTruthy, 10),
				code.Make(code.OpConstant, 0), // 6
				code.Make(code.OpJump, 13),    // 9
				code.Make(code.OpConstant, 1), // 12
				code.Make(code.OpPop),         // 13
				code.Make(code.OpConstant, 2), // 16
				code.Make(code.OpPop),         // 17
			},
		},
		{
			input:             "if(false){ 10 }; 3333;",
			expectedConstants: []interface{}{10, 3333},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpFalse),
				code.Make(code.OpJumpNotTruthy, 10),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpJump, 11),
				code.Make(code.OpNull),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		// ! if(true){  }; syntax error
		// {
		// 	input:             "if(true){  }; 3333;",
		// 	expectedConstants: []interface{}{3333},
		// 	expectedInstruction: []code.Instruction{
		// 		code.Make(code.OpTrue),
		// 		code.Make(code.OpJumpNotTruthy, 10),
		// 		code.Make(code.OpJump, 11),
		// 		code.Make(code.OpNull),
		// 		code.Make(code.OpPop),
		// 		code.Make(code.OpConstant, 1),
		// 		code.Make(code.OpPop),
		// 	},
		// },
	}

	runCompilerTest(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "let one = 1; let two = 2;",
			expectedConstants: []interface{}{1, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input:             "let one = 1; one;",
			expectedConstants: []interface{}{1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "let one = 1; let two = one; one;",
			expectedConstants: []interface{}{1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "let one = 1;",
			expectedConstants: []interface{}{1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []interface{}{"monkey"},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon" + "key"`,
			expectedConstants: []interface{}{"mon", "key"},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1+2, 2-1, 3*2]",
			expectedConstants: []interface{}{1, 2, 2, 1, 3, 2},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []interface{}{},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2, 3:4}",
			expectedConstants: []interface{}{1, 2, 3, 4},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1+1:2, 3:4*3}",
			expectedConstants: []interface{}{1, 1, 2, 3, 4, 3},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1,2,3][1+1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2}[2-1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestFunctionsReturnValueAndWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn(){ return 5 + 10;};",
			expectedConstants: []interface{}{
				5, 10,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(){ 5 + 10;};",
			expectedConstants: []interface{}{
				5, 10,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(){ 1; 2};",
			expectedConstants: []interface{}{
				1, 2,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(){ };",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpReturn),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	comp := New()
	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 0)
	}

	globalSymbolTable := comp.symbolTable

	comp.emit(code.OpMul)

	comp.enterScope()
	if comp.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 1)
	}
	comp.emit(code.OpSub)
	if len(comp.scopes[comp.scopeIndex].instruction) != 1 {
		t.Errorf("instruction length wrong. got=%d", len(comp.scopes[comp.scopeIndex].instruction))
	}
	last := comp.scopes[comp.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction OpSub wrong. got=%d, want=%d", last.Opcode, code.OpSub)
	}

	if comp.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable.")
	}
	comp.leaveScope()
	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 0)
	}

	if comp.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table.")
	}
	if comp.symbolTable.Outer != nil {
		t.Errorf("compiler modified  global symbol table incorrectly.")
	}

	comp.emit(code.OpAdd)
	if len(comp.scopes[comp.scopeIndex].instruction) != 2 {
		t.Errorf("instructions length wrong. got=%d", len(comp.scopes[comp.scopeIndex].instruction))
	}
	last = comp.scopes[comp.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction OpSub wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}
	prev := comp.scopes[comp.scopeIndex].previousInstruction
	if prev.Opcode != code.OpMul {
		t.Errorf("lastInstruction OpSub wrong. got=%d, want=%d", prev.Opcode, code.OpMul)
	}
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn(){ 66 }();",
			expectedConstants: []interface{}{
				66,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let nArg = fn(){ 66 }; nArg();",
			expectedConstants: []interface{}{
				66,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let oneArg = fn(a){  }; oneArg(24);",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let Args = fn(a,b,c){  }; Args(1,2,3);",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpReturn),
				},
				1, 2, 3,
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let oneArg = fn(a){ a }; oneArg(24);",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 0, 0), //function
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let Args = fn(a,b,c){ a; b; c }; Args(1,2,3);",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				1, 2, 3,
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "let num = 6; fn(){ num };",
			expectedConstants: []interface{}{
				6,
				[]code.Instruction{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(){ let num = 6; num };",
			expectedConstants: []interface{}{
				6,
				[]code.Instruction{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: "let gNum = 2; fn(){ let num = 6; num + gNum };",
			expectedConstants: []interface{}{
				2,
				6,
				[]code.Instruction{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestBuiltin(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "len([]); push([], 1);",
			expectedConstants: []interface{}{1},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpArray, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetBuiltin, 5),
				code.Make(code.OpArray, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: "fn(){ len([1]); };",
			expectedConstants: []interface{}{
				1,
				[]code.Instruction{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpArray, 1),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTest(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: "fn(a){ fn(b){ a+b } }",
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpGetFreeVar, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instruction{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
				fn(a){
					fn(b){
						fn(c){
							a+b+c
						}
					}
				}
			`,
			expectedConstants: []interface{}{
				[]code.Instruction{
					code.Make(code.OpGetFreeVar, 0),
					code.Make(code.OpGetFreeVar, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instruction{
					code.Make(code.OpGetFreeVar, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instruction{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func TestRecursiveFun(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
				let current = fn(x){ current(x - 1) };
				current(1);
			`,
			expectedConstants: []interface{}{
				1,
				[]code.Instruction{
					code.Make(code.OpCurrnetClosure),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSub),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
				1,
			},
			expectedInstruction: []code.Instruction{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTest(t, tests)
}

func runCompilerTest(t *testing.T, tests []compilerTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)

		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		byteCode := compiler.ByteCode()
		err = testInstructions(tt.expectedInstruction, byteCode.Instruction)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(tt.expectedConstants, byteCode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(expected []code.Instruction, actual code.Instruction) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\n got=%q,\nwant=%q", actual, concatted)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instructions at %d. want=%d,got=%d", i, ins, actual[i])
		}
	}
	return nil
}

func concatInstructions(ins []code.Instruction) code.Instruction {
	out := code.Instruction{}

	for _, ins := range ins {
		out = append(out, ins...)
	}
	return out
}

func testConstants(expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants length: got=%d,want=%d", len(actual), len(expected))
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
		case []code.Instruction:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a fn: %T", i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}

	}
	return nil
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
		return fmt.Errorf("object val has wrong.got=%s,want=%s", res.Value, expected)
	}

	return nil
}
