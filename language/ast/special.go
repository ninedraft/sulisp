package ast

import (
	"slices"
	"strings"
)

type ImportGo struct {
	PosRange
	Items []Node // string | symbol | (symbol string)
}

func (importgo *ImportGo) Name() string {
	return "import-go"
}

func (importgo *ImportGo) Equal(other Node) bool {
	if importgo == nil {
		return other == nil
	}

	if o, ok := other.(*ImportGo); ok {
		return equalSlices(importgo.Items, o.Items)
	}

	return false
}

func (importgo *ImportGo) Clone() Node {
	if importgo == nil {
		return nil
	}

	clone := *importgo

	for i, item := range clone.Items {
		clone.Items[i] = Clone(item)
	}

	return &clone
}

func (importgo *ImportGo) String() string {
	str := &strings.Builder{}

	const sep = "\n    "
	writeStrs(str, "(import-go", sep)
	joinStringers(str, sep, importgo.Items)
	writeStrs(str, ")")

	return str.String()
}

type If struct {
	PosRange
	Cond, Then Node
	Else       Node // optional
}

func (*If) Name() string { return "if" }

func (if_ *If) Equal(other Node) bool {
	if if_ == nil {
		return other == nil
	}

	o, ok := other.(*If)
	if !ok {
		return false
	}

	ok = if_.Cond.Equal(o.Cond) && if_.Then.Equal(o.Then)

	if if_.Else != nil {
		ok = ok && if_.Else.Equal(o.Else)
	}

	return ok
}

func (if_ *If) Clone() Node {
	if if_ == nil {
		return nil
	}

	clone := *if_
	clone.Cond = Clone(if_.Cond)
	clone.Then = Clone(if_.Then)
	clone.Else = Clone(if_.Else)

	return &clone
}

func (if_ *If) String() string {
	str := &strings.Builder{}

	writeStrs(str, "(if ",
		if_.Cond.String(), "\n",
		"    ", if_.Then.String())

	if if_.Else != nil {
		writeStrs(str, "\n    ", if_.Else.String())
	}

	writeStrs(str, ")")

	return str.String()
}

type SpecialOp struct {
	PosRange
	Op    string
	Items []Node
}

func (special *SpecialOp) Name() string {
	return special.Op
}

func (special *SpecialOp) Equal(other Node) bool {
	if special == nil {
		return other == nil
	}

	if o, ok := other.(*SpecialOp); ok {
		return special.Op == o.Op && equalSlices(special.Items, o.Items)
	}

	return false
}

func (special *SpecialOp) Clone() Node {
	if special == nil {
		return nil
	}

	clone := *special
	clone.Items = cloneSlice(special.Items)

	return &clone
}

func (special *SpecialOp) String() string {
	str := &strings.Builder{}

	writeStrs(str, "(", special.Op)
	joinStringers(str, " ", special.Items)
	writeStrs(str, ")")

	return str.String()
}

/*
spec:

	(fn add (int int) (int error)
		(x, y)
		(+ x y))
*/
type FunctionLiteral struct {
	PosRange
	FnName *Symbol
	Spec   *FunctionSpec
	Body   Node
}

func (fn *FunctionLiteral) Name() string {
	return "*defn"
}

func (fn *FunctionLiteral) Equal(other Node) bool {
	if fn == nil {
		return other == nil
	}

	if o, ok := other.(*FunctionLiteral); ok {
		return fn.FnName.Equal(o.FnName) && fn.Spec.Equal(o.Spec) && fn.Body.Equal(o.Body)
	}

	return false
}

func (fn *FunctionLiteral) Clone() Node {
	if fn == nil {
		return nil
	}

	clone := *fn
	clone.FnName = Clone(fn.FnName)
	clone.Spec = Clone(fn.Spec)
	clone.Body = Clone(fn.Body)

	return &clone
}

func (fn *FunctionLiteral) String() string {
	str := &strings.Builder{}

	writeStrs(str, "(fn ", fn.FnName.String(), " ", fn.Spec.String(), "\n    ", fn.Body.String(), ")")

	return str.String()
}

type FunctionSpec struct {
	PosRange
	Params []*FieldSpec
	Ret    []Node
}

func (spec *FunctionSpec) Name() string {
	return "*fn"
}

func (spec *FunctionSpec) Equal(other Node) bool {
	if spec == nil {
		return other == nil
	}

	if o, ok := other.(*FunctionSpec); ok {
		return slices.EqualFunc(spec.Params, o.Params, (*FieldSpec).equal) &&
			equalSlices(spec.Ret, o.Ret)
	}

	return false
}

func (spec *FunctionSpec) Clone() Node {
	if spec == nil {
		return nil
	}

	clone := shallow(spec)

	for i, param := range clone.Params {
		clone.Params[i] = param.clone()
	}

	clone.Ret = cloneSlice(spec.Ret)

	return clone
}

func (spec *FunctionSpec) String() string {
	str := &strings.Builder{}

	writeStrs(str, "(*fn ")
	joinStringers(str, " ", spec.Params)

	if len(spec.Ret) > 0 {
		writeStrs(str, " -> ")
		joinStringers(str, " ", spec.Ret)
	}

	writeStrs(str, ")")

	return str.String()
}

type FieldSpec struct {
	Names []*Symbol
	Type  Node
}

func (field *FieldSpec) String() string {
	str := &strings.Builder{}

	joinStringers(str, " ", field.Names)
	writeStrs(str, " :- ", field.Type.String())

	return str.String()
}

func (field *FieldSpec) clone() *FieldSpec {
	if field == nil {
		return nil
	}

	clone := shallow(field)
	clone.Names = cloneSlice(field.Names)
	clone.Type = Clone(field.Type)

	return clone
}

func (field *FieldSpec) equal(other *FieldSpec) bool {
	return equalSlices(field.Names, other.Names) && field.Type.Equal(other.Type)
}
