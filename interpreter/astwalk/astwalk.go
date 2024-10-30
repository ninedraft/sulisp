package astwalk

import (
	"fmt"
	"iter"

	"github.com/ninedraft/sulisp/internal/seq"
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/object"
)

var (
	Null  = object.Null{}
	True  = object.PrimitiveOf(true)
	False = object.PrimitiveOf(false)
)

func DefaultEnv() *object.Env {
	env := object.NewEnv()

	env.Assign("assign", newBuiltin(assign))
	env.Assign("apply", newBuiltin(builtinApply))
	env.Assign("namespace", newBuiltin(createNamespace))
	env.Assign(">", newBuiltin(gt))
	env.Assign("array", newBuiltin(func(sexp *ast.SExp, env *object.Env, eval object.Eval) object.Object {
		elements, err := seq.CollectErr(resolveMany(sexp.Items, env, eval))
		if err != nil {
			return fmtError(sexp.PosRange, "array values: %w", err)
		}

		return &object.Array{
			Elements: elements,
		}
	}))
	return env
}

func newBuiltin(fn object.BuiltinFn) *object.Builtin {
	return &object.Builtin{
		Fn: fn,
	}
}

func createNamespace(sexp *ast.SExp, env *object.Env, eval object.Eval) object.Object {
	ns := env.Child()

	if err := asError(assign(sexp, ns, eval)); err != nil {
		return fmtError(sexp.Pos(), "creating namespace: %w", err)
	}

	return &object.Namespace{
		Env: ns,
	}
}

func assign(sexp *ast.SExp, env *object.Env, eval object.Eval) object.Object {
	args := sexp.Items
	if len(args)%2 != 0 {
		return &object.Error{
			Err: fmt.Errorf("assign requires an even number of arguments, got %d", len(args)),
		}
	}

	values := &object.Array{}

	for k, v := range seq.SlicePairs(args) {
		symbol, isSymbol := k.(*ast.Symbol)
		if !isSymbol {
			return &object.Error{
				Err: fmt.Errorf("%s: assign can only use symbols as names, got %s %q as an argument", k.Pos().From, k.Name(), k.String()),
			}
		}

		value := eval(v, env)

		values.Elements = append(values.Elements, value)

		env.Assign(symbol.Value, value)
	}

	return values
}

func Eval(node ast.Node, env *object.Env) object.Object {
	if env == nil {
		env = DefaultEnv()
	}

	switch node := node.(type) {
	case *ast.Literal[int64]:
		return object.PrimitiveOf(node.Value)
	case *ast.Literal[string]:
		return object.PrimitiveOf(node.Value)
	case *ast.Literal[float64]:
		return object.PrimitiveOf(node.Value)
	case *ast.Literal[bool]:
		return object.PrimitiveOf(node.Value)
	case *ast.Symbol:
		o, ok := env.LookUp(node.Value)
		if ok {
			return o
		}
		return Null
	case *ast.If:
		return evalIf(node, env, Eval)
	case *ast.SpecialOp:
		return evalSpecialOp(node, env, Eval)
	case *ast.Package:
		var result object.Object
		for _, n := range node.Nodes {
			result = Eval(n, env)
			if err, isErr := result.(*object.Error); isErr {
				return &object.Error{
					Err: fmt.Errorf("%s: %w", n.Pos().From, err.Err),
				}
			}
		}

		return result
	case *ast.SExp:
		return apply(node.Items[0], node.Items[1:], env, Eval)
	}

	return &object.Error{
		Err: fmt.Errorf("unexpected node %s", node.Name()),
	}
}

func builtinApply(sexp *ast.SExp, env *object.Env, eval object.Eval) object.Object {
	if len(sexp.Items) != 2 {
		return fmtError(sexp.Pos(), "apply got %d arguments, but expects 2: a function and an array", len(sexp.Items))
	}

	return apply(sexp.Items[0], sexp.Items[1:], env, eval)
}

func apply(fn ast.Node, args []ast.Node, env *object.Env, eval object.Eval) object.Object {
	head := Eval(fn, env)

	switch head := head.(type) {
	case *object.Builtin:
		return head.Fn(&ast.SExp{
			PosRange: fn.Pos(),
			Items:    args,
		}, env, Eval)
	case *object.Namespace:
		if len(args) < 1 {
			return fmtError(fn.Pos(), "namespace missing an argument")
		}
		k := args[0]
		key, isSymbol := k.(*ast.Symbol)
		if !isSymbol {
			return fmtError(k.Pos(), "can use only symbols to look into namespace, used a %s %q", k.Name(), k.String())
		}
		o, _ := head.Env.LookUp(key.Value)
		return o
	case *object.Array:
		if len(args) == 0 {
			return head
		}
		index := eval(args[0], env)
		if err := asError(index); err != nil {
			return fmtError(args[0].Pos(), "evaluating array index: %w", err)
		}

		idx, ok := index.(*object.Primitive[int64])
		if !ok {
			return fmtError(args[0].Pos(), "evaluating array index: want a int64, got %s", index.Type())
		}

		if idx.Value < 0 || idx.Value >= int64(len(head.Elements)) {
			return fmtError(args[0].Pos(), "array index %d is out of bounds 0..%d", idx.Value, len(head.Elements))
		}

		return head.Elements[idx.Value]
	}

	return &object.Error{
		Err: fmt.Errorf("unexpected apply argument %T %s", head, head.Inspect()),
	}
}

