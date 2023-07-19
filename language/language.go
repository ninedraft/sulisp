package language

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/exp/constraints"
)

type Expression interface {
	String() string
	Clone() Expression
	Equal(Expression) bool
}

type Sexp []Expression

func (s Sexp) String() string {
	buf := &strings.Builder{}
	buf.WriteByte('(')

	for i, e := range s {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(e.String())
	}

	buf.WriteByte(')')

	return buf.String()
}

func (s Sexp) Clone() Expression {
	clone := make(Sexp, len(s))
	for i, e := range s {
		clone[i] = e.Clone()
	}
	return clone
}

func (s Sexp) Equal(e Expression) bool {
	if sexp, ok := e.(Sexp); ok {
		if len(s) != len(sexp) {
			return false
		}

		for i, e := range s {
			if !e.Equal(sexp[i]) {
				return false
			}
		}

		return true
	}

	return false
}

type Symbol string

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) Clone() Expression {
	return s
}

func (s Symbol) Equal(e Expression) bool {
	if sym, ok := e.(Symbol); ok {
		return s == sym
	}

	return false
}

type Keyword string

func (k Keyword) String() string {
	return ":" + string(k)
}

func (k Keyword) Clone() Expression {
	return k
}

func (k Keyword) Equal(e Expression) bool {
	if key, ok := e.(Keyword); ok {
		return k == key
	}

	return false
}

func (k Keyword) Name() string {
	return string(k)
}

type LiteralValue interface {
	constraints.Complex | constraints.Signed | constraints.Unsigned | constraints.Float | ~string | ~bool | ~struct{}
}

type Literal[E LiteralValue] struct {
	Value E
}

func (*Literal[E]) IsLiteral() {}

func (lit *Literal[E]) String() string {
	if unsafe.Sizeof(lit.Value) == 0 {
		return "nothing"
	}

	if str, isStr := any(lit.Value).(string); isStr {
		return strconv.Quote(str)
	}

	return fmt.Sprint(lit.Value)
}

func (lit *Literal[E]) Clone() Expression {
	return &Literal[E]{lit.Value}
}

func (lit *Literal[E]) Equal(e Expression) bool {
	if l, ok := e.(*Literal[E]); ok {
		return l.Value == lit.Value
	}

	return false
}

type Comment struct {
	Pos   Position
	Value string
}

func (c *Comment) String() string {
	return ";" + c.Value
}

func (c *Comment) Clone() Expression {
	return &Comment{c.Pos, c.Value}
}

func (c *Comment) Equal(e Expression) bool {
	if comment, ok := e.(*Comment); ok {
		return c.Value == comment.Value
	}

	return false
}
