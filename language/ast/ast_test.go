package ast_test

import (
	"testing"

	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
)

func TestNodesCarrySourceRanges(t *testing.T) {
	t.Parallel()

	at := ast.PosRange{
		From: tokens.Position{File: "test.su", Line: 2, Column: 3},
		To:   tokens.Position{File: "test.su", Line: 2, Column: 8},
	}
	nodes := []ast.Node{
		ast.Atom[int64]{PosRange: at, Kind: tokens.TokenInt, Value: 42},
		ast.Atom[float64]{PosRange: at, Kind: tokens.TokenFloat, Value: 3.5},
		ast.Atom[string]{PosRange: at, Kind: tokens.TokenStr, Value: "hello"},
		ast.Atom[bool]{PosRange: at, Value: true},
		ast.Atom[string]{PosRange: at, Kind: tokens.TokenSymbol, Value: "answer"},
		ast.Atom[string]{PosRange: at, Kind: tokens.TokenKeyword, Value: ":answer"},
		ast.List{PosRange: at},
	}

	for _, node := range nodes {
		if got := node.Pos(); got != at {
			t.Errorf("%T.Pos() = %+v, want %+v", node, got, at)
		}
	}
}

func TestListIsAnOrdinarySExpression(t *testing.T) {
	t.Parallel()

	list := ast.List{Items: []ast.Node{
		ast.Atom[string]{Kind: tokens.TokenSymbol, Value: "+"},
		ast.Atom[int64]{Kind: tokens.TokenInt, Value: 1},
		ast.Atom[int64]{Kind: tokens.TokenInt, Value: 2},
	}}

	if got, want := list.String(), "(+ 1 2)"; got != want {
		t.Errorf("List.String() = %q, want %q", got, want)
	}
}
