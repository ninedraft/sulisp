package object

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/ninedraft/itermore"
	"github.com/ninedraft/sulisp/language/ast"
)

type Kind string

const (
	ObjType      Kind = "type"
	ObjKind      Kind = "kind"
	ObjInteger   Kind = "integer"
	ObjAny       Kind = "any"
	ObjFloat64   Kind = "float64"
	ObjBool      Kind = "boolean"
	ObjString    Kind = "string"
	ObjNull      Kind = "null"
	ObjError     Kind = "error"
	ObjReturn    Kind = "return"
	ObjFunc      Kind = "function"
	ObjBuiltin   Kind = "builtin"
	ObjArray     Kind = "array"
	ObjAST       Kind = "ast"
	ObjNamespace Kind = "namespace"
)

var Kinds = []Kind{
	ObjType,
	ObjKind,
	ObjInteger,
	ObjAny,
	ObjFloat64,
	ObjBool,
	ObjString,
	ObjNull,
	ObjError,
	ObjReturn,
	ObjFunc,
	ObjBuiltin,
	ObjArray,
	ObjAST,
	ObjNamespace,
}

func (ot Kind) Kind() Kind { return ObjKind }

func (ot Kind) Inspect() string {
	return string(ot)
}

type Object interface {
	Kind() Kind
	Inspect() string
}

type Ordered interface {
	Compare(other Object) (int, bool)
}

type markPrimitive interface{ isPrimitive() }

func IsPrimitive(obj Object) bool {
	_, ok := obj.(markPrimitive)
	return ok
}

func PrimitiveOf[E PrimitiveTypes](value E) *Primitive[E] {
	return &Primitive[E]{Value: value}
}

type PrimitiveTypes interface {
	string | int64 | float64 | bool
}

type Primitive[E PrimitiveTypes] struct {
	Value E
	markPrimitive
}

func (primitive *Primitive[E]) Compare(other Object) (_ int, ok bool) {
	o, ok := other.(*Primitive[E])
	if !ok {
		return 0, false
	}

	if primitive.Value == o.Value {
		return 0, true
	}

	switch v := any(primitive.Value).(type) {
	case int64:
		return cmp.Compare(v, any(o.Value).(int64)), true
	case float64:
		return cmp.Compare(v, any(o.Value).(float64)), true
	case string:
		return cmp.Compare(v, any(o.Value).(string)), true
	case bool:
		if v && primitive.Value != o.Value {
			return 1, true
		}
		return -1, false
	}

	return 0, false
}

func (primitive *Primitive[E]) Inspect() string {
	return fmt.Sprint(primitive.Value)
}

func (primitive *Primitive[E]) Kind() Kind {
	v := any(Primitive[E]{}.Value)
	switch v.(type) {
	case string:
		return ObjString
	case int64:
		return ObjInteger
	case float64:
		return ObjFloat64
	case bool:
		return ObjBool
	default:
		panic(fmt.Sprintf("unexpected primitive type %T", v))
	}
}

type Type struct {
	ObjKind Kind
	Params  []Type
}

func TypeFor(kind Kind, paramsKinds ...Kind) *Type {
	params := make([]Type, 0, len(paramsKinds))
	for _, param := range paramsKinds {
		params = append(params, Type{ObjKind: param})
	}

	return &Type{
		ObjKind: kind,
		Params:  params,
	}
}

func (ot *Type) Equal(other *Type) bool {
	return ot.ObjKind == other.ObjKind &&
		slices.EqualFunc(ot.Params, other.Params, func(a, b Type) bool {
			return a.Equal(&b)
		})
}

func (ot *Type) Kind() Kind {
	return ObjType
}

func (ot *Type) Inspect() string {
	str := &strings.Builder{}
	str.WriteString("(type ")
	str.WriteString(ot.ObjKind.Inspect())

	for _, param := range ot.Params {
		str.WriteString(" ")
		str.WriteString(param.Inspect())
	}

	str.WriteString(")")
	return str.String()
}

type Namespace struct {
	Env *Env
}

func (*Namespace) Kind() Kind {
	return ObjNamespace
}

func (ns *Namespace) Inspect() string {
	names := slices.Sorted(maps.Keys(ns.Env.values))

	str := &strings.Builder{}
	str.WriteString("(namespace ")

	if len(names) > 0 {
		str.WriteString("\n\t")
	}

	rows := make([]string, 0, len(ns.Env.values))
	for _, name := range names {
		decl, _ := ns.Env.LookUp(name)
		rows = append(rows, name+" "+decl.Inspect())
	}

	itermore.CollectJoin(str, itermore.Slice(rows), "\n\t")

	str.WriteString(")")

	return str.String()
}

type Null struct{}

func (Null) Kind() Kind {
	return ObjNull
}

func (Null) Inspect() string {
	return "null"
}

type Error struct {
	Err error
}

func (err *Error) Kind() Kind {
	return ObjError
}

func (err *Error) Inspect() string {
	return "!!! " + err.Err.Error()
}

type Return struct {
	Value Object
}

func (ret *Return) Kind() Kind { return ObjReturn }

func (ret *Return) Inspect() string {
	return ret.Value.Inspect()
}

type Function struct {
	Parameters []*ast.Symbol
	Body       *ast.SExp
	Env        *Env
}

func (fn *Function) Kind() Kind {
	return ObjFunc
}

func (fn *Function) Inspect() string {
	str := &strings.Builder{}

	params := make([]string, 0, len(fn.Parameters))
	for _, param := range fn.Parameters {
		params = append(params, param.Value)
	}

	str.WriteString("fn(")
	str.WriteString(strings.Join(params, ", "))
	str.WriteString(") {\n")
	str.WriteString(fn.Body.String())
	str.WriteString("\n}")

	return str.String()
}

type Eval = func(node ast.Node, env *Env) Object

type BuiltinFn func(args *ast.SExp, env *Env, eval Eval) Object

type Builtin struct {
	Name string
	Fn   BuiltinFn
	Type *Type
}

func (*Builtin) Kind() Kind { return ObjBuiltin }

func (builtin *Builtin) Inspect() string {
	return fmt.Sprintf("<builtin %s>", builtin.Name)
}

type Array struct {
	Elements []Object
}

func (array *Array) Kind() Kind {
	return ObjArray
}

func (array *Array) Inspect() string {
	str := &strings.Builder{}
	str.WriteRune('[')

	joinObjects(str, ", ", array.Elements)

	str.WriteRune(']')
	return str.String()
}

type AST struct {
	Node ast.Node
}

func (a *AST) Kind() Kind { return ObjAST }

func (a *AST) Inspect() string {
	return a.Node.String()
}

func joinObjects[O Object](dst io.StringWriter, sep string, strs []O) {
	if len(strs) == 0 {
		return
	}

	_, _ = dst.WriteString(strs[0].Inspect())

	for _, str := range strs[1:] {
		_, _ = dst.WriteString(sep)
		_, _ = dst.WriteString(str.Inspect())
	}
}
