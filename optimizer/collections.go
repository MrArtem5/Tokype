package optimizer

import (
	"Tokype/ast"
)

func (o *Optimizer) optimizeCollections(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.optimizeCollections(stmt)
		}
		return n

	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			optimized := o.optimizeCollections(stmt)
			if optimized != nil {
				newStatements = append(newStatements, optimized)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.optimizeCollections(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.ForStatement:
		return o.optimizeListCreation(n)

	default:
		return n
	}
}

func (o *Optimizer) optimizeListCreation(forStmt *ast.ForStatement) ast.Node {

	if forStmt.Body == nil || len(forStmt.Body.Statements) != 1 {
		return forStmt
	}

	stmt := forStmt.Body.Statements[0]
	assign, ok := stmt.(*ast.AssignStatement)
	if !ok {
		return forStmt
	}

	call, ok := assign.Value.(*ast.CallExpression)
	if !ok {
		return forStmt
	}

	ident, ok := call.Function.(*ast.Identifier)
	if !ok || ident.Value != "push" {
		return forStmt
	}

	if len(call.Arguments) != 2 {
		return forStmt
	}

	listIdent, ok := call.Arguments[0].(*ast.Identifier)
	if !ok {
		return forStmt
	}

	if listIdent.Value != assign.Name.Value {
		return forStmt
	}

	startVal, ok := forStmt.Initialization.Value.(*ast.IntegerLiteral)
	if !ok {
		return forStmt
	}

	endVal, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return forStmt
	}

	endInt, ok := endVal.Right.(*ast.IntegerLiteral)
	if !ok {
		return forStmt
	}

	count := endInt.Value - startVal.Value + 1

	value := call.Arguments[1]

	if _, ok := value.(*ast.IntegerLiteral); !ok {
		return forStmt
	}

	return &ast.AssignStatement{
		Name: listIdent,
		Value: &ast.CallExpression{
			Function: &ast.Identifier{Value: "repeat"},
			Arguments: []ast.Expression{
				value,
				&ast.IntegerLiteral{Value: count},
			},
		},
	}
}
