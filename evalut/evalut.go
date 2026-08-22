package evalut

import (
	"Tokype/ast"
	"Tokype/value"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var ProgramStartTime = time.Now()

type Environment struct {
	store map[string]value.Value
	cache map[string]int
	data  []value.Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]value.Value),
		cache: make(map[string]int),
		data:  make([]value.Value, 0, 16),
		outer: nil,
	}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Set(name string, val value.Value) value.Value {
	if id, ok := e.cache[name]; ok {
		e.data[id] = val
		e.store[name] = val
		return val
	}

	id := len(e.data)
	e.cache[name] = id
	e.data = append(e.data, val)
	e.store[name] = val
	return val
}

func (e *Environment) Get(name string) (value.Value, bool) {
	if id, ok := e.cache[name]; ok {
		if id < len(e.data) {
			return e.data[id], true
		}
	}

	val, ok := e.store[name]
	if ok {
		id := len(e.data)
		e.cache[name] = id
		e.data = append(e.data, val)
		return val, true
	}

	if e.outer != nil {
		return e.outer.Get(name)
	}

	return value.NewNilValue(), false
}

type ReturnSignal struct {
	Value value.Value
}

var FunctionStore = make(map[string]*value.FunctionObject)
var Reader = bufio.NewReader(os.Stdin)

func isNumber(v value.Value) bool {
	return v.Type == value.ValInt || v.Type == value.ValFloat
}

func evalNumberOperation(left, right value.Value, operator string) value.Value {
	leftVal, ok1 := left.ToFloat64()
	rightVal, ok2 := right.ToFloat64()

	if !ok1 || !ok2 {
		return value.NewNilValue()
	}

	var result float64

	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			fmt.Println("Error: division by zero")
			return value.NewNilValue()
		}
		result = leftVal / rightVal
	case "%":
		if rightVal == 0 {
			fmt.Println("Error: modulo by zero")
			return value.NewNilValue()
		}
		if left.Type == value.ValInt && right.Type == value.ValInt {
			return value.NewIntValue(int64(leftVal) % int64(rightVal))
		}
		result = float64(int64(leftVal) % int64(rightVal))
	default:
		return value.NewNilValue()
	}

	if result == float64(int64(result)) {
		return value.NewIntValue(int64(result))
	}
	return value.NewFloatValue(result)
}

func evalNumberComparison(left, right value.Value, operator string) value.Value {
	leftVal, ok1 := left.ToFloat64()
	rightVal, ok2 := right.ToFloat64()

	if !ok1 || !ok2 {
		return value.NewNilValue()
	}

	switch operator {
	case "<":
		return value.NewBoolValue(leftVal < rightVal)
	case ">":
		return value.NewBoolValue(leftVal > rightVal)
	case "<=":
		return value.NewBoolValue(leftVal <= rightVal)
	case ">=":
		return value.NewBoolValue(leftVal >= rightVal)
	default:
		return value.NewNilValue()
	}
}

