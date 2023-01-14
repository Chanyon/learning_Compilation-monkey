package evaluator

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestDefineMacro(t *testing.T) {
	input := `
		let num = 1;
		let func = fn(x,y){x+y};
		let mac = macro(x,y){x+y;};
	`
	env := object.NewEnvironment()
	program := testParserProgram(input)
	DefineMacro(program, env)

	if len(program.Statements) != 2 {
		t.Fatalf("wrong number of statements. got=%d", len(program.Statements))
	}

	_, ok := env.Get("num")
	if ok {
		t.Fatalf("num should not be defined.")
	}
	_, ok = env.Get("func")
	if ok {
		t.Fatalf("func should not be defined.")
	}

	obj, ok := env.Get("mac")
	if !ok {
		t.Fatalf("macro not in env.")
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		t.Fatalf("object is not a macro. got=%T (%+v)", obj, obj)
	}

	if len(macro.Parameters) != 2 {
		t.Fatalf("wrong number of macro parameters. gto=%d", len(macro.Parameters))
	}

	if macro.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", macro.Parameters[0])
	}
	if macro.Parameters[1].String() != "y" {
		t.Fatalf("parameter is not 'y'. got=%q", macro.Parameters[1])
	}

	if macro.Body.String() != "(x + y)" {
		t.Fatalf("body is not %s. got=%q", "(x + y)", macro.Body.String())
	}
}

func testParserProgram(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParserProgram()
}

func TestExpandMacro(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`
		let infixExp = macro(){quote(1+2)};
		infixExp();
		`,
			"(1 + 2)"},
		{`
		let reverse = macro(a,b){quote(unquote(b) - unquote(a))};
		reverse(2+2,10-5);
		`,
			"(10 - 5) - (2 + 2)"},
		{
			`let unless=macro(condition,consequence,alternative){
			quote(
			if(!(unquote(condition))){
				unquote(consequence);
			}else{
				unquote(alternative);
			})
			};
			unless(10>5,puts("not greater."),puts("greater!"));`,
			`if(!(10 > 5)){puts("not greater.")}else{puts("greater!")}`,
		},
	}

	for _, tt := range tests {
		expected := testParserProgram(tt.expected)
		program := testParserProgram(tt.input)
		env := object.NewEnvironment()
		DefineMacro(program, env)
		expanded := ExpandMacros(program, env)

		if expanded.String() != expected.String() {
			t.Errorf("not equal. got=%q, want=%q", expanded.String(), expected.String())
		}
	}
}
