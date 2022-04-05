package ast

import "fmt"

type Expr interface {
	fmt.Stringer
	IsExpr()
}

type LiteralTypes interface {
	int64 | float64 | string | rune
}

type Literal[E LiteralTypes] struct {
	Kind  TokenKind
	Value E
}

func (*Literal[E]) IsExpr() {}

func (literal *Literal[E]) String() string {
	return "(" + literal.Kind.String() + " " + fmt.Sprint(literal.Value) + ")"
}
