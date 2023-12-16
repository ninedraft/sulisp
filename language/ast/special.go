package ast

import (
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
