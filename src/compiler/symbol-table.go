package compiler

import "fmt"

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	NativeScope   SymbolScope = "NATIVE"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int // Symbol address
}

type SymbolTable struct {
	Outer       *SymbolTable
	FreeSymbols []Symbol

	store           map[string]Symbol
	numDefinitions  int
	isFunctionScope bool
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Outer:           nil,
		FreeSymbols:     make([]Symbol, 0),
		store:           make(map[string]Symbol),
		numDefinitions:  0,
		isFunctionScope: true,
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		Outer:           outer,
		store:           make(map[string]Symbol),
		numDefinitions:  0,
		isFunctionScope: false,
	}
}

func (s *SymbolTable) Define(name string) (Symbol, error) {
	if _, exists := s.store[name]; exists {
		return Symbol{}, fmt.Errorf(
			"Cannot redeclare already existing Identifier '%s'", name,
		)
	}

	symbol := Symbol{
		Name:  name,
		Index: s.numDefinitions,
	}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++

	return symbol, nil
}

func (s *SymbolTable) DefineNative(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: NativeScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunction(name string) Symbol {
	// NOTE: Index number doesnt matter here
	symbol := Symbol{Name: name, Index: 9999, Scope: FunctionScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{
		Name:  original.Name,
		Index: len(s.FreeSymbols) - 1,
		Scope: FreeScope,
	}
	s.store[original.Name] = symbol
	return symbol
}

// NOTE: This is just a validator/resolve wrapper
// I might change this in the future
func (s *SymbolTable) Reassign(name string) (Symbol, error) {
	symbol, err := s.Resolve(name)
	if err != nil {
		return Symbol{}, fmt.Errorf("Cannot reassign undefined variable '%s'", name)
	}

	return symbol, nil
}

func (s *SymbolTable) Resolve(name string) (Symbol, error) {
	symbol, exists := s.store[name]

	if exists {
		return symbol, nil
	}

	if s.Outer == nil {
		return Symbol{}, fmt.Errorf("Cannot resolve identifier '%s'", name)
	}

	symbol, err := s.Outer.Resolve(name)
	if err != nil {
		return Symbol{}, err
	}

	if symbol.Scope == FunctionScope {
		return symbol, nil
	}

	if symbol.Scope == GlobalScope || symbol.Scope == NativeScope {
		return symbol, nil
	}

	if symbol.Scope == LocalScope && !s.isFunctionScope {
		return symbol, nil
	}

	// Only set function scopes as free scope ---
	return s.DefineFree(symbol), nil
}
