package parser

import (
	"errors"
	"fmt"

	"github.com/ninedraft/sulisp/language/ast"
)

func sexpMatch(sexp *ast.SExp, pats ...pattern) error {
	errs := make([]error, 0, len(pats))
	for i, match := range pats {
		var node ast.Node
		if i < len(sexp.Items) {
			node = sexp.Items[i]
		}

		if err := match(node); err != nil {
			err = fmt.Errorf("item %d: %w", i, err)
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

var (
	errUnexpectedNode = errors.New("unexpected node")
	errNoNode         = errors.New("no node")
	errNodeDontMatch  = errors.New("node don't match")
)

func pSexp(patterns ...pattern) pattern {
	return func(n ast.Node) error {
		sexp, ok := n.(*ast.SExp)
		if !ok {
			return fmt.Errorf("%w: want s-expr, got %s", errUnexpectedNode, n.Name())
		}

		return sexpMatch(sexp, patterns...)
	}
}

func pMatch[N ast.Node]() pattern {
	return func(n ast.Node) error {
		_, ok := n.(N)

		if ok {
			return nil
		}

		var want N

		if n == nil {
			return fmt.Errorf("%w: want %T", errNoNode, want)
		}

		return fmt.Errorf("%w: want %T, got %s", errUnexpectedNode, want, n.Name())
	}
}

func p[N ast.Node](dst *N, sub ...pattern) pattern {
	if dst == nil {
		panic("[parser] pattern matcher got a nil node target")
	}

	return func(n ast.Node) error {
		got, ok := n.(N)
		if !ok {
			return fmt.Errorf("%w: want %T, got %s", errUnexpectedNode, got, n.Name())
		}

		if isEqual(got, *dst) {
			return fmt.Errorf("%w: got %q, want %q", errNodeDontMatch, got, *dst)
		}
		for _, match := range sub {
			err := match(n)
			if err != nil {
				return err
			}
		}

		*dst = got

		return nil
	}
}

type pattern func(n ast.Node) error

func pOr(patts ...pattern) pattern {
	return func(n ast.Node) error {
		errs := make([]error, 0, len(patts))

		for _, match := range patts {
			err := match(n)
			if err == nil {
				return nil
			}
			errs = append(errs, err)
		}

		return errors.Join(errs...)
	}
}

func pOpt[N ast.Node](dst *N) pattern {
	return func(n ast.Node) error {
		if n == nil {
			return nil
		}

		return p(dst)(n)
	}
}

func isEqual(a, b ast.Node) bool {
	if a == nil {
		return b == nil
	}

	return a.Equal(b)
}