func isSymbolName(node ast.Node, name string) bool {
	symbol, ok := node.(*ast.Symbol)
	return ok && symbol.Value == name
}

func isNode[N ast.Node](node ast.Node) bool {
	_, ok := node.(N)
	return ok
}

func evalSpecialOp(op *ast.SpecialOp, env *object.Env, eval object.Eval) object.Object {
	switch op.Op {
	case "*":
		return multiply(op, env, eval)
	case "+":
		return sum(op, env)
	default:
		return &object.Error{
			Err: fmt.Errorf("%s: unexpected operation %q", op.From, op.Op),
		}
	}
}

func evalIf(op *ast.If, env *object.Env, eval object.Eval) object.Object {
	condition := eval(op.Cond, env)

	var ok bool
	switch condition := condition.(type) {
	case *object.Primitive[bool]:
		ok = condition.Value
	default:
		return fmtError(op.Pos(), "unexpected condition value: %s %q", condition.Type(), condition.Inspect())
	}

	if ok {
		return eval(op.Then, env)
	}

	if op.Else == nil {
		return Null
	}

	return eval(op.Else, env)
}

func gt(op *ast.SExp, env *object.Env, eval object.Eval) object.Object {
	if len(op.Items) < 2 {
		return fmtError(op.Pos(), "< want at lease 2 arguments, got %d", len(op.Items)-1)
	}

	args := op.Items
	var prev object.Object
	result := true
	for arg, err := range resolveMany(args, env, eval) {
		if err != nil {
			return fmtError(op.Pos(), "evaluation argument: %w", err)
		}

		if !object.IsPrimitive(arg) {
			return fmtError(op.Pos(), "only primitive types are supported, got %s", arg.Type())
		}

		if prev != nil && prev.Type() != arg.Type() {
			return fmtError(op.Pos(), ">: type error, want %s, got %s", prev.Type(), arg.Type())
		}

		if prev != nil {
			gt, ok := arg.(object.Ordered).Compare(prev)
			if !ok {
				return fmtError(op.Pos(), "unable to compare %s and %s", prev.Type(), arg.Type())
			}
			result = result && gt < 0
		}

		prev = arg
	}

	if result {
		return True
	}
	return False
}

func sum(op *ast.SpecialOp, env *object.Env) object.Object {
	xf, xi := 0.0, int64(0)
	hasFloats := false
	for i, a := range op.Items {
		arg := Eval(a, env)

		switch arg := arg.(type) {
		case *object.Error:
			return &object.Error{
				Err: fmt.Errorf("%s: +: argument %d: %w", a.Pos().From, i, arg.Err),
			}
		case *object.Primitive[int64]:
			xf += float64(arg.Value)
			xi += arg.Value
		case *object.Primitive[float64]:
			hasFloats = true
			xf += arg.Value
			xi += int64(arg.Value)
		default:
			return &object.Error{
				Err: fmt.Errorf("%s: +: unexpected argument %d type %s %q", a.Pos().From, i, arg.Type(), arg.Inspect()),
			}
		}
	}

	if hasFloats {
		return object.PrimitiveOf(xf)
	}

	return object.PrimitiveOf(xi)
}

func multiply(op *ast.SpecialOp, env *object.Env, eval object.Eval) object.Object {
	xf, xi := 1.0, int64(1)
	hasFloats := false
	for i, a := range op.Items {
		arg := eval(a, env)

		switch arg := arg.(type) {
		case *object.Error:
			return &object.Error{
				Err: fmt.Errorf("%s: *: argument %d: %w", a.Pos().From, i, arg.Err),
			}
		case *object.Primitive[int64]:
			xf *= float64(arg.Value)
			xi *= arg.Value
		case *object.Primitive[float64]:
			hasFloats = true
			xf *= arg.Value
			xi *= int64(arg.Value)
		default:
			return &object.Error{
				Err: fmt.Errorf("%s: *: unexpected argument %d type %s %q", a.Pos().From, i, arg.Type(), arg.Inspect()),
			}
		}
	}

	if hasFloats {
		return object.PrimitiveOf(xf)
	}

	return object.PrimitiveOf(xi)
}

func resolveMany(nodes []ast.Node, env *object.Env, eval object.Eval) iter.Seq2[object.Object, error] {
	return func(yield func(object.Object, error) bool) {
		for i, node := range nodes {
			result := eval(node, env)

			if err, isErr := result.(*object.Error); isErr {
				_ = yield(nil, fmt.Errorf("%s: %d: %w", node.Pos().From, i, err.Err))
				return
			}

			if !yield(result, nil) {
				return
			}
		}
	}
}

func resolveSkipKeys(nodes []ast.Node, env *object.Env) iter.Seq2[object.Object, error] {
	return func(yield func(object.Object, error) bool) {
		for i, node := range nodes {
			var result object.Object
			switch {
			case i%2 == 0:
				result = &object.AST{Node: node}
			default:
				result = Eval(node, env)
			}

			if err, isErr := result.(*object.Error); isErr {
				_ = yield(nil, fmt.Errorf("%s: %d: %w", node.Pos().From, i, err.Err))
				return
			}

			if !yield(result, nil) {
				return
			}
		}
	}
}

func fmtError(pos ast.PosRange, msg string, args ...any) *object.Error {
	args = append([]any{pos.From, pos.To}, args...)
	return &object.Error{
		Err: fmt.Errorf("%s-%s: "+msg, args...),
	}
}

func asError(obj object.Object) error {
	err, _ := obj.(*object.Error)
	if err == nil {
		return nil
	}

	return err.Err
}
