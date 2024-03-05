package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"monkey/token"
	"sort"
)

//	词法分析    语法分析      字节码     执行输出
//
// code--->token--->AST--->compiler--->VM
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// 引入作用域, 解决函数字节码与主程序的字节码指令纠缠问题
type CompilationScope struct {
	instruction         code.Instruction
	lastInstruction     EmittedInstruction // 最后一条指令
	previousInstruction EmittedInstruction // 倒数第二条
}

type CompilerCtx struct {
	isCompClass bool
	infixAssign bool
	infixDot    bool
	index       int
	// hasExtends  bool
}

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable //符号表, 保存、处理变量
	scopes      []CompilationScope
	scopeIndex  int
	compilerCtx *CompilerCtx
}

type ByteCode struct {
	Instruction code.Instruction
	Constants   []object.Object
}

func New() *Compiler {
	mainScope := CompilationScope{
		instruction:         code.Instruction{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	symbolTable := NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		compilerCtx: &CompilerCtx{isCompClass: false, index: -1, infixAssign: false},
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.constants = constants
	compiler.symbolTable = s
	return compiler
}

func (c *Compiler) currentInstructions() code.Instruction {
	return c.scopes[c.scopeIndex].instruction
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
		_, ok := node.Expression.(*ast.AssignExpression)
		if ok {
			c.removeLastOpPop()
		}
	case *ast.InfixExpression:
		if node.Operator == "&&" {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}
			endJump := c.emit(code.OpAnd, 9999)

			err = c.Compile(node.Right)
			if err != nil {
				return err
			}
			afterPos := len(c.currentInstructions())
			c.changeOperand(endJump, afterPos)
			return nil
		}

		if node.Operator == "||" {
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}
			elseJump := c.emit(code.OpOr, 9999)
			endJump := c.emit(code.OpJump, 9999)
			afterPos := len(c.currentInstructions())
			c.changeOperand(elseJump, afterPos)
			err = c.Compile(node.Right)
			if err != nil {
				return err
			}
			afterPos = len(c.currentInstructions())
			c.changeOperand(endJump, afterPos)
			return nil
		}

		if node.Operator == token.ASSIGN {
			c.compilerCtx.infixAssign = true
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		if node.Operator == token.DOT {
			c.compilerCtx.infixDot = true
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "<":
			c.emit(code.OpLessThan)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case "<=":
			c.emit(code.OpGreaterThan)
			c.emit(code.OpBang)
		case ">=":
			c.emit(code.OpLessThan)
			c.emit(code.OpBang)
		// case "=":
		// 	// this.foo = 1;
		// 	if c.compilerCtx.index >= 0 && c.compilerCtx.infixDot {
		// 		c.emit(code.OpSetProperty, c.compilerCtx.index)
		// 	}
		// 	c.compilerCtx.infixAssign = false
		// 	c.compilerCtx.index = -1
		// case ".":
		// 	// foo.a
		// 	c.compilerCtx.infixDot = false
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	// case *ast.ThisLiteral:
	// c.compilerCtx.infixDot = true
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		//虚假的偏移量9999 jump not truthy
		jumpNotTPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastOpPop()
		}
		// else {
		// 	return fmt.Errorf("syntax error: if(true){ _ }")
		// }

		jumpPos := c.emit(code.OpJump, 9999)
		//回填操作,修正偏移量
		// afterConsequencePos := len(c.instructions)
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {

			// afterConsequencePos := len(c.instructions)
			// c.changeOperand(jumpNotTPos,afterConsequencePos)

			//else {}
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastOpPop()
			}
		}
		//修正跳出备选位置 9999 -> len(c.instructions)
		// afterAlternativePos := len(c.instructions)
		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, int(code.SetTypeVar), symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable `%s`", node.Value)
		}

		c.loadSymbol(symbol)
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, ele := range node.Elements {
			err := c.Compile(ele)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for key := range node.Pairs {
			keys = append(keys, key)
		}
		// 排序为了方便测试
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastOpPopToOpReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		numLocals := c.symbolTable.numDefinitions
		freeSymbols := c.symbolTable.FreeSymbol
		instruction := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instruction,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
			Name:          node.Name,
		}
		// c.emit(code.OpConstant, c.addConstant(compiledFn))
		constantFnIndex := c.addConstant(compiledFn)

		if c.compilerCtx.isCompClass {
			c.emit(code.OpMethod, constantFnIndex, len(freeSymbols))
		} else {
			c.emit(code.OpClosure, constantFnIndex, len(freeSymbols))
		}

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))
	case *ast.WhileStatement:
		loopStart := len(c.currentInstructions())

		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}
		c.emit(code.OpLoop, loopStart)
		afterPos := len(c.currentInstructions())
		c.changeOperand(jumpNotPos, afterPos)
	case *ast.AssignExpression:
		var symbol Symbol
		var set_type byte
		switch child_node := node.Left.(type) {
		case *ast.Identifier:
			symbol = c.symbolTable.Define(child_node.Value)
			set_type = code.SetTypeVar
		case *ast.IndexExpression:
			symbol = c.symbolTable.Define(child_node.Left.(*ast.Identifier).Value)
			set_type = code.SetTypeArray
			err := c.Compile(child_node.Index)
			if err != nil {
				return nil
			}

		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			switch set_type {
			case code.SetTypeVar:
				c.emit(code.OpSetGlobal, int(code.SetTypeVar), symbol.Index)
			case code.SetTypeArray:
				c.emit(code.OpSetGlobal, int(code.SetTypeArray), symbol.Index)

			}
		} else {
			switch set_type {
			case code.SetTypeVar:
				c.emit(code.OpSetLocal, symbol.Index)
			case code.SetTypeArray:
				c.emit(code.OpSetLocal, symbol.Index)
			}
		}
	case *ast.ForStatement:
		// loopStart := len(c.currentInstructions())
		loopStart := 0
		// c.enterScope()
		err := c.Compile(node.LetStmt)
		if err != nil {
			return err
		}
		letEndPos := len(c.currentInstructions())

		err = c.Compile(node.Condition)
		if err != nil {
			return err
		}
		jumpNotPos := c.emit(code.OpJumpNotTruthy, 9999)
		jumpPos := c.emit(code.OpJump, 9999)

		incStart := len(c.currentInstructions())

		err = c.Compile(node.Inc)
		if err != nil {
			return err
		}
		c.emit(code.OpLoop, loopStart+letEndPos)
		incEnd := len(c.currentInstructions())
		c.changeOperand(jumpPos, incEnd+loopStart)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}
		c.emit(code.OpLoop, incStart+loopStart)
		jumpNotEndPos := len(c.currentInstructions())
		c.changeOperand(jumpNotPos, jumpNotEndPos+loopStart)

		// instruction := c.leaveScope()
		// c.scopes[c.scopeIndex].instruction = append(c.scopes[c.scopeIndex].instruction, instruction...)
		// fmt.Println(instruction)

	case *ast.ClassStmt:
		c.compilerCtx.isCompClass = true
		symbol := c.symbolTable.Define(node.Name.Value)
		//emit OP_CLASS create a class
		c.emit(code.OpClass)
		c.emit(code.OpDefineClass, symbol.Index)

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
		c.compilerCtx.isCompClass = false
	} //switch end
	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// 生成指令并将其添加到最终结果
// operands是操作数在常量池里的index
// pos操作码位于指令集合中的位置
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	// posNewInstruction := len(c.instructions)
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	// c.instructions = append(c.instructions, ins...)
	c.scopes[c.scopeIndex].instruction = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		// Instruction: c.instructions,
		Instruction: c.currentInstructions(),
		Constants:   c.constants,
	}
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	// c.previousInstruction = previous
	// c.lastInstruction = last
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	// return c.lastInstruction.Opcode == code.OpPop
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastOpPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	prev := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]
	c.scopes[c.scopeIndex].instruction = new
	c.scopes[c.scopeIndex].lastInstruction = prev
}

func (c *Compiler) changeOperand(opPos int, operand int) {

	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceInstruction(opPos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[opPos+i] = newInstruction[i]
	}
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instruction:         code.Instruction{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex += 1
	// scope enclose
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instruction {
	instruction := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex -= 1
	c.symbolTable = c.symbolTable.Outer
	return instruction
}

func (c *Compiler) replaceLastOpPopToOpReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		if c.compilerCtx.infixDot {
			c.emit(code.OpGetProperty, s.Index)
			return
		}
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFreeVar, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrnetClosure)
	}
}
