package parser

import (
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
)

func sexpMatch(sexp *ast.SExp, pats ...pattern) bool {
	if len(sexp.Items) != len(pats) {
		return false
	}

	for i, match := range pats {
		node := sexp.Items[i]
		if !match(node) {
			return false
		}
	}

	return true
}

func pMatch[N ast.Node]() pattern {
	return func(n ast.Node) bool {
		_, ok := n.(N)
		return ok
	}
}

func p[N ast.Node](dst *N) pattern {
	if dst == nil {
		panic("[parser] pattern matcher got a nil node target")
	}

	kind := tokens.TokenUndefined

	if n := ast.Node(*dst); n != nil {
		kind = n.Tok()
	}

	return func(n ast.Node) bool {
		if kind != tokens.TokenUndefined && n.Tok() != kind {
			return false
		}

		got, ok := n.(N)
		if !ok {
			return false
		}

		if ast.Node(*dst) != nil && !(*dst).Equal(got) {
			return false
		}

		*dst = got

		return true
	}
}

type pattern func(n ast.Node) bool

func pOr(patts ...pattern) pattern {
	return func(n ast.Node) bool {
		for _, match := range patts {
			if match(n) {
				return true
			}
		}
		return false
	}
}