func Eval(node ast.Node, env *Environment) (result value.Value) {
	defer func() {
		if r := recover(); r != nil {
			if signal, ok := r.(ReturnSignal); ok {
				result = signal.Value
				return
			}
			panic(r)
		}
	}()

	switch node := node.(type) {
	case *ast.Program:
		var result value.Value = value.NewNilValue()
		for _, statement := range node.Statements {
			if _, ok := statement.(*ast.FunctionStatement); ok {
				Eval(statement, env)
			}
		}
		for _, statement := range node.Statements {
			if _, ok := statement.(*ast.FunctionStatement); !ok {
				result = Eval(statement, env)
			}
		}
		return result

	case *ast.BlockStatement:
		var result value.Value = value.NewNilValue()
		for _, statement := range node.Statements {
			result = Eval(statement, env)
		}
		return result

	case *ast.ReturnStatement:
		var val value.Value = value.NewNilValue()
		if node.Value != nil {
			val = Eval(node.Value, env)
		}
		panic(ReturnSignal{Value: val})

	case *ast.IntegerLiteral:
		return value.NewIntValue(node.Value)

	case *ast.FloatLiteral:
		return value.NewFloatValue(node.Value)

	case *ast.StringLiteral:
		return value.NewStringValue(node.Value)

	case *ast.BooleanLiteral:
		return value.NewBoolValue(node.Value)

	case *ast.ListLiteral:
		elements := []value.Value{}
		for _, el := range node.Elements {
			evaluated := Eval(el, env)
			elements = append(elements, evaluated)
		}
		return value.NewListValue(&value.ListObject{Elements: elements})

	case *ast.MapLiteral:
		pairs := make(map[value.Value]value.Value)
		for keyNode, valueNode := range node.Pairs {
			key := Eval(keyNode, env)
			val := Eval(valueNode, env)
			pairs[key] = val
		}
		return value.NewMapValue(&value.MapObject{Elements: pairs})

	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			if fn, ok := FunctionStore[node.Value]; ok {
				return value.NewFunctionValue(fn)
			}
			return value.NewNilValue()
		}
		return val

	case *ast.InfixExpression:
		leftVal := Eval(node.Left, env)
		rightVal := Eval(node.Right, env)

		if leftVal.IsNil() || rightVal.IsNil() {
			return value.NewNilValue()
		}

		switch node.Operator {
		case "+":
			if leftVal.Type == value.ValString {
				return value.NewStringValue(leftVal.Str + rightVal.String())
			}
			if rightVal.Type == value.ValString {
				return value.NewStringValue(leftVal.String() + rightVal.Str)
			}
			return evalNumberOperation(leftVal, rightVal, "+")
		case "-":
			return evalNumberOperation(leftVal, rightVal, "-")
		case "*":
			return evalNumberOperation(leftVal, rightVal, "*")
		case "/":
			return evalNumberOperation(leftVal, rightVal, "/")
		case "%":
			return evalNumberOperation(leftVal, rightVal, "%")
		case "==":
			if isNumber(leftVal) && isNumber(rightVal) {
				l, _ := leftVal.ToFloat64()
				r, _ := rightVal.ToFloat64()
				return value.NewBoolValue(l == r)
			}
			return value.NewBoolValue(leftVal == rightVal)
		case "?=":
			if isNumber(leftVal) && isNumber(rightVal) {
				l, _ := leftVal.ToFloat64()
				r, _ := rightVal.ToFloat64()
				return value.NewBoolValue(l != r)
			}
			return value.NewBoolValue(leftVal != rightVal)
		case "<", ">", "<=", ">=":
			return evalNumberComparison(leftVal, rightVal, node.Operator)
		default:
			return value.NewNilValue()
		}

	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		env.Set(node.Name.Value, val)
		return val

	case *ast.FunctionStatement:
		fn := &value.FunctionObject{
			Parameters: make([]interface{}, len(node.Parameters)),
			Body:       node.Body,
			Env:        env,
		}
		for i, p := range node.Parameters {
			fn.Parameters[i] = p.Value
		}
		FunctionStore[node.Name.Value] = fn
		return value.NewFunctionValue(fn)

	case *ast.ForStatement:
		Eval(node.Initialization, env)

		for {
			cond := Eval(node.Condition, env)
			if cond.IsNil() {
				break
			}
			if cond.Type == value.ValBool && !cond.Bool {
				break
			}

			Eval(node.Body, env)
			Eval(node.Update, env)
		}
		return value.NewNilValue()

	case *ast.WhileStatement:
		for {
			cond := Eval(node.Condition, env)
			if cond.IsNil() {
				break
			}
			if cond.Type == value.ValBool && !cond.Bool {
				break
			}
			Eval(node.Body, env)
		}
		return value.NewNilValue()

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}
		index := Eval(node.Index, env)
		if index.IsNil() {
			return value.NewNilValue()
		}

		if left.Type == value.ValRepeat {
			repeatObj, ok := left.Obj.(*value.RepeatObject)
			if !ok {
				fmt.Println("Error: invalid repeat object")
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Println("Error: index must be a number")
				return value.NewNilValue()
			}
			if idx < 0 || idx >= repeatObj.Count {
				fmt.Printf("Error: index %d is out of bounds (length: %d)\n", idx, repeatObj.Count)
				return value.NewNilValue()
			}
			return repeatObj.Value
		}

		if left.Type == value.ValList {
			list, ok := left.Obj.(*value.ListObject)
			if !ok {
				fmt.Println("Error: invalid list object")
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Println("Error: index must be a number")
				return value.NewNilValue()
			}
			if idx < 0 || idx >= int64(len(list.Elements)) {
				fmt.Printf("Error: index %d is out of list bounds (length: %d)\n", idx, len(list.Elements))
				return value.NewNilValue()
			}
			return list.Elements[idx]
		}

		if left.Type == value.ValMap {
			mapObj, ok := left.Obj.(*value.MapObject)
			if !ok {
				fmt.Println("Error: invalid map object")
				return value.NewNilValue()
			}
			val, ok := mapObj.Elements[index]
			if !ok {
				fmt.Printf("Error: key %v not found in the dictionary\n", index)
				return value.NewNilValue()
			}
			return val
		}

		fmt.Printf("Error: indexing is not supported for the type %s\n", left.Type.String())
		return value.NewNilValue()

	case *ast.AssignIndexStatement:
		left := Eval(node.Left.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}
		index := Eval(node.Left.Index, env)
		if index.IsNil() {
			return value.NewNilValue()
		}
		val := Eval(node.Value, env)
		if val.IsNil() {
			return value.NewNilValue()
		}

		if left.Type == value.ValRepeat {
			fmt.Println("Error: cannot assign to repeat object (read-only)")
			return value.NewNilValue()
		}

		if left.Type == value.ValList {
			list, ok := left.Obj.(*value.ListObject)
			if !ok {
				fmt.Println("Error: invalid list object")
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Println("Error: index must be a number")
				return value.NewNilValue()
			}
			if idx < 0 || idx >= int64(len(list.Elements)) {
				fmt.Printf("Error: index %d out of bounds\n", idx)
				return value.NewNilValue()
			}
			list.Elements[idx] = val
			return val
		}

		if left.Type == value.ValMap {
			mapObj, ok := left.Obj.(*value.MapObject)
			if !ok {
				fmt.Println("Error: invalid map object")
				return value.NewNilValue()
			}
			mapObj.Elements[index] = val
			return val
		}

		fmt.Printf("Error: index assignment not supported for %s\n", left.Type.String())
		return value.NewNilValue()

	case *ast.SliceExpression:
		left := Eval(node.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}

		if left.Type != value.ValList {
			fmt.Printf("Error: slicing works only with lists; got %s\n", left.Type.String())
			return value.NewNilValue()
		}

		list, ok := left.Obj.(*value.ListObject)
		if !ok {
			fmt.Println("Error: invalid list object")
			return value.NewNilValue()
		}

		start := 0
		end := len(list.Elements)

		if node.Start != nil {
			startVal := Eval(node.Start, env)
			if startVal.IsNil() {
				return value.NewNilValue()
			}
			s, ok := startVal.ToInt64()
			if !ok {
				fmt.Println("Error: slice start must be a number")
				return value.NewNilValue()
			}
			start = int(s)
		}

		if node.End != nil {
			endVal := Eval(node.End, env)
			if endVal.IsNil() {
				return value.NewNilValue()
			}
			e, ok := endVal.ToInt64()
			if !ok {
				fmt.Println("Error: slice end must be a number")
				return value.NewNilValue()
			}
			end = int(e)
		}

		if start < 0 {
			start = 0
		}
		if end > len(list.Elements) {
			end = len(list.Elements)
		}
		if start > end {
			return value.NewListValue(&value.ListObject{Elements: []value.Value{}})
		}

		return value.NewListValue(&value.ListObject{
			Elements: list.Elements[start:end],
		})

	case *ast.IfExpression:
		condition := Eval(node.Condition, env)
		if condition.IsNil() {
			return value.NewNilValue()
		}

		if condition.Type == value.ValBool && condition.Bool {
			return Eval(node.Consequence, env)
		}

		for _, elif := range node.Alternative {
			elifCond := Eval(elif.Condition, env)
			if elifCond.IsNil() {
				continue
			}
			if elifCond.Type == value.ValBool && elifCond.Bool {
				return Eval(elif.Consequence, env)
			}
		}

		if node.Else != nil {
			return Eval(node.Else, env)
		}

		return value.NewNilValue()

	case *ast.CallExpression:
		var funcName string
		var fnObj *value.FunctionObject

		switch fn := node.Function.(type) {
		case *ast.Identifier:
			funcName = fn.Value

			if funcName == "repeat" {
				if len(node.Arguments) != 2 {
					fmt.Println("Error: repeat() requires 2 arguments (value, count)")
					return value.NewNilValue()
				}

				val := Eval(node.Arguments[0], env)
				countVal := Eval(node.Arguments[1], env)

				if val.IsNil() || countVal.IsNil() {
					return value.NewNilValue()
				}

				count, ok := countVal.ToInt64()
				if !ok {
					fmt.Println("Error: count must be a number")
					return value.NewNilValue()
				}

				if count < 0 || count > 1000000000 {
					fmt.Printf("Error: count %d is too large (max 1,000,000,000)\n", count)
					return value.NewNilValue()
				}

				return value.Value{
					Type: value.ValRepeat,
					Obj:  &value.RepeatObject{Value: val, Count: count},
				}
			}

			if funcName == "reserve" {
				if len(node.Arguments) != 2 {
					fmt.Println("Error: reserve() requires 2 arguments (list, capacity)")
					return value.NewNilValue()
				}

				listArg := Eval(node.Arguments[0], env)

				var list *value.ListObject
				if listArg.IsNil() || listArg.Type != value.ValList {
					list = &value.ListObject{
						Elements: []value.Value{},
					}
				} else {
					var ok bool
					list, ok = listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
				}

				capacityVal := Eval(node.Arguments[1], env)
				if capacityVal.IsNil() {
					fmt.Println("Error: capacity is nil")
					return value.NewNilValue()
				}

				capacity, ok := capacityVal.ToInt64()
				if !ok {
					fmt.Println("Error: capacity must be a number")
					return value.NewNilValue()
				}

				if capacity < 0 {
					fmt.Println("Error: capacity cannot be negative")
					return value.NewNilValue()
				}

				if capacity > 100000000 {
					fmt.Printf("Error: capacity %d is too large (max 100,000,000)\n", capacity)
					return value.NewNilValue()
				}

				newList := &value.ListObject{
					Elements: make([]value.Value, 0, capacity),
				}

				if list != nil {
					newList.Elements = append(newList.Elements, list.Elements...)
				}

				return value.NewListValue(newList)
			}

			if funcName == "print" {
				for i, arg := range node.Arguments {
					val := Eval(arg, env)
					if i > 0 {
						fmt.Print(" ")
					}
					if !val.IsNil() {
						fmt.Print(val.String())
					}
				}
				fmt.Println()
				return value.NewNilValue()
			}

			if funcName == "input" {
				if len(node.Arguments) > 0 {
					prompt := Eval(node.Arguments[0], env)
					if !prompt.IsNil() {
						fmt.Print(prompt.String())
					}
				}
				input, _ := Reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if num, err := strconv.ParseFloat(input, 64); err == nil {
					if num == float64(int64(num)) {
						return value.NewIntValue(int64(num))
					}
					return value.NewFloatValue(num)
				}
				if num, err := strconv.ParseInt(input, 10, 64); err == nil {
					return value.NewIntValue(num)
				}
				return value.NewStringValue(input)
			}

			if funcName == "negate" {
				if len(node.Arguments) != 1 {
					fmt.Println("Error: negate() requires exactly 1 argument")
					return value.NewNilValue()
				}
				arg := Eval(node.Arguments[0], env)
				if arg.IsNil() {
					return value.NewNilValue()
				}
				if arg.Type == value.ValInt {
					return value.NewIntValue(-arg.Int)
				}
				if arg.Type == value.ValFloat {
					return value.NewFloatValue(-arg.Float)
				}
				fmt.Printf("Error: negate() works only with numbers, got %s\n", arg.Type.String())
				return value.NewNilValue()
			}

			if funcName == "len" {
				if len(node.Arguments) != 1 {
					fmt.Println("Error: len requires 1 argument")
					return value.NewNilValue()
				}
				arg := Eval(node.Arguments[0], env)
				if arg.IsNil() {
					return value.NewNilValue()
				}
				switch arg.Type {
				case value.ValList:
					list, ok := arg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					return value.NewIntValue(int64(len(list.Elements)))
				case value.ValString:
					return value.NewIntValue(int64(len(arg.Str)))
				default:
					fmt.Printf("Error: len() is not supported for the type %s\n", arg.Type.String())
					return value.NewNilValue()
				}
			}

			if funcName == "push" {
				if len(node.Arguments) != 2 {
					fmt.Println("Error: push() requires 2 arguments (a list and an element)")
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				elemArg := Eval(node.Arguments[1], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					list.Elements = append(list.Elements, elemArg)
					return listArg
				}
				fmt.Println("Error: push() works only with lists.")
				return value.NewNilValue()
			}

			if funcName == "pop" {
				if len(node.Arguments) != 1 {
					fmt.Println("Error: pop() requires 1 argument (a list)")
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						fmt.Println("Error: pop() on an empty list")
						return value.NewNilValue()
					}
					newList := &value.ListObject{
						Elements: list.Elements[:len(list.Elements)-1],
					}
					return value.NewListValue(newList)
				}
				fmt.Println("Error: pop() works only with lists.")
				return value.NewNilValue()
			}

			if funcName == "first" {
				if len(node.Arguments) != 1 {
					fmt.Println("Error: first() requires 1 argument (a list)")
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						return value.NewNilValue()
					}
					return list.Elements[0]
				}
				fmt.Println("Error: first() works only with lists")
				return value.NewNilValue()
			}

			if funcName == "rest" {
				if len(node.Arguments) != 1 {
					fmt.Println("Error: rest() requires 1 argument (a list)")
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						return value.NewListValue(&value.ListObject{Elements: []value.Value{}})
					}
					return value.NewListValue(&value.ListObject{Elements: list.Elements[1:]})
				}
				fmt.Println("Error: rest() works only with lists")
				return value.NewNilValue()
			}

			if funcName == "contains" {
				if len(node.Arguments) != 2 {
					fmt.Println("Error: contains() requires 2 arguments (list and element)")
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				elemArg := Eval(node.Arguments[1], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Println("Error: invalid list object")
						return value.NewNilValue()
					}
					for _, el := range list.Elements {
						if el == elemArg {
							return value.NewBoolValue(true)
						}
					}
					return value.NewBoolValue(false)
				}
				fmt.Println("Error: contains() works only with lists")
				return value.NewNilValue()
			}

			if funcName == "get_time" {
				if len(node.Arguments) != 0 {
					fmt.Println("Error: arguments cannot be passed")
					return value.NewNilValue()
				}

				duration := time.Since(ProgramStartTime)

				seconds := duration.Seconds()

				return value.NewFloatValue(seconds)
			}

			var ok bool
			fnObj, ok = FunctionStore[funcName]
			if !ok {
				fmt.Printf("Error: function '%s' not found\n", funcName)
				return value.NewNilValue()
			}

		default:
			fnObjVal := Eval(node.Function, env)
			if fnObjVal.IsNil() {
				fmt.Println("Error: cannot call non-function expression")
				return value.NewNilValue()
			}
			if fnObjVal.Type != value.ValFunction {
				fmt.Printf("Error: cannot call non-function type %s\n", fnObjVal.Type.String())
				return value.NewNilValue()
			}
			var ok bool
			fnObj, ok = fnObjVal.Obj.(*value.FunctionObject)
			if !ok {
				fmt.Println("Error: invalid function object")
				return value.NewNilValue()
			}
		}

		if fnObj == nil {
			fmt.Println("Error: function object is nil")
			return value.NewNilValue()
		}

		if len(node.Arguments) != len(fnObj.Parameters) {
			fmt.Printf("Error: function expects %d arguments, got %d\n",
				len(fnObj.Parameters), len(node.Arguments))
			return value.NewNilValue()
		}

		args := []value.Value{}
		for _, arg := range node.Arguments {
			args = append(args, Eval(arg, env))
		}

		extendedEnv := NewEnclosedEnvironment(fnObj.Env.(*Environment))

		for paramIdx, param := range fnObj.Parameters {
			if paramIdx < len(args) {
				if paramStr, ok := param.(string); ok {
					extendedEnv.Set(paramStr, args[paramIdx])
				}
			}
		}

		return Eval(fnObj.Body.(*ast.BlockStatement), extendedEnv)
	}

	return value.NewNilValue()
}

var stringPool = make(map[string]string)

func Intern(s string) string {
	if cached, ok := stringPool[s]; ok {
		return cached
	}
	copied := string([]byte(s))
	stringPool[s] = copied
	return copied
}
