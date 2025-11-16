package ast

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/ninedraft/sulisp/language/tokens"
)

type Node interface {
	fmt.Stringer
	Equal(other Node) bool
	Name() string
	Pos() PosRange
	Clone() Node
}

type PosRange struct {
	From, To tokens.Position
}

func (pos PosRange) Pos() PosRange {
	return pos
}

type LiteralValue interface {
	string | int64 | float64 | bool
}

type Literal[L LiteralValue] struct {
	PosRange
	Value L
}

func (lit *Literal[E]) Equal(other Node) bool {
	if lit == nil {
		return other == nil
	}

	o, ok := other.(*Literal[E])
	if !ok {
		return false
	}

	return lit.Value == o.Value
}

func (*Literal[L]) Name() string {
	var v L
	switch any(v).(type) {
	case string:
		return "string"
	case int64:
		return "int"
	case float64:
		return "float"
	case bool:
		return "bool"
	}

	return fmt.Sprintf("literal[%T]", v)
}

func (lit *Literal[L]) String() string {
	return fmt.Sprint(lit.Value)
}

func (lit *Literal[L]) Clone() Node {
	return shallow(lit)
}

type Package struct {
	PosRange
	Nodes []Node
}

func (pkg *Package) Equal(other Node) bool {
	if pkg == nil {
		return other == nil
	}

	if o, _ := other.(*Package); o != nil {
		return equalSlices(pkg.Nodes, o.Nodes)
	}

	return false
}

func (*Package) Name() string {
	return "package"
}

func (pkg *Package) String() string {
	str := &strings.Builder{}
	joinStringers(str, "\n\n", pkg.Nodes)
	return str.String()
}

func (pkg *Package) Clone() Node {
	if pkg == nil {
		return nil
	}

	clone := *pkg
	clone.Nodes = cloneSlice(pkg.Nodes)

	return &clone
}

type Symbol struct {
	PosRange
	Value string
}

func (sym *Symbol) Equal(node Node) bool {
	if sym == nil {
		return node == nil
	}

	if o, _ := node.(*Symbol); o != nil {
		return sym.Value == o.Value
	}

	return false
}

func (*Symbol) Name() string { return tokens.TokenSymbol.String() }

func (sym *Symbol) String() string { return sym.Value }

func (sym *Symbol) Clone() Node {
	return shallow(sym)
}

type Keyword struct {
	PosRange
	Value string
}

func (kw *Keyword) Equal(node Node) bool {
	if node == nil {
		return node == nil
	}

	if o, _ := node.(*Keyword); o != nil {
		return kw.Value == o.Value
	}

	return false
}

func (*Keyword) Name() string { return tokens.TokenKeyword.String() }

func (kw *Keyword) String() string { return kw.Value }

func (kw *Keyword) Clone() Node {
	return shallow(kw)
}

type SExp struct {
	PosRange
	Items []Node
}

func NewSexp(items ...Node) *SExp {
	return &SExp{
		Items: items,
	}
}

func (sexp *SExp) Name() string {
	return "s-expr"
}

func (sexp *SExp) String() string {
	str := &strings.Builder{}

	str.WriteRune('(')
	joinStringers(str, " ", sexp.Items)
	str.WriteRune(')')

	return str.String()
}

func (sexp *SExp) Equal(other Node) bool {
	if sexp == nil {
		return other == nil
	}

	if o, _ := other.(*SExp); o != nil {
		return equalSlices(sexp.Items, o.Items)
	}

	return false
}

func (sexp *SExp) Clone() Node {
	if sexp == nil {
		return nil
	}

	clone := *sexp
	clone.Items = cloneSlice(sexp.Items)

	return &clone
}

type DotSelector struct {
	PosRange
	Left, Right Node // left.right
}

func (dot *DotSelector) Equal(other Node) bool {
	if dot == nil {
		return other == nil
	}

	if o, _ := other.(*DotSelector); o != nil {
		return dot.Left.Equal(o.Left) && dot.Right.Equal(o.Right)
	}

	return false
}

func (*DotSelector) Name() string {
	return "dot-selector"
}

func (dot *DotSelector) String() string {
	return "(. " + dot.Left.String() + " " + dot.Right.String() + ")"
}

func (dot *DotSelector) Clone() Node {
	if dot == nil {
		return nil
	}

	clone := shallow(dot)
	clone.Left = Clone(dot.Left)
	clone.Right = Clone(dot.Right)

	return clone
}

type Sequence struct {
	PosRange
	Items []Node
}

func (seq *Sequence) Name() string {
	return "sequence"
}

func (seq *Sequence) String() string {
	if seq == nil {
		return "nil-sequence"
	}

	str := &strings.Builder{}
	str.WriteString("(begin")
	for _, item := range seq.Items {
		str.WriteRune(' ')
		str.WriteString(item.String())
	}
	str.WriteString(")")
	return str.String()
}

