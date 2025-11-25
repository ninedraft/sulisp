package object

import (
	"errors"
	"hash/maphash"
	"testing"

	"github.com/ninedraft/sulisp/language/ast"
)

func hashWithSeed(seed maphash.Seed, fn func(*maphash.Hash)) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	fn(&h)
	return h.Sum64()
}

func TestKindHash(t *testing.T) {
	seed := maphash.MakeSeed()
	if got, want := hashWithSeed(seed, ObjInteger.Hash), hashWithSeed(seed, ObjInteger.Hash); got != want {
		t.Fatalf("expected matching hashes for the same kind, got %d vs %d", got, want)
	}

	if kindHash := hashWithSeed(seed, ObjInteger.Hash); kindHash == hashWithSeed(seed, ObjBool.Hash) {
		t.Fatalf("expected different hashes for different kinds, got %d twice", kindHash)
	}
}

func TestPrimitiveHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := PrimitiveOf(int64(42))
	second := PrimitiveOf(int64(42))
	third := PrimitiveOf(int64(100))

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected same hash for identical primitives")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, third.Hash) {
		t.Fatalf("expected different hashes for different primitives")
	}
}

func TestTypeHash(t *testing.T) {
	seed := maphash.MakeSeed()

	a := TypeFor(ObjArray, ObjInteger)
	b := TypeFor(ObjArray, ObjInteger)
	c := TypeFor(ObjArray, ObjFloat64)

	if hashWithSeed(seed, a.Hash) != hashWithSeed(seed, b.Hash) {
		t.Fatalf("expected same hash for equivalent types")
	}

	if hashWithSeed(seed, a.Hash) == hashWithSeed(seed, c.Hash) {
		t.Fatalf("expected different hashes for different types")
	}
}

func TestNamespaceHash(t *testing.T) {
	seed := maphash.MakeSeed()

	env := NewEnv()
	ns := &Namespace{Env: env}
	nsSibling := &Namespace{Env: env}
	nsChild := &Namespace{Env: env.Child()}

	if hashWithSeed(seed, ns.Hash) != hashWithSeed(seed, nsSibling.Hash) {
		t.Fatalf("expected namespaces wrapping the same env to hash equal")
	}

	if hashWithSeed(seed, ns.Hash) == hashWithSeed(seed, nsChild.Hash) {
		t.Fatalf("expected namespaces wrapping different envs to hash differently")
	}
}

func TestNullHash(t *testing.T) {
	seed := maphash.MakeSeed()

	if got, want := hashWithSeed(seed, Null{}.Hash), hashWithSeed(seed, Null{}.Hash); got != want {
		t.Fatalf("expected null to hash consistently, got %d vs %d", got, want)
	}
}

func TestErrorHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := NewError(errors.New("boom"))
	second := NewError(errors.New("boom"))

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected error to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct errors to hash differently")
	}
}

func TestFunctionHash(t *testing.T) {
	seed := maphash.MakeSeed()

	env := NewEnv()
	first := NewFunction(nil, &ast.SExp{}, env)
	second := NewFunction(nil, &ast.SExp{}, env)

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected function to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct functions to hash differently")
	}
}

func TestBuiltinHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := &Builtin{
		Fn:   nil,
		Type: TypeFor(ObjAny),
	}
	second := &Builtin{
		Fn:   nil,
		Type: TypeFor(ObjAny),
	}

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected builtin to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct builtins to hash differently")
	}
}

func TestArrayHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := &Array{
		Elements: []Object{PrimitiveOf(int64(1))},
	}
	second := &Array{
		Elements: []Object{PrimitiveOf(int64(1))},
	}

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected array to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct arrays to hash differently")
	}
}

func TestReturnHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := &Return{Value: Null{}}
	second := &Return{Value: Null{}}

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected return wrapper to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct returns to hash differently")
	}
}

func TestASTHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := &AST{Node: &ast.Literal[int64]{Value: 5}}
	second := &AST{Node: &ast.Literal[int64]{Value: 5}}
	different := &AST{Node: &ast.Literal[int64]{Value: 6}}

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected AST nodes with the same value to hash the same")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, different.Hash) {
		t.Fatalf("expected AST nodes with different values to hash differently")
	}
}
