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
	Tok() tokens.TokenKind
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

func (lit *Literal[L]) Tok() tokens.TokenKind {
	switch any(lit.Value).(type) {
	case string:
		return tokens.TokenStr
	case int64:
		return tokens.TokenInt
	case float64:
		return tokens.TokenFloat
	}

	return tokens.TokenMalformed
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

	if o, ok := other.(*Package); ok {
		return equalSlices(pkg.Nodes, o.Nodes)
	}

	return false
}

func (*Package) Tok() tokens.TokenKind {
	return tokens.TokenLParen
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
	if node == nil {
		return node == nil
	}

	if o, ok := node.(*Symbol); ok {
		return sym.Value == o.Value
	}

	return false
}

func (*Symbol) Tok() tokens.TokenKind { return tokens.TokenSymbol }

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

	if o, ok := node.(*Keyword); ok {
		return kw.Value == o.Value
	}

	return false
}

func (*Keyword) Tok() tokens.TokenKind { return tokens.TokenKeyword }

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

func (sexp *SExp) Tok() tokens.TokenKind {
	return tokens.TokenLParen
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

	if o, ok := other.(*SExp); ok {
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

func (*True) Tok() tokens.TokenKind { return tokens.TokenSymbol }

func (*True) String() string { return "true" }

func (t *True) Clone() Node {
	return shallow(t)
}

type False struct {
	PosRange
}

func (*False) Tok() tokens.TokenKind { return tokens.TokenSymbol }

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