func (seq *Sequence) Equal(other Node) bool {
	if seq == nil {
		return other == nil
	}

	o, ok := other.(*Sequence)
	if !ok {
		return false
	}

	return equalSlices(seq.Items, o.Items)
}

func (seq *Sequence) Clone() Node {
	if seq == nil {
		return nil
	}

	clone := *seq
	clone.Items = cloneSlice(seq.Items)
	return &clone
}

type While struct {
	PosRange
	Cond Node
	Body Node
}

func (while *While) Name() string {
	return "while"
}

func (while *While) String() string {
	if while == nil {
		return "nil-while"
	}

	str := &strings.Builder{}
	str.WriteString("(while ")
	if while.Cond != nil {
		str.WriteString(while.Cond.String())
	}
	str.WriteString(" ")
	if while.Body != nil {
		str.WriteString(while.Body.String())
	}
	str.WriteString(")")
	return str.String()
}

func (while *While) Equal(other Node) bool {
	if while == nil {
		return other == nil
	}

	o, ok := other.(*While)
	if !ok {
		return false
	}

	return equalNodes(while.Cond, o.Cond) && equalNodes(while.Body, o.Body)
}

func (while *While) Clone() Node {
	if while == nil {
		return nil
	}

	clone := *while
	clone.Cond = Clone(while.Cond)
	clone.Body = Clone(while.Body)
	return &clone
}

type Function struct {
	PosRange
	Identifier string
	Parameters []*Symbol
	Body       Node
}

func (fn *Function) Name() string {
	return "function"
}

func (fn *Function) Equal(other Node) bool {
	if fn == nil {
		return other == nil
	}

	o, ok := other.(*Function)
	if !ok {
		return false
	}

	return fn.Identifier == o.Identifier &&
		equalSlices(fn.Parameters, o.Parameters) &&
		equalNodes(fn.Body, o.Body)
}

func (fn *Function) Clone() Node {
	if fn == nil {
		return nil
	}

	clone := *fn
	clone.Parameters = cloneSlice(fn.Parameters)
	clone.Body = Clone(fn.Body)
	return &clone
}

func (fn *Function) String() string {
	if fn == nil {
		return "nil-function"
	}

	str := &strings.Builder{}
	str.WriteString("(fn ")
	str.WriteString(fn.Identifier)
	str.WriteString(" (")
	for i, param := range fn.Parameters {
		if i > 0 {
			str.WriteString(" ")
		}
		str.WriteString(param.String())
	}
	str.WriteString(") ")
	if fn.Body != nil {
		str.WriteString(fn.Body.String())
	}
	str.WriteString(")")
	return str.String()
}

type Assign struct {
	PosRange
	Target *Symbol
	Value  Node
}

func (assign *Assign) Name() string {
	return "assign"
}

func (assign *Assign) Equal(other Node) bool {
	if assign == nil {
		return other == nil
	}

	o, ok := other.(*Assign)
	if !ok {
		return false
	}

	return equalNodes(assign.Target, o.Target) &&
		equalNodes(assign.Value, o.Value)
}

func (assign *Assign) Clone() Node {
	if assign == nil {
		return nil
	}

	clone := *assign
	if assign.Target != nil {
		clone.Target = Clone(assign.Target)
	}
	clone.Value = Clone(assign.Value)
	return &clone
}

func (assign *Assign) String() string {
	if assign == nil {
		return "nil-assign"
	}

	str := &strings.Builder{}
	str.WriteString("(assign ")
	if assign.Target != nil {
		str.WriteString(assign.Target.String())
	}
	str.WriteString(" ")
	if assign.Value != nil {
		str.WriteString(assign.Value.String())
	}
	str.WriteString(")")
	return str.String()
}

func equalNodes(a, b Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(b)
}

func Clone[E Node](node E) E {
	n := Node(node)

	if n == nil {
		var empty E
		return empty
	}

	return node.Clone().(E)
}

func cloneSlice[E Node](slice []E) []E {
	clone := make([]E, len(slice))

	for i, item := range slice {
		clone[i] = Clone(item)
	}

	return clone
}

func shallow[E any](ptr *E) *E {
	if ptr == nil {
		return nil
	}

	clone := *ptr
	return &clone
}

func writeStrs(wr io.StringWriter, items ...string) {
	for _, item := range items {
		_, _ = wr.WriteString(item)
	}
}

func joinStringers[S fmt.Stringer](wr io.StringWriter, sep string, items []S) {
	if len(items) == 0 {
		return
	}

	_, _ = wr.WriteString(items[0].String())

	for _, item := range items[1:] {
		_, _ = wr.WriteString(sep)
		_, _ = wr.WriteString(item.String())
	}
}

func equalSlices[N1, N2 Node](slice []N1, other []N2) bool {
	return slices.EqualFunc(slice, other, func(n1 N1, n2 N2) bool {
		return n1.Equal(n2)
	})
}
