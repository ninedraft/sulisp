package object

import (
	"cmp"
	"errors"
	"fmt"
	"hash/maphash"
	"io"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/ninedraft/itermore"
	"github.com/ninedraft/sulisp/language/ast"
)

type Kind string

const (
	ObjType         Kind = "type"
	ObjKind         Kind = "kind"
	ObjInteger      Kind = "integer"
	ObjAny          Kind = "any"
	ObjFloat64      Kind = "float64"
	ObjBool         Kind = "boolean"
	ObjString       Kind = "string"
	ObjNull         Kind = "null"
	ObjError        Kind = "error"
	ObjReturn       Kind = "return"
	ObjFunc         Kind = "function"
	ObjBuiltin      Kind = "builtin"
	ObjArray        Kind = "array"
	ObjAST          Kind = "ast"
	ObjNamespace    Kind = "namespace"
	ObjContinuation Kind = "continuation"
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

func (ot Kind) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, ObjKind)
	maphash.WriteComparable(h, string(ot))
}

type Object interface {
	Kind() Kind
	Hash(h *maphash.Hash)
	Compare(other Object) (int, bool)
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

func (primitive *Primitive[E]) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, primitive.Kind())
	maphash.WriteComparable(h, primitive.Value)
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

func (ot *Type) Compare(other Object) (int, bool) {
	o, ok := other.(*Type)
	if !ok {
		return 0, false
	}

	if ot.ObjKind != o.ObjKind {
		return 0, false
	}

	if len(ot.Params) != len(o.Params) {
		return 0, false
	}

	return 0, slices.EqualFunc(ot.Params, o.Params, func(a, b Type) bool {
		ord, ok := a.Compare(&b)
		return ord == 0 && ok
	})
}

func (ot *Type) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, ot.Kind())
	maphash.WriteComparable(h, ot.ObjKind)
	maphash.WriteComparable(h, len(ot.Params))

	//TODO: handle recursive types
	for _, param := range ot.Params {
		param.Hash(h)
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

func (ns *Namespace) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, ns.Kind())
	maphash.WriteComparable(h, ns.Env.id)
}

func (ns *Namespace) Compare(other Object) (int, bool) {
	o, ok := other.(*Namespace)
	if !ok {
		return 0, false
	}

	if ns.Env.id == o.Env.id {
		return 0, true
	}

	return 0, false
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

func (Null) Compare(other Object) (int, bool) {
	if _, ok := other.(Null); ok {
		return 0, true
	}

	return 0, false
}

var nullTypeNonce = reflect.TypeFor[Null]()

func (Null) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, nullTypeNonce)
}

type Error struct {
	ID  uint64
	Err error
}

func NewError(err error) *Error {
	return &Error{
		ID:  makeObjectID(),
		Err: err,
	}
}

var errorTypeNonce = reflect.TypeFor[Error]()

func (err *Error) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, err.Kind())
	maphash.WriteComparable(h, errorTypeNonce)
	maphash.WriteComparable(h, err.ID)
}

func (err *Error) Compare(other Object) (int, bool) {
	o, ok := other.(*Error)
	if !ok {
		return 0, false
	}

	isOther := errors.Is(err.Err, o.Err)
	isThis := errors.Is(o.Err, err.Err)

	if isOther && isThis {
		return 0, true
	}

	if isOther {
		return 1, true
	}

	if isThis {
		return -1, true
	}

	return 0, false
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

func (ret *Return) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, ret.Kind())
	maphash.WriteComparable(h, ret)
}

func (ret *Return) Inspect() string {
	return ret.Value.Inspect()
}

type Function struct {
	ID         uint64
	Parameters []*ast.Symbol
	Body       *ast.SExp
	Env        *Env
}

func NewFunction(
	Parameters []*ast.Symbol,
	Body *ast.SExp,
	Env *Env,
) *Function {
	return &Function{
		ID:         makeObjectID(),
		Parameters: Parameters,
		Body:       Body,
		Env:        Env,
	}
}

func (fn *Function) Kind() Kind {
	return ObjFunc
}

var fnTypeNonce = reflect.TypeFor[Function]()

func (fn *Function) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, fn.Kind())
	maphash.WriteComparable(h, fnTypeNonce)
	maphash.WriteComparable(h, fn.ID)
}

func (fn *Function) Inspect() string {
	str := &strings.Builder{}

	params := make([]string, 0, len(fn.Parameters))
	for _, param := range fn.Parameters {
		params = append(params, param.Value)
	}

	str.WriteString("(fn ")
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

func (builtin *Builtin) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, builtin.Kind())
	maphash.WriteComparable(h, builtin)
}

func (builtin *Builtin) Inspect() string {
	return fmt.Sprintf("<builtin %s>", builtin.Name)
}

func (builtin *Builtin) Compare(other Object) (int, bool) {
	o, ok := other.(*Builtin)
	if !ok {
		return 0, false
	}

	if builtin.Name == o.Name &&
		builtin.Type.Equal(o.Type) {
		return 0, true
	}

	return 0, false
}

type Array struct {
	Elements []Object
}

func (array *Array) Kind() Kind {
	return ObjArray
}

func (array *Array) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, array.Kind())
	maphash.WriteComparable(h, array)
}

func (array *Array) Compare(other Object) (int, bool) {
	o, ok := other.(*Array)
	if !ok {
		return 0, false
	}

	if len(array.Elements) != len(o.Elements) {
		return 0, false
	}

	for i := range array.Elements {
		ord, ok := array.Elements[i].Compare(o.Elements[i])
		if !ok {
			return 0, false
		}
		if ord != 0 {
			return ord, true
		}
	}

	return 0, true
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

func (a *AST) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, a.Kind())
	maphash.WriteComparable(h, a.Node.String())
}

func (a *AST) Compare(other Object) (int, bool) {
	o, ok := other.(*AST)
	if !ok {
		return 0, false
	}

	if a.Node.Equal(o.Node) {
		return 0, true
	}

	return 0, false
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
