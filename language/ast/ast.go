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

type Atom[E AtomValue] struct {
	PosRange
	Kind  tokens.TokenKind
	Value E
}

func (*Atom[E]) IsAtom() struct{} {
	return struct{}{}
}

func (atom *Atom[E]) Equal(other Node) bool {
	if atom == nil {
		return other == nil
	}

	o, ok := other.(*Atom[E])
	if !ok {
		return false
	}

	return atom.Kind == o.Kind && atom.Value == o.Value
}

func (atom *Atom[E]) Name() string {
	return atom.Kind.String()
}

func (atom *Atom[E]) String() string {
	if atom == nil {
		return "<nil>"
	}
	return fmt.Sprint(atom.Value)
}

func (atom *Atom[E]) Clone() Node {
	return cloneShallow(atom)
}

type AtomValue interface {
	string | int64 | float64 | bool
}

type List struct {
	PosRange
	Items []Node
}

func NewList(items ...Node) *List {
	return &List{
		Items: items,
	}
}

func (list *List) Name() string {
	return "s-expr"
}

func (list *List) String() string {
	str := &strings.Builder{}

	str.WriteRune('(')
	joinStringers(str, " ", list.Items)
	str.WriteRune(')')

	return str.String()
}

func (list *List) Equal(other Node) bool {
	if list == nil {
		return other == nil
	}

	if o, _ := other.(*List); o != nil {
		return equalSlices(list.Items, o.Items)
	}

	return false
}

func (list *List) Clone() Node {
	if list == nil {
		return nil
	}

	clone := *list
	clone.Items = cloneSlice(list.Items)

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

	clone := cloneShallow(dot)
	clone.Left = Clone(dot.Left)
	clone.Right = Clone(dot.Right)

	return clone
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

func cloneShallow[E any](ptr *E) *E {
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
