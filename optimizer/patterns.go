package optimizer

import (
	"Tokype/ast"
)

func (o *Optimizer) foldArithmeticProgression(forStmt *ast.ForStatement) ast.Node {
	if forStmt.Body == nil || len(forStmt.Body.Statements) != 1 {
		return nil
	}

	assign, ok := forStmt.Body.Statements[0].(*ast.AssignStatement)
	if !ok {
		return nil
	}

	infix, ok := assign.Value.(*ast.InfixExpression)
	if !ok || infix.Operator != "+" {
		return nil
	}

	leftIdent, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 {
		return nil
	}

	if leftIdent.Value != assign.Name.Value {
		return nil
	}

	rightIdent, ok2 := infix.Right.(*ast.Identifier)
	if !ok2 {
		return o.foldSimpleArithmetic(forStmt)
	}

	if rightIdent.Value != forStmt.Initialization.Name.Value {
		return nil
	}

	startVal, ok := forStmt.Initialization.Value.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	endVal, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return nil
	}

	endInt, ok := endVal.Right.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	count := endInt.Value - startVal.Value + 1
	if count <= 0 {
		return nil
	}

	sum := count * (startVal.Value + endInt.Value) / 2

	return &ast.AssignStatement{
		Name:  assign.Name,
		Value: &ast.IntegerLiteral{Value: sum},
	}
}

func (o *Optimizer) foldSimpleArithmetic(forStmt *ast.ForStatement) ast.Node {
	if forStmt.Body == nil || len(forStmt.Body.Statements) != 1 {
		return nil
	}

	assign, ok := forStmt.Body.Statements[0].(*ast.AssignStatement)
	if !ok {
		return nil
	}

	infix, ok := assign.Value.(*ast.InfixExpression)
	if !ok || infix.Operator != "+" {
		return nil
	}

	leftIdent, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 {
		return nil
	}

	if leftIdent.Value != assign.Name.Value {
		return nil
	}

	_, ok2 := infix.Right.(*ast.IntegerLiteral)
	if !ok2 {
		return nil
	}

	startVal, ok := forStmt.Initialization.Value.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	endVal, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return nil
	}

	endInt, ok := endVal.Right.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	count := endInt.Value - startVal.Value + 1
	if count <= 0 {
		return nil
	}

	return &ast.AssignStatement{
		Name:  assign.Name,
		Value: &ast.IntegerLiteral{Value: count},
	}
}

func (o *Optimizer) optimizeNestedLoops(program *ast.Program) *ast.Program {
	if program == nil {
		return program
	}

	newStatements := []ast.Node{}
	for _, stmt := range program.Statements {
		newStmt := o.optimizeNestedLoopsInNode(stmt)
		if newStmt != nil {
			newStatements = append(newStatements, newStmt)
		}
	}
	program.Statements = newStatements
	return program
}

func (o *Optimizer) optimizeNestedLoopsInNode(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.optimizeNestedLoopsInNode(stmt)
		}
		return n
	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			newStmt := o.optimizeNestedLoopsInNode(stmt)
			if newStmt != nil {
				newStatements = append(newStatements, newStmt)
			}
		}
		n.Statements = newStatements
		return n
	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.optimizeNestedLoopsInNode(n.Body).(*ast.BlockStatement)
		}
		return n
	case *ast.ForStatement:
		return o.mergeNestedLoop(n)
	default:
		return n
	}
}

func (o *Optimizer) mergeNestedLoop(outerLoop *ast.ForStatement) ast.Node {
	if outerLoop.Body == nil || len(outerLoop.Body.Statements) != 1 {
		return outerLoop
	}

	innerLoop, ok := outerLoop.Body.Statements[0].(*ast.ForStatement)
	if !ok {
		return outerLoop
	}

	if innerLoop.Body == nil {
		return outerLoop
	}

	if len(innerLoop.Body.Statements) != 1 {
		return outerLoop
	}

	outerStart, ok1 := outerLoop.Initialization.Value.(*ast.IntegerLiteral)
	if !ok1 {
		return outerLoop
	}
	outerEnd, ok2 := outerLoop.Condition.(*ast.InfixExpression)
	if !ok2 {
		return outerLoop
	}
	outerEndInt, ok3 := outerEnd.Right.(*ast.IntegerLiteral)
	if !ok3 {
		return outerLoop
	}

	innerStart, ok4 := innerLoop.Initialization.Value.(*ast.IntegerLiteral)
	if !ok4 {
		return outerLoop
	}
	innerEnd, ok5 := innerLoop.Condition.(*ast.InfixExpression)
	if !ok5 {
		return outerLoop
	}
	innerEndInt, ok6 := innerEnd.Right.(*ast.IntegerLiteral)
	if !ok6 {
		return outerLoop
	}

	outerCount := outerEndInt.Value - outerStart.Value + 1
	if outerCount <= 0 {
		return outerLoop
	}
	innerCount := innerEndInt.Value - innerStart.Value + 1
	if innerCount <= 0 {
		return outerLoop
	}

	totalCount := outerCount * innerCount

	newStart := &ast.IntegerLiteral{Value: 1}
	newEnd := &ast.IntegerLiteral{Value: totalCount}

	newInit := &ast.AssignStatement{
		Name:  innerLoop.Initialization.Name,
		Value: newStart,
	}

	newCondition := &ast.InfixExpression{
		Left:     &ast.Identifier{Value: innerLoop.Initialization.Name.Value},
		Operator: "<=",
		Right:    newEnd,
	}

	innerLoop.Initialization = newInit
	innerLoop.Condition = newCondition

	return innerLoop
}

