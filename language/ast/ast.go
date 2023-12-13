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

func (lit *Literal[L]) Name() string {
	switch any(lit.Value).(type) {
	case string:
		return "string"
	case int64:
		return "int"
	case float64:
		return "float"
	}

	return tokens.TokenMalformed.String()
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

type True struct {
	PosRange
}

func (t *True) Equal(other Node) bool {
	if t == nil {
		return other == nil
	}

	_, ok := other.(*True)
	return ok
}

func (*True) Name() string { return "true" }

func (*True) String() string { return "true" }

func (t *True) Clone() Node {
	return shallow(t)
}

type False struct {
	PosRange
}

func (*False) Name() string { return "false" }

func (*False) String() string { return "false" }

func (f *False) Equal(other Node) bool {
	if f == nil {
		return other == nil
	}

	_, ok := other.(*False)
	return ok
}

func (f *False) Clone() Node {
	return shallow(f)
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
