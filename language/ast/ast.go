// Package ast defines Sulisp syntax as generic atoms and ordinary lists.
// Forms are recognized by the compiler rather than represented by special
// syntax-tree node types.
package ast

import (
	"fmt"
	"strings"

	"github.com/ninedraft/sulisp/language/tokens"
)

// Node is any Sulisp syntax node.
type Node interface {
	fmt.Stringer
	Pos() PosRange
}

// PosRange identifies a node's source extent.
type PosRange struct {
	From, To tokens.Position
}

// Pos returns the source extent.
func (pos PosRange) Pos() PosRange {
	return pos
}

// AtomValue is a value that can be held by an Atom. String atoms are
// discriminated by Kind as string literals, symbols, or keywords.
type AtomValue interface {
	int64 | float64 | string | bool
}

// Atom is a scalar literal, symbol, or keyword.
type Atom[T AtomValue] struct {
	PosRange
	Kind  tokens.TokenKind
	Value T
}

func (atom Atom[T]) String() string {
	return fmt.Sprint(atom.Value)
}

// List is a parenthesized sequence. It deliberately has no special-form
// variants: the compiler interprets the first item when compiling a form.
type List struct {
	PosRange
	Items []Node
}

// NewList constructs an ordinary list at an unspecified source position.
func NewList(items ...Node) *List {
	return &List{Items: items}
}

func (list List) String() string {
	var text strings.Builder
	text.WriteByte('(')
	for index, item := range list.Items {
		if index > 0 {
			text.WriteByte(' ')
		}
		text.WriteString(item.String())
	}
	text.WriteByte(')')
	return text.String()
}
