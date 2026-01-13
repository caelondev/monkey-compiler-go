package compiler

import "fmt"

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
	NativeScope SymbolScope = "NATIVE"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int // Symbol address
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Outer: nil,

		store:          make(map[string]Symbol),
		numDefinitions: 0,
	}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		Outer:          outer,
		store:          make(map[string]Symbol),
		numDefinitions: 0,
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

func (s *SymbolTable) Reassign(name string) (Symbol, error) {
	symbol, err := s.Resolve(name)
	if err != nil {
		return Symbol{}, fmt.Errorf("Cannot reassign undefined variable '%s'", name)
	}

	return symbol, nil
}

func (s *SymbolTable) Resolve(name string) (Symbol, error) {
	symbol, exists := s.store[name]

	if !exists {
		if s.Outer == nil {
			return Symbol{}, fmt.Errorf("Cannot resolve Identifier '%s' as it is undefined", name)
		} else {
			return s.Outer.Resolve(name)
		}
	}

	return symbol, nil
}
