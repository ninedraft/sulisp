package match

import (
	"fmt"

	"github.com/ninedraft/sulisp/language"
)

type special int

const (
	// AnyMore matches the rest of S-expression.
	AnyMore special = iota + 1
)

type Pattern func(expr language.Expression) bool

// S is a pattern for S-expression.
func S(patterns ...any) Pattern {
	return func(expr language.Expression) bool {
		sexp, ok := expr.(language.Sexp)
		if !ok {
			return false
		}

		return Match(sexp, patterns...)
	}
}

// P tests if given expression matches given expression type.
// If expression matches, P writes expression to dst.
// It can be used as placeholder for expression.
func P[E language.Expression](dst *E) Pattern {
	return func(expr language.Expression) bool {
		e, ok := expr.(E)
		if ok {
			*dst = e
		}
		return ok
	}
}

// Match checks if given S-expression matches given patterns.
// Examples in pseudocode:
//
//	Match(`(x :- int => int)`, &name, &type, &type) // true
//	Match(`(x :- int => int)`, &name, &type) // false
//	Match(`(+ y 1)`, &op, &left, &right) // true
//	Match(`(+ y 1)`, &op, &left, &right, &extra) // false
func Match(sexp language.Sexp, patterns ...any) bool {
	for i, expr := range sexp {
		if i >= len(patterns) {
			return false
		}

		pattern := patterns[i]
		if p, isSpecial := pattern.(special); isSpecial {
			return p == AnyMore
		}

		switch pattern := pattern.(type) {
		case Pattern:
			if !pattern(expr) {
				return false
			}
		case language.Expression:
			if !pattern.Equal(expr) {
				return false
			}
		default:
			msg := fmt.Sprintf("invalid pattern type: %T", pattern)
			panic(msg)
		}
	}

	return true
}
