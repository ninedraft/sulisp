package parser_test

import (
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_Symbol(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(applorange)
	`)

	sexp := requireItem[*ast.SExp](t, pkg.Nodes, 0, "parsed package")
	symbol := requireItem[*ast.Symbol](t, sexp.Items, 0, "parsed symbol")

	assertEqual(t, &ast.Symbol{Value: "applorange"}, symbol, "parsed symbol")
}

func TestParse_Keyword(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(:applorange)
	`)

	sexp := requireItem[*ast.SExp](t, pkg.Nodes, 0, "parsed package")
	symbol := requireItem[*ast.Keyword](t, sexp.Items, 0, "parsed symbol")

	assertEqual(t, &ast.Keyword{Value: ":applorange"}, symbol, "parsed symbol")
}

func TestParse_Numbers(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(123 1.2)
	`)

	sexp := requireItem[*ast.SExp](t, pkg.Nodes, 0, "parsed package")

	intLit := requireItem[*ast.Literal[int64]](t, sexp.Items, 0, "parsed int literal")
	require.Equal(t, int64(123), intLit.Value, "parsed int literal")

	floatLit := requireItem[*ast.Literal[float64]](t, sexp.Items, 1, "parsed float literal")
	require.Equal(t, float64(1.2), floatLit.Value, "parsed float literal")
}

func TestParse_String(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		("applorange")
	`)

	sexp := requireItem[*ast.SExp](t, pkg.Nodes, 0, "parsed package")

	strLit := requireItem[*ast.Literal[string]](t, sexp.Items, 0, "parsed string literal")
	assertEqual(t, &ast.Literal[string]{Value: `"applorange"`}, strLit, "parsed string literal")
}

func TestParseImportGo(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(import-go  
			"fmt"
			net/http
			(_ "embed")
			(! database/sql))
	`)

	node := requireItem[*ast.ImportGo](t, pkg.Nodes, 0, "parsed package")

	want := &ast.ImportGo{
		Items: []ast.Node{
			&ast.Literal[string]{Value: `"fmt"`},
			&ast.Symbol{Value: "net/http"},
			ast.NewSexp(&ast.Symbol{Value: "_"}, &ast.Literal[string]{Value: `"embed"`}),
			ast.NewSexp(&ast.Symbol{Value: "!"}, &ast.Symbol{Value: "database/sql"}),
		},
	}

	if !node.Equal(want) {
		t.Error("parsed node is not equal to the expected one")
		t.Error("got:\n", node)
		t.Error("want:\n", want)
	}
}

func TestParseIf(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(if (cond1)
			(then1))`)

	node := requireItem[*ast.If](t, pkg.Nodes, 0, "parsed package")

	want := &ast.If{
		Cond: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "cond1"},
			},
		},
		Then: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "then1"},
			},
		},
	}

	assertEqual(t, want, node, "parsed if")
}

func TestParseIfElse(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(if (cond1)
			(then1)
			(else1))`)

	node := requireItem[*ast.If](t, pkg.Nodes, 0, "parsed package")

	want := &ast.If{
		Cond: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "cond1"},
			},
		},
		Then: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "then1"},
			},
		},
		Else: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "else1"},
			},
		},
	}

	assertEqual(t, want, node, "parsed if")
}

func TestParseCond(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		(cond (cond1)
			(then1))`)

	node := requireItem[*ast.If](t, pkg.Nodes, 0, "parsed package")

	want := &ast.If{
		Cond: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "cond1"},
			},
		},
		Then: &ast.SExp{
			Items: []ast.Node{
				&ast.Symbol{Value: "then1"},
			},
		},
	}

	assertEqual(t, want, node, "parsed if")
}

func TestParseFunctionDecl(t *testing.T) {
	t.Parallel()

	t.Run("named", func(t *testing.T) {
		t.Parallel()
		pkg := assertParse(t, `
		(*fn sum (a :_ int b :_ int) (int) (+ a b))
	`)

		node := requireItem[*ast.FunctionDecl](t, pkg.Nodes, 0, "parsed package")

		want := &ast.FunctionDecl{
			FnName: &ast.Symbol{Value: "sum"},
			Spec: &ast.FunctionSpec{
				Params: []*ast.FieldSpec{
					{Name: &ast.Symbol{Value: "a"}, Type: &ast.Symbol{Value: "int"}},
					{Name: &ast.Symbol{Value: "b"}, Type: &ast.Symbol{Value: "int"}},
				},
				Ret: []ast.Node{&ast.Symbol{Value: "int"}},
			},
			Body: &ast.SpecialOp{
				Op: "+",
				Items: []ast.Node{
					&ast.Symbol{Value: "a"},
					&ast.Symbol{Value: "b"},
				},
			},
		}

		assertEqual(t, want, node, "parsed function declaration")
	})

	t.Run("anonymous", func(t *testing.T) {
		t.Parallel()

		pkg := assertParse(t, `
			(*fn (a :_ int b :_ int) (int) (+ a b))
		`)

		node := requireItem[*ast.FunctionDecl](t, pkg.Nodes, 0, "parsed package")

		want := &ast.FunctionDecl{
			Spec: &ast.FunctionSpec{
				Params: []*ast.FieldSpec{
					{Name: &ast.Symbol{Value: "a"}, Type: &ast.Symbol{Value: "int"}},
					{Name: &ast.Symbol{Value: "b"}, Type: &ast.Symbol{Value: "int"}},
				},
				Ret: []ast.Node{&ast.Symbol{Value: "int"}},
			},
			Body: &ast.SpecialOp{
				Op: "+",
				Items: []ast.Node{
					&ast.Symbol{Value: "a"},
					&ast.Symbol{Value: "b"},
				},
			},
		}

		assertEqual(t, want, node, "parsed function declaration")
	})
}

