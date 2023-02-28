package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION" //Function Name
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	FreeSymbol     []Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbol: free, numDefinitions: 0}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (sym *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: sym.numDefinitions}

	// 先查找符号，如果存在，就使用找到符号的索引, 有必要重写一个函数
	res, ok := sym.Resolve(name)
	if ok {
		symbol.Index = res.Index
		if sym.Outer != nil {
			symbol.Scope = res.Scope
		} else {
			symbol.Scope = res.Scope
		}
	} else {
		sym.numDefinitions += 1
		if sym.Outer == nil {
			symbol.Scope = GlobalScope
		} else {
			symbol.Scope = LocalScope
		}
	}

	sym.store[name] = symbol

	return symbol
}

func (sym *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, ok := sym.store[name]
	if !ok && sym.Outer != nil {
		symbol, ok = sym.Outer.Resolve(name)
		if !ok {
			return symbol, ok
		}

		if symbol.Scope == GlobalScope || symbol.Scope == BuiltinScope {
			return symbol, ok
		}

		free := sym.DefineFree(symbol)
		return free, ok
	}
	return symbol, ok
}

func (sym *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	sym.store[name] = symbol
	return symbol
}

func (sym *SymbolTable) DefineFree(original Symbol) Symbol {
	sym.FreeSymbol = append(sym.FreeSymbol, original)
	symbol := Symbol{Name: original.Name, Index: len(sym.FreeSymbol) - 1,
		Scope: FreeScope}

	sym.store[original.Name] = symbol
	return symbol
}

func (sym *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Scope: FunctionScope, Index: 0}
	sym.store[name] = symbol

	return symbol
}
