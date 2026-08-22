package optimizer

import (
	"Tokype/ast"
	"Tokype/evalut"
)

func (o *Optimizer) constantFold(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantFold(stmt)
		}
		return n

	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			folded := o.constantFold(stmt)
			if folded != nil {
				newStatements = append(newStatements, folded)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.constantFold(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.AssignStatement:
		if n.Value != nil {
			n.Value = o.constantFoldExpression(n.Value)
			if folded := o.evaluateConstant(n.Value); folded != nil {
				n.Value = folded
			}
		}
		return n

	case *ast.InfixExpression:
		n.Left = o.constantFoldExpression(n.Left)
		n.Right = o.constantFoldExpression(n.Right)

		if folded := o.evaluateConstant(n); folded != nil {
			return folded
		}
		return n

	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "print", "input", "get_time", "len", "push", "pop", "first", "rest", "contains", "negate":
				return n
			}
		}

		for i, arg := range n.Arguments {
			n.Arguments[i] = o.constantFoldExpression(arg)
		}
		return n

	case *ast.IfExpression:
		n.Condition = o.constantFoldExpression(n.Condition)
		if n.Consequence != nil {
			n.Consequence = o.constantFold(n.Consequence).(*ast.BlockStatement)
		}
		for i, elif := range n.Alternative {
			elif.Condition = o.constantFoldExpression(elif.Condition)
			if elif.Consequence != nil {
				elif.Consequence = o.constantFold(elif.Consequence).(*ast.BlockStatement)
			}
			n.Alternative[i] = elif
		}
		if n.Else != nil {
			n.Else = o.constantFold(n.Else).(*ast.BlockStatement)
		}
		return n

	case *ast.ForStatement:
		if folded := o.foldArithmeticProgression(n); folded != nil {
			return folded
		}

		if n.Initialization != nil {
			n.Initialization = o.constantFold(n.Initialization).(*ast.AssignStatement)
		}
		if n.Condition != nil {
			n.Condition = o.constantFoldExpression(n.Condition)
		}
		if n.Update != nil {
			n.Update = o.constantFold(n.Update).(*ast.AssignStatement)
		}
		if n.Body != nil {
			n.Body = o.constantFold(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.WhileStatement:
		if n.Condition != nil {
			n.Condition = o.constantFoldExpression(n.Condition)
		}
		if n.Body != nil {
			n.Body = o.constantFold(n.Body).(*ast.BlockStatement)
		}
		return n

	default:
		return n
	}
}

func (o *Optimizer) constantFoldExpression(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}
	result := o.constantFold(expr)
	if result == nil {
		return nil
	}
	if exprResult, ok := result.(ast.Expression); ok {
		return exprResult
	}
	return expr
}

func (o *Optimizer) evaluateConstant(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	tempEnv := evalut.NewEnvironment()
	result := evalut.Eval(expr, tempEnv)

	if result.IsNil() {
		return nil
	}

	if o.hasVariables(expr) {
		return nil
	}

	return o.valueToLiteral(result)
}