func TestParseSpecialOperator(t *testing.T) {
	t.Parallel()

	// + - * /

	pkg := assertParse(t, `
		(+ 1 2 3 4)
		(* 1 2 3 4)

		(- 1 2)
		(/ 1 2)
	`)

	argumens := []ast.Node{
		&ast.Literal[int64]{Value: 1},
		&ast.Literal[int64]{Value: 2},
		&ast.Literal[int64]{Value: 3},
		&ast.Literal[int64]{Value: 4},
	}

	want := &ast.Package{
		Nodes: []ast.Node{
			&ast.SpecialOp{Op: "+", Items: argumens},
			&ast.SpecialOp{Op: "*", Items: argumens},
			&ast.SpecialOp{Op: "-", Items: argumens[:2]},
			&ast.SpecialOp{Op: "/", Items: argumens[:2]},
		},
	}

	assertEqual(t, want, pkg, "parsed special operators")
}

func TestParseDotSelector(t *testing.T) {
	t.Parallel()

	pkg := assertParse(t, `
		a.(x y)
	`)

	selector := requireItem[*ast.DotSelector](t, pkg.Nodes, 0, "parsed selector")

	a := &ast.Symbol{Value: "a"}
	x := &ast.Symbol{Value: "x"}
	y := &ast.Symbol{Value: "y"}
	want := &ast.DotSelector{
		Left:  a,
		Right: ast.NewSexp(x, y),
	}

	assertEqual(t, want, selector, "parsed dot selector")
}

func assertParse(t *testing.T, input string) *ast.Package {
	t.Helper()

	lex := lexer.NewLexer(t.Name(), strings.NewReader(input))

	par := parser.New(lex)

	pkg, err := par.Parse()

	if !assert.NoError(t, err, "parsing") {
		logLines(t, highlightErr(err, input))
		t.FailNow()
	}

	return pkg
}

func logLines(t *testing.T, lines []string) {
	t.Helper()

	for _, line := range lines {
		t.Log(line)
	}
}

func highlightErr(err error, input string) []string {
	var errs []error

	switch err := err.(type) {
	case nil:
		return nil
	case interface{ Unwrap() []error }:
		errs = err.Unwrap()
	default:
		errs = []error{err}
	}

	highlights := []string{}

	for _, err := range errs {
		var syntaxErr *parser.Error
		if errors.As(err, &syntaxErr) {
			highlights = append(highlights, highlight(input, syntaxErr.Pos)...)
		}
	}

	return highlights
}

func highlight(input string, pos tokens.Position) []string {
	log.Println("got position", pos)
	l, col := pos.Line, pos.Column

	lines := strings.Split(input, "\n")

	if l < 0 || l > len(lines) {
		log.Printf("<invalid line position %s: out of bounds 0..%d", pos, len(lines))
		return nil
	}

	line := strings.ReplaceAll(lines[l], "\t", " ")

	if col < 0 || col > len(line) {
		log.Printf("<invalid column position %s: out of bounds 0..%d", pos, len(line))
		return nil
	}

	leftpad := ""
	if col > 0 {
		leftpad = strings.Repeat(" ", col)
	}

	filling := ""
	if col < len(line)-1 {
		filling = strings.Repeat("~", len(line)-col-1)
	}

	// Creating the highlighted line
	highlightedLine := leftpad + "^" + filling

	return []string{line, highlightedLine}
}

func requireItem[N ast.Node](t *testing.T, items []ast.Node, index int, msg ...any) N {
	t.Helper()

	if index < 0 || index >= len(items) {
		t.Error(msg...)
		t.Fatalf("index out of bounds: %d", index)
	}

	item := items[index]
	got, ok := item.(N)

	if !ok {
		var want N
		t.Error(msg...)
		t.Fatalf("item at index %d is not a %T, but a %T", index, want, item)
	}

	return got
}

func assertEqual(t *testing.T, want, got ast.Node, msg ...any) {
	t.Helper()

	if !got.Equal(want) {
		t.Error(msg...)
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
