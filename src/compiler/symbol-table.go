package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int // Symbol address
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

func (s *SymbolTable) Define(name string) (Symbol, bool) {
	_, exists := s.Resolve(name)
	if exists {
		return Symbol{}, exists
	}

	symbol := Symbol{Name: name, Scope: GlobalScope, Index: s.numDefinitions}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol, false
}


func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	symbol, exists := s.store[name]
	return symbol, exists
}