func (o *Optimizer) inlineFunctionCalls(program *ast.Program) *ast.Program {
	if program == nil {
		return program
	}

	newStatements := []ast.Node{}
	functionCalls := make(map[string][]*ast.CallExpression)
	functionNames := make(map[string]bool)

	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Name != nil {
				functionNames[fnStmt.Name.Value] = true
			}
		}
		o.collectCalls(stmt, functionCalls)
	}

	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Name != nil {
				if calls, exists := functionCalls[fnStmt.Name.Value]; exists {
					for _, call := range calls {
						inlined := o.inlineFunctionCall(call, fnStmt)
						if inlined != nil {
							newStatements = append(newStatements, inlined)
						}
					}
				}
			}
		} else {
			if call, ok := stmt.(*ast.CallExpression); ok {
				if ident, ok := call.Function.(*ast.Identifier); ok {
					if functionNames[ident.Value] {
						continue
					}
				}
			}
			newStatements = append(newStatements, stmt)
		}
	}

	program.Statements = newStatements
	return program
}

func (o *Optimizer) collectCalls(node ast.Node, calls map[string][]*ast.CallExpression) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			o.collectCalls(stmt, calls)
		}
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			o.collectCalls(stmt, calls)
		}
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if _, exists := calls[ident.Value]; !exists {
				calls[ident.Value] = []*ast.CallExpression{}
			}
			calls[ident.Value] = append(calls[ident.Value], n)
		}
	}
}

func (o *Optimizer) inlineFunctionCall(call *ast.CallExpression, fnStmt *ast.FunctionStatement) ast.Node {
	if fnStmt.Body == nil {
		return call
	}

	if len(fnStmt.Body.Statements) == 0 {
		return &ast.BlockStatement{Statements: []ast.Node{}}
	}

	if len(fnStmt.Body.Statements) == 1 {
		stmt := fnStmt.Body.Statements[0]
		return o.replaceInStatement(stmt, fnStmt.Parameters, call.Arguments)
	}

	newBody := &ast.BlockStatement{
		Statements: []ast.Node{},
	}

	for _, stmt := range fnStmt.Body.Statements {
		newStmt := o.replaceInStatement(stmt, fnStmt.Parameters, call.Arguments)
		if newStmt != nil {
			newBody.Statements = append(newBody.Statements, newStmt)
		}
	}

	return newBody
}

func (o *Optimizer) replaceInStatement(stmt ast.Node, params []*ast.Identifier, args []ast.Expression) ast.Node {
	switch s := stmt.(type) {
	case *ast.AssignStatement:
		return &ast.AssignStatement{
			Name:  s.Name,
			Value: o.replaceParameters(s.Value, params, args),
		}
	case *ast.CallExpression:
		newCall := &ast.CallExpression{
			Function:  o.replaceParameters(s.Function, params, args),
			Arguments: []ast.Expression{},
		}
		for _, arg := range s.Arguments {
			newCall.Arguments = append(newCall.Arguments, o.replaceParameters(arg, params, args))
		}
		return newCall
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Left:     o.replaceParameters(s.Left, params, args),
			Operator: s.Operator,
			Right:    o.replaceParameters(s.Right, params, args),
		}
	default:
		return stmt
	}
}

func (o *Optimizer) replaceParameters(expr ast.Expression, params []*ast.Identifier, args []ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		for i, param := range params {
			if param.Value == e.Value && i < len(args) {
				return args[i]
			}
		}
		return e
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Left:     o.replaceParameters(e.Left, params, args),
			Operator: e.Operator,
			Right:    o.replaceParameters(e.Right, params, args),
		}
	case *ast.CallExpression:
		newCall := &ast.CallExpression{
			Function:  o.replaceParameters(e.Function, params, args),
			Arguments: []ast.Expression{},
		}
		for _, arg := range e.Arguments {
			newCall.Arguments = append(newCall.Arguments, o.replaceParameters(arg, params, args))
		}
		return newCall
	default:
		return e
	}
}

func (o *Optimizer) removeEmptyFunctions(program *ast.Program) *ast.Program {
	newStatements := []ast.Node{}
	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Body != nil && len(fnStmt.Body.Statements) > 0 {
				newStatements = append(newStatements, stmt)
			}
		} else {
			newStatements = append(newStatements, stmt)
		}
	}
	program.Statements = newStatements
	return program
}
