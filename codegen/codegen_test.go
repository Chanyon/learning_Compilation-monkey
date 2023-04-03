package codegen

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/parser"
	"strings"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParserProgram()
}

func TestCodeGen(t *testing.T) {
	test := `
		1+2;
	`
	program := parse(test)
	codegen := New()
	codegen.FreeAllRegisters()

	codegen.CgPreamble()
	reg := codegen.CodeGenAST(program)
	fmt.Println(reg)
	if reg == -1 {
		t.Fatal("codegen error\n")
	}
	codegen.CgPrintInt(reg)
	codegen.CgPostamble()

	content := strings.Join(codegen.Assembly, "")
	fmt.Println(content)
}
