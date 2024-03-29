package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"monkey/code"
	"strings"
)

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() string
}

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING                = "STRING"
	BUILTIN               = "BUILTIN"
	ARRAY_OBJ             = "ARRAY_OBJ"
	HASH_OBJ              = "HASH"
	QUOTE_OBJ             = "QUOTE" //宏
	MACRO_OBJ             = "MACRO"
	COMPILER_FUNCTION_OBJ = "COMPILER_FUNCTION_OBJ"
	CLOSURE_OBJ           = "CLOSURE_OBJ"
	CLASS_OBJ             = "CLASS_OBJ"
)

// 值系统
//数据类型： 空值、boolean、整数、函数字面量

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}
func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%v", b.Value)
}

type Null struct{}

func (n *Null) Inspect() string {
	return "null"
}
func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}
func (e *Error) Inspect() string {
	return "ERROR:" + e.Message
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	// fmt.Printf("st (%+v)\n", obj)
	return obj, ok
}
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type String struct {
	Value string
}

func (s *String) Inspect() string {
	return s.Value
}
func (s *String) Type() ObjectType {
	return STRING
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType {
	return BUILTIN
}
func (b *Builtin) Inspect() string {
	return "builtin function"
}

type Array struct {
	ELements []Object
}

func (a *Array) Type() ObjectType {
	return ARRAY_OBJ
}
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range a.ELements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type HashAble interface {
	HashKey() HashKey
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType {
	return HASH_OBJ
}
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, pair.Key.Inspect()+":"+pair.Value.Inspect())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type HashPair struct {
	Key   Object
	Value Object
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type Quote struct {
	Node ast.Node
}

func (q *Quote) Type() ObjectType {
	return QUOTE_OBJ
}

func (q *Quote) Inspect() string {
	return "QUOTE(" + q.Node.String() + ")"
}

type Macro struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (m *Macro) Type() ObjectType {
	return MACRO_OBJ
}
func (m *Macro) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(m.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type CompiledFunction struct {
	Instructions  code.Instruction
	NumLocals     int
	NumParameters int
	Name          string
}

func (cf *CompiledFunction) Type() ObjectType {
	return COMPILER_FUNCTION_OBJ
}
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type Closure struct {
	Fn      *CompiledFunction
	FreeVar []Object
}

func (cl *Closure) Type() ObjectType {
	return CLOSURE_OBJ
}
func (cl *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", cl)
}

type Class struct {
	Name    string
	Methods map[string]Object
	Fields  map[string]Object
}

func (class *Class) Type() ObjectType {
	return CLASS_OBJ
}
func (class *Class) Inspect() string {
	return fmt.Sprintf("Class[%p]", class)
}
