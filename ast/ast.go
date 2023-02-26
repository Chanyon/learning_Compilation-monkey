package ast

import (
	"bytes"
	"fmt"
	"monkey/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string //print ast node, 还原语句
}

// 语句 Statement
type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// ast 根节点
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// let语句
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (let *LetStatement) statementNode() {}
func (let *LetStatement) TokenLiteral() string {
	return let.Token.Literal
}
func (let *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(let.TokenLiteral() + " ")
	out.WriteString(let.Name.String())
	out.WriteString(" = ")

	if let.Value != nil {
		out.WriteString(let.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// identifier
type Identifier struct {
	Token token.Token
	Value string
}

func (ls *Identifier) expressionNode() {}
func (ls *Identifier) TokenLiteral() string {
	return ls.Token.Literal
}
func (ls *Identifier) String() string {
	return ls.Value
}

// return statements
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      BlockStatement
}

func (w *WhileStatement) statementNode() {}
func (w *WhileStatement) TokenLiteral() string {
	return w.Token.Literal
}
func (w *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString(w.TokenLiteral() + " ")
	out.WriteString("(")
	out.WriteString(w.Condition.String())
	out.WriteString(") {\n")
	out.WriteString(w.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// 1+5;
// x+10;
// fn(x,y){ x+5; };
// Expression Statement
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// 处理数字
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// prefix expression
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// infix expression / binary expression
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(fmt.Sprintf(" %s ", ie.Operator))
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// boolean
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}
func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}
func (b *Boolean) String() string {
	return b.Token.Literal
}

// if expression
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ife *IfExpression) expressionNode() {}
func (ife *IfExpression) TokenLiteral() string {
	return ife.Token.Literal
}
func (ife *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("  if")
	out.WriteString(ife.Condition.String())
	out.WriteString(" {\n  ")
	out.WriteString(ife.Consequence.String())
	out.WriteString("\n  }")
	if ife.Alternative != nil {
		out.WriteString("else {\n  ")
		out.WriteString(ife.Alternative.String())
		out.WriteString(" \n  }")
	}
	return out.String()
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

// fn
type FunctionLiteral struct {
	Token      token.Token     // token fn
	Parameters []*Identifier   // parameter list
	Body       *BlockStatement // block statement
	Name       string          // let binding functionLiteral name
}

func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	// out.WriteString(" ")
	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(fmt.Sprintf("<%s>", fl.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString("{ ")
	out.WriteString(fl.Body.String())
	out.WriteString(" }")
	return out.String()
}

// callExpression
type CallExpression struct {
	Token     token.Token
	Function  Expression //ident or function literal
	Arguments []Expression
}

func (call *CallExpression) expressionNode() {}
func (call *CallExpression) TokenLiteral() string {
	return call.Token.Literal
}
func (call *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range call.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(call.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}
func (sl *StringLiteral) String() string {
	return sl.Token.Literal
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}
func (al *ArrayLiteral) TokenLiteral() string {
	return al.Token.Literal
}
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) TokenLiteral() string {
	return ie.Token.Literal
}
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode() {}
func (hl *HashLiteral) TokenLiteral() string {
	return hl.Token.Literal
}
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	paris := []string{}
	for key, value := range hl.Pairs {
		paris = append(paris, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(paris, ","))
	out.WriteString("}")

	return out.String()
}

type MacroLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (m *MacroLiteral) expressionNode() {}
func (m *MacroLiteral) TokenLiteral() string {
	return m.Token.Literal
}
func (m *MacroLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(m.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	// out.WriteString("{")
	out.WriteString(m.Body.String())
	// out.WriteString("}")

	return out.String()
}
