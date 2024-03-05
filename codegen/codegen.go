package codegen

import (
	"fmt"
	"monkey/ast"
	"os"
)

type CodeGen struct {
	freeReg  []int
	regList  []string
	Assembly []string
}

func New() *CodeGen {
	return &CodeGen{
		regList: []string{"r8", "r9", "r10", "r11"},
	}
}

// func (cg *CodeGen) generateCode(node ast.Node) {
// }

func (cg *CodeGen) CodeGenAST(node ast.Node) int {
	var reg int
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			reg = cg.CodeGenAST(stmt)
			if reg == -1 {
				return reg
			}
		}
	case *ast.ExpressionStatement:
		reg = cg.CodeGenAST(node.Expression)
		if reg == -1 {
			return reg
		}

	case *ast.InfixExpression:
		leftReg := cg.CodeGenAST(node.Left)
		if leftReg == -1 {
			return -1
		}
		rightReg := cg.CodeGenAST(node.Right)
		if rightReg == -1 {
			return rightReg
		}
		switch node.Operator {
		case "+":
			//add
			reg := cg.Add(leftReg, rightReg)
			return reg
		}
	case *ast.IntegerLiteral:
		reg = cg.Load(node.Value)
		//switch end
	}
	return reg
}

func (cg *CodeGen) Load(value int64) int {
	r := cg.allocatorRegister()
	if r == -1 {
		fmt.Println("Out of register!")
		os.Exit(0)
	}
	cg.Assembly = append(cg.Assembly, fmt.Sprintf("\tmov\t%s, %d\n", cg.regList[r], value))
	return r
}

func (cg *CodeGen) Add(r1 int, r2 int) int {
	cg.Assembly = append(cg.Assembly, fmt.Sprintf("\tadd\t%s, %s\n", cg.regList[r2], cg.regList[r1]))
	cg.freeRegister(r1)
	return r2
}

func (cg *CodeGen) FreeAllRegisters() {
	cg.freeReg = []int{1, 1, 1, 1}
}

func (cg *CodeGen) allocatorRegister() int {
	for idx, val := range cg.freeReg {
		if val == 1 {
			cg.freeReg[idx] = 0
			return idx
		}
	}
	return -1
}

func (cg *CodeGen) freeRegister(reg int) {
	if cg.freeReg[reg] != 0 {
		fmt.Printf("Error trying to free register %q\n.", reg)
	}
	cg.freeReg[reg] = 1
}

func (cg *CodeGen) CgPreamble() {
	preamble := `
	global	main
	extern	printf
	section	.text
LC0:	db	"%d",10,0
printint:
	push	rbp
	mov	rbp, rsp
	sub	rsp, 16
	mov	[rbp-4], edi
	mov	eax, [rbp-4]
	mov	esi, eax
	lea	rdi, [rel LC0]
	mov	eax, 0
	call	printf
	nop
	leave
	ret

main:
	push	rbp
	mov		rbp, rsp
	`
	cg.Assembly = append(cg.Assembly, preamble)
}

func (cg *CodeGen) CgPostamble() {
	postamble := `
	mov eax,	0
	pop rbp
	ret

	`
	cg.Assembly = append(cg.Assembly, postamble)
}

func (cg *CodeGen) CgPrintInt(reg int) {
	cg.Assembly = append(cg.Assembly, fmt.Sprintf("\tmov\trdi,	%s\n\tcall\tprintint", cg.regList[reg]))
}
