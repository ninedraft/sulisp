package tests

import (
	"slices"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/interpreter/astwalk"
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/object"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
)

func TestASTWalk(t *testing.T) {
	t.Run("assign", func(t *testing.T) {
		testASTWalk(t, `
			(assign a 1)
			a
		`, object.PrimitiveOf[int64](1))
	})

	t.Run("namespace", func(t *testing.T) {
		testASTWalk(t, `
			(assign 
				ns (namespace
					a 1
					b (* 2 2)))
			
			(ns b)
		`, &object.Primitive[int64]{Value: 4})
	})

	t.Run(">", func(t *testing.T) {
		testASTWalk(t, `
			(> 10 1)
		`, &object.Primitive[bool]{Value: true})
	})

	t.Run("array constructor", func(t *testing.T) {
		testASTWalk(t, `
			(array 1 2)
		`, &object.Array{Elements: []object.Object{
			object.PrimitiveOf[int64](1),
			object.PrimitiveOf[int64](2),
		}})
	})

	t.Run("empty array", func(t *testing.T) {
		testASTWalk(t, `
			(array)
		`, &object.Array{})
	})

	t.Run("index array", func(t *testing.T) {
		testASTWalk(t, `
			((array 1 2 3) 2)
		`, object.PrimitiveOf[int64](3))
	})

	t.Run("if-then", func(t *testing.T) {
		testASTWalk(t, `
			(if true 1 2)
		`, object.PrimitiveOf[int64](1))
	})

	t.Run("if-else", func(t *testing.T) {
		testASTWalk(t, `
			(if false 1 2)
		`, object.PrimitiveOf[int64](2))
	})

	t.Run("if-else-nil", func(t *testing.T) {
		testASTWalk(t, `
			(if false 1)
		`, object.Null{})
	})
}

func testASTWalk(t *testing.T, input string, want object.Object) {
	pkg := read(t, input)

	got, env := eval(t, pkg)

	t.Log("names", slices.Sorted(env.Names))

	if err, isErr := got.(*object.Error); isErr {
		t.Fatal("unexpected error", err.Err)
	}

	assertEq(t, want, got, "evaluation result")
}

func assertEq(t *testing.T, want, got object.Object, msg string, args ...any) {
	t.Helper()

	if want.Kind() != got.Kind() || want.Inspect() != got.Inspect() {
		t.Errorf(msg, args...)
		t.Errorf("\tgot(%s) != want(%s)", got.Kind(), want.Kind())
		t.Errorf("\tgot  %s", got.Inspect())
		t.Errorf("\twant %s", want.Inspect())
		return
	}
}

func eval(_ *testing.T, pkg *ast.Package) (object.Object, *object.Env) {
	env := astwalk.DefaultEnv()

	return astwalk.Eval(pkg, env), env
}

func read(t *testing.T, input string) *ast.Package {
	t.Helper()

	lex := lexer.NewLexer(t.Name(), strings.NewReader(input))
	pkg, err := parser.New(lex).Parse()

	if err != nil {
		t.Fatalf("parsing: %v", err)
	}

	return pkg
}
