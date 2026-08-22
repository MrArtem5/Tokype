package optimizer

import (
	"Tokype/ast"
	"sync"
)

type Optimizer struct {
	env        interface{}
	cache      map[ast.Node]bool
	cacheMutex sync.RWMutex
	functions  map[string]*ast.FunctionStatement
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		cache:     make(map[ast.Node]bool),
		functions: make(map[string]*ast.FunctionStatement),
	}
}

func (o *Optimizer) Optimize(program *ast.Program) *ast.Program {
	o.cache = make(map[ast.Node]bool)
	o.functions = make(map[string]*ast.FunctionStatement)

	o.collectFunctions(program)

	o.constantFold(program)
	o.optimizeControlFlow(program)
	o.removeDeadCode(program)
	o.optimizeCollections(program)
	program = o.optimizeNestedLoops(program)
	program = o.inlineFunctionCalls(program)
	program = o.removeEmptyFunctions(program)

	return program
}

func (o *Optimizer) collectFunctions(node ast.Node) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			o.collectFunctions(stmt)
		}
	case *ast.FunctionStatement:
		if n.Name != nil {
			o.functions[n.Name.Value] = n
		}
	}
}

func (o *Optimizer) Reset() {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()
	o.cache = make(map[ast.Node]bool)
	o.functions = make(map[string]*ast.FunctionStatement)
}

func OptimizeAST(program *ast.Program) *ast.Program {
	optimizer := NewOptimizer()
	return optimizer.Optimize(program)
}
