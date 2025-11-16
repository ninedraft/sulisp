package compiler_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/compiler"
	"github.com/ninedraft/sulisp/interpreter/bytecode"
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/object"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/stretchr/testify/require"
)

func TestCompileLiteral(t *testing.T) {
	pkg := &ast.Package{
		Nodes: []ast.Node{
			&ast.Literal[int64]{Value: 7},
		},
	}

	vm := runCompiled(t, pkg)

	value := popInt(t, vm)
	require.Equal(t, int64(7), value)
}

func TestCompileAddition(t *testing.T) {
	pkg := parsePackage(t, "(+ 1 2 3)")

	vm := runCompiled(t, pkg)

	require.Equal(t, int64(6), popInt(t, vm))
}

func TestCompileIf(t *testing.T) {
	pkg := parsePackage(t, "(if true 1 2)")

	vm := runCompiled(t, pkg)

	require.Equal(t, int64(1), popInt(t, vm))
}

func TestCompileMatchLiteral(t *testing.T) {
	src := `
		(fn choose (x)
			(match x
				(1 10)
				(_ 20)))

		(choose 1)
	`

	vm := runSource(t, src)
	require.Equal(t, int64(10), popInt(t, vm))
}

func TestCompileMatchDefaultNull(t *testing.T) {
	src := `
		(fn choose (x)
			(match x
				(1 10)
				(2 20)))

		(choose 3)
	`

	vm := runSource(t, src)

	result, ok := vm.Stack.Pop()
	require.True(t, ok, "stack should contain a result")
	_, isNull := result.(*object.Null)
	require.True(t, isNull, "match should return null when no case matches")
}

func TestCompileMatchVariableBinding(t *testing.T) {
	src := `
		(fn choose (x)
			(match x
				(x (+ x 1))
				(_ 0)))

		(choose 5)
	`

	vm := runSource(t, src)
	require.Equal(t, int64(6), popInt(t, vm))
}

func TestCompileMatchLiteralPrecedence(t *testing.T) {
	src := `
		(fn choose (x)
			(match x
				(1 100)
				(_ 200)))

		(choose 2)
	`

	vm := runSource(t, src)
	require.Equal(t, int64(200), popInt(t, vm))
}

func TestCompileFunctionCall(t *testing.T) {
	fn := &ast.Function{
		Identifier: "square",
		Parameters: []*ast.Symbol{
			{Value: "x"},
		},
		Body: &ast.SpecialOp{
			Op: "*",
			Items: []ast.Node{
				&ast.Symbol{Value: "x"},
				&ast.Symbol{Value: "x"},
			},
		},
	}

	call := &ast.SExp{
		Items: []ast.Node{
			&ast.Symbol{Value: "square"},
			&ast.Literal[int64]{Value: 5},
		},
	}

	pkg := &ast.Package{
		Nodes: []ast.Node{
			fn,
			call,
		},
	}

	vm := runCompiled(t, pkg)
	require.Equal(t, int64(25), popInt(t, vm))
}

func TestCompileFunctionWithLocals(t *testing.T) {
	fn := &ast.Function{
		Identifier: "double",
		Parameters: []*ast.Symbol{
			{Value: "x"},
		},
		Body: &ast.SpecialOp{
			Op: "+",
			Items: []ast.Node{
				&ast.Assign{
					Target: &ast.Symbol{Value: "sum"},
					Value: &ast.SpecialOp{
						Op: "*",
						Items: []ast.Node{
							&ast.Symbol{Value: "x"},
							&ast.Literal[int64]{Value: 2},
						},
					},
				},
				&ast.Literal[int64]{Value: 0},
			},
		},
	}

	call := &ast.SExp{
		Items: []ast.Node{
			&ast.Symbol{Value: "double"},
			&ast.Literal[int64]{Value: 3},
		},
	}

	pkg := &ast.Package{
		Nodes: []ast.Node{
			call,
			fn,
		},
	}

	vm := runCompiled(t, pkg)
	require.Equal(t, int64(6), popInt(t, vm))
}

func TestCompileLetCreatesFreshSlot(t *testing.T) {
	src := `
		(fn double (x)
			(let ((x (+ x 1)))
				(+ x 2)))

		(double 5)
	`

	vm := runSource(t, src)
	require.Equal(t, int64(8), popInt(t, vm))
}

func TestCompileLetRestoresOuterVariable(t *testing.T) {
	src := `
		(fn shadow (x)
			(begin
				(let ((x (+ x 1)))
					x)
				x))

		(shadow 3)
	`

	vm := runSource(t, src)
	require.Equal(t, int64(3), popInt(t, vm))
}

func TestCompileHandleInjectsContinuation(t *testing.T) {
	src := `
		(handle io
			((print (fn print-handler (msg k) (k msg))))
			(io print "hi"))
	`

	pkg := parsePackage(t, src)

	comp := compiler.New()
	commands, err := comp.Compile(pkg)
	require.NoError(t, err)

	callIdx := -1
	for i, cmd := range commands {
		if strings.HasPrefix(cmd.Repr, "Call(") {
			callIdx = i
			break
		}
	}

	require.NotEqual(t, -1, callIdx, "expected a call command")
	require.Greater(t, callIdx, 0, "call must have preceding command")
	require.Equal(t, "PushCont", commands[callIdx-1].Repr, "effect call should push continuation")
}

func TestCompileErrorEffect(t *testing.T) {
	src := `
		(fn main ()
			(handle error
				((raise (fn raise-handler (msg k) (k msg))))
				(let ((value (error raise "boom")))
					value)))

		(main)
	`

	vm := runSource(t, src)
	result, ok := vm.Stack.Pop()
	require.True(t, ok)

	str, ok := result.(*object.Primitive[string])
	require.True(t, ok)
	require.Equal(t, `"boom"`, str.Value)
}

func runCompiled(t *testing.T, pkg *ast.Package) *bytecode.VM {
	t.Helper()

	comp := compiler.New()

	commands, err := comp.Compile(pkg)
	require.NoError(t, err)

	vm := bytecode.NewVM(commands)
	vm.Run()
	require.NoError(t, vm.Err)

	return vm
}

func parsePackage(t *testing.T, src string) *ast.Package {
	t.Helper()

	lex := lexer.NewLexer("test", strings.NewReader(src))
	p := parser.New(lex)
	pkg, err := p.Parse()
	require.NoError(t, err)
	return pkg
}

func popInt(t *testing.T, vm *bytecode.VM) int64 {
	t.Helper()

	obj, ok := vm.Stack.Pop()
	require.True(t, ok, "stack should not be empty")

	prim, ok := obj.(*object.Primitive[int64])
	require.True(t, ok, "expected primitive int")

	return prim.Value
}

func runSource(t *testing.T, src string) *bytecode.VM {
	t.Helper()

	lex := lexer.NewLexer("test", strings.NewReader(src))
	p := parser.New(lex)
	pkg, err := p.Parse()
	require.NoError(t, err)

	return runCompiled(t, pkg)
}
