package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

// 将宏保存到env中
func DefineMacro(program *ast.Program, env *object.Environment) {
	definitions := []int{}

	for i, statement := range program.Statements {
		if isMacroDefinition(statement) {
			addMacro(statement, env)
			definitions = append(definitions, i)
		}
	}

	for i := len(definitions) - 1; i >= 0; i -= 1 {
		idx := definitions[i]
		// 从AST切片中删除宏 macro
		program.Statements = append(program.Statements[:idx], program.Statements[idx+1:]...)
	}
}

func isMacroDefinition(node ast.Statement) bool {
	stmt, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = stmt.Value.(*ast.MacroLiteral)
	if !ok {
		return false
	}

	return true
}

func addMacro(node ast.Statement, env *object.Environment) {
	letStmt, _ := node.(*ast.LetStatement)
	macroLit, _ := letStmt.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLit.Parameters,
		Env:        env,
		Body:       macroLit.Body,
	}

	env.Set(letStmt.Name.Value, macro)
}

func ExpandMacros(program ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(program, func(node ast.Node) ast.Node {
		callExp, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}
		macro, ok := isMacroCall(callExp, env)
		if !ok {
			return node
		}

		args := quoteArgs(callExp)
		evalEnv := extendMacroEnv(macro, args)
		evaluated := Eval(macro.Body, evalEnv)

		quote, ok := evaluated.(*object.Quote)
		if !ok {
			panic("we only support returning AST-nodes from macro")
		}
		return quote.Node
	})
}

func isMacroCall(callExp *ast.CallExpression, env *object.Environment) (*object.Macro, bool) {
	ident, ok := callExp.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(ident.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

func quoteArgs(callExp *ast.CallExpression) []*object.Quote {
	args := []*object.Quote{}
	// callExp arg to Quote
	for _, arg := range callExp.Arguments {
		args = append(args, &object.Quote{Node: arg})
	}
	return args
}

func extendMacroEnv(macro *object.Macro, args []*object.Quote) *object.Environment {
	extended := object.NewEnclosedEnvironment(macro.Env)
	for paramIdx, param := range macro.Parameters {
		extended.Set(param.Value, args[paramIdx])
	}

	return extended
}
