package formatter

import (
	"monkey/ast"
)

type FormatConfig struct {
	maxLineLength uint32
	maxHashOnline uint32
}

type Formatter struct {
	indent uint32
	column uint32
	config *FormatConfig
}

func new() *Formatter {
	return &Formatter{
		indent: 0,
		column: 1,
		config: &FormatConfig{
			maxLineLength: 80,
			maxHashOnline: 3,
		},
	}
}

func ignoreSemicolonExpr(node ast.Node) bool {
	switch node.(type) {
	case *ast.IfExpression:
		return true
	case *ast.FunctionLiteral:
		return false
	}
	return false
}
