package compiler_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/ninedraft/sulisp/compiler"
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
	"github.com/ninedraft/sulisp/vm"
)

func TestCompileConstantsAndReuseTopLevelRegister(t *testing.T) {
	t.Parallel()

	nodes := []ast.Node{
		intAtom(42, 1),
		floatAtom(3.5, 2),
		stringAtom("hello", 3),
		boolAtom(true, 4),
		keywordAtom(":ok", 5),
	}
	want := vm.Program{
		Ops: vm.Tape{
			vm.LoadConst{Pos: position(1), Dst: 0, Val: vm.Int(42)},
			vm.LoadConst{Pos: position(2), Dst: 0, Val: vm.Float(3.5)},
			vm.LoadConst{Pos: position(3), Dst: 0, Val: vm.String("hello")},
			vm.LoadConst{Pos: position(4), Dst: 0, Val: vm.Bool(true)},
			vm.LoadConst{Pos: position(5), Dst: 0, Val: vm.Keyword(":ok")},
		},
		NumRegs: 1,
	}

	assertCompile(t, nodes, want)
}

func TestCompilePrintAndArithmeticLeftFold(t *testing.T) {
	t.Parallel()

	add := formAt("+", 4, intAtom(1, 6), intAtom(2, 8), intAtom(3, 10))
	print := formAt("print", 2, add)
	want := vm.Program{
		Ops: vm.Tape{
			vm.LoadConst{Pos: position(6), Dst: 0, Val: vm.Int(1)},
			vm.LoadConst{Pos: position(8), Dst: 1, Val: vm.Int(2)},
			vm.Add{Pos: position(4), Dst: 0, A: 0, B: 1},
			vm.LoadConst{Pos: position(10), Dst: 1, Val: vm.Int(3)},
			vm.Add{Pos: position(4), Dst: 0, A: 0, B: 1},
			vm.Print{Pos: position(2), Src: 0},
		},
		NumRegs: 2,
	}

	assertCompile(t, []ast.Node{print}, want)
}

func TestCompileArithmeticOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		form ast.List
		op   vm.Operation
	}{
		{name: "add", form: formAt("+", 2, intAtom(8, 4), intAtom(2, 6)), op: vm.Add{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: "sub", form: formAt("-", 2, intAtom(8, 4), intAtom(2, 6)), op: vm.Sub{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: "mul", form: formAt("*", 2, intAtom(8, 4), intAtom(2, 6)), op: vm.Mul{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: "div", form: formAt("/", 2, intAtom(8, 4), intAtom(2, 6)), op: vm.Div{Pos: position(2), Dst: 0, A: 0, B: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertCompile(t, []ast.Node{test.form}, vm.Program{
				Ops: vm.Tape{
					vm.LoadConst{Pos: position(4), Dst: 0, Val: vm.Int(8)},
					vm.LoadConst{Pos: position(6), Dst: 1, Val: vm.Int(2)},
					test.op,
				},
				NumRegs: 2,
			})
		})
	}
}

func TestCompileComparisons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		op   vm.Operation
	}{
		{name: "=", op: vm.Eq{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: "<", op: vm.Lt{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: ">", op: vm.Gt{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: "<=", op: vm.Le{Pos: position(2), Dst: 0, A: 0, B: 1}},
		{name: ">=", op: vm.Ge{Pos: position(2), Dst: 0, A: 0, B: 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertCompile(t, []ast.Node{formAt(test.name, 2, intAtom(1, 5), floatAtom(2, 7))}, vm.Program{
				Ops: vm.Tape{
					vm.LoadConst{Pos: position(5), Dst: 0, Val: vm.Int(1)},
					vm.LoadConst{Pos: position(7), Dst: 1, Val: vm.Float(2)},
					test.op,
				},
				NumRegs: 2,
			})
		})
	}
}

func TestCompileIf(t *testing.T) {
	t.Parallel()

	node := formAt("if", 2, boolAtom(true, 5), stringAtom("then", 10), stringAtom("else", 17))
	want := vm.Program{
		Ops: vm.Tape{
			vm.LoadConst{Pos: position(5), Dst: 1, Val: vm.Bool(true)},
			vm.JumpIfFalse{Pos: position(5), Cond: 1, Target: 4},
			vm.LoadConst{Pos: position(10), Dst: 0, Val: vm.String("then")},
			vm.Jump{Pos: position(2), Target: 5},
			vm.LoadConst{Pos: position(17), Dst: 0, Val: vm.String("else")},
		},
		NumRegs: 2,
	}

	assertCompile(t, []ast.Node{node}, want)
}

func TestCompileNotAndShortCircuitForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node ast.List
		want vm.Program
	}{
		{
			name: "not",
			node: formAt("not", 2, boolAtom(true, 6)),
			want: vm.Program{
				Ops: vm.Tape{
					vm.LoadConst{Pos: position(6), Dst: 0, Val: vm.Bool(true)},
					vm.Not{Pos: position(2), Dst: 0, Src: 0},
				},
				NumRegs: 1,
			},
		},
		{
			name: "and",
			node: formAt("and", 2, boolAtom(true, 6), boolAtom(false, 11)),
			want: vm.Program{
				Ops: vm.Tape{
					vm.LoadConst{Pos: position(6), Dst: 1, Val: vm.Bool(true)},
					vm.JumpIfFalse{Pos: position(6), Cond: 1, Target: 4},
					vm.LoadConst{Pos: position(11), Dst: 0, Val: vm.Bool(false)},
					vm.Jump{Pos: position(2), Target: 5},
					vm.LoadConst{Pos: position(2), Dst: 0, Val: vm.Bool(false)},
				},
				NumRegs: 2,
			},
		},
		{
			name: "or",
			node: formAt("or", 2, boolAtom(false, 5), boolAtom(true, 11)),
			want: vm.Program{
				Ops: vm.Tape{
					vm.LoadConst{Pos: position(5), Dst: 1, Val: vm.Bool(false)},
					vm.JumpIfFalse{Pos: position(5), Cond: 1, Target: 4},
					vm.LoadConst{Pos: position(2), Dst: 0, Val: vm.Bool(true)},
					vm.Jump{Pos: position(2), Target: 5},
					vm.LoadConst{Pos: position(11), Dst: 0, Val: vm.Bool(true)},
				},
				NumRegs: 2,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertCompile(t, []ast.Node{test.node}, test.want)
		})
	}
}

func TestCompileReusesArithmeticTemporary(t *testing.T) {
	t.Parallel()

	node := formAt("+", 2, intAtom(1, 4), intAtom(2, 6), intAtom(3, 8), intAtom(4, 10))
	program, err := compiler.Compile([]ast.Node{node})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if got, want := program.NumRegs, 2; got != want {
		t.Errorf("NumRegs = %d, want %d", got, want)
	}
	assertValidProgram(t, program)
}

func TestCompiledLogicShortCircuits(t *testing.T) {
	t.Parallel()

	danger := formAt("=", 12,
		formAt("/", 15, intAtom(1, 17), intAtom(0, 19)),
		intAtom(0, 22),
	)
	tests := []struct {
		name string
		node ast.List
		want string
	}{
		{name: "and", node: formAt("print", 2, formAt("and", 8, boolAtom(false, 12), danger)), want: "false\n"},
		{name: "or", node: formAt("print", 2, formAt("or", 8, boolAtom(true, 11), danger)), want: "true\n"},
		{
			name: "variadic and",
			node: formAt("print", 2, formAt("and", 8,
				boolAtom(true, 12), boolAtom(true, 17), boolAtom(false, 22), danger,
			)),
			want: "false\n",
		},
		{
			name: "variadic or",
			node: formAt("print", 2, formAt("or", 8,
				boolAtom(false, 11), boolAtom(false, 17), boolAtom(true, 23), danger,
			)),
			want: "true\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			program, err := compiler.Compile([]ast.Node{test.node})
			if err != nil {
				t.Fatalf("Compile() error = %v", err)
			}
			var output bytes.Buffer
			if err := vm.New(&output).Run(program); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("output = %q, want %q", got, test.want)
			}
		})
	}
}

func TestCompileErrorsArePositioned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node ast.Node
		want string
	}{
		{name: "empty list", node: ast.List{PosRange: sourceRange(1)}, want: "bad.su:1:1: empty form"},
		{name: "integer head", node: listAt(1, intAtom(1, 2)), want: "bad.su:1:2: form head must be a symbol, got integer"},
		{name: "float head", node: listAt(1, floatAtom(1, 2)), want: "bad.su:1:2: form head must be a symbol, got float"},
		{name: "string head", node: listAt(1, stringAtom("x", 2)), want: "bad.su:1:2: form head must be a symbol, got string"},
		{name: "boolean head", node: listAt(1, boolAtom(true, 2)), want: "bad.su:1:2: form head must be a symbol, got boolean"},
		{name: "keyword head", node: listAt(1, keywordAtom(":x", 2)), want: "bad.su:1:2: form head must be a symbol, got keyword"},
		{name: "list head", node: listAt(1, ast.List{PosRange: sourceRange(2)}), want: "bad.su:1:2: form head must be a symbol, got list"},
		{name: "unknown symbol", node: symbolAtom("missing", 3), want: "bad.su:1:3: unknown symbol \"missing\""},
		{name: "unknown form", node: formAt("missing", 2), want: "bad.su:1:2: unknown form \"missing\""},
		{name: "print arity", node: formAt("print", 2), want: "bad.su:1:2: print expects 1 argument, got 0"},
		{name: "arithmetic arity", node: formAt("+", 2, intAtom(1, 4)), want: "bad.su:1:2: + expects at least 2 arguments, got 1"},
		{name: "comparison arity", node: formAt("=", 2, intAtom(1, 4)), want: "bad.su:1:2: = expects 2 arguments, got 1"},
		{name: "if arity", node: formAt("if", 2, boolAtom(true, 5), intAtom(1, 10)), want: "bad.su:1:2: if expects 3 arguments (cond then else), got 2"},
		{name: "not arity", node: formAt("not", 2), want: "bad.su:1:2: not expects 1 argument, got 0"},
		{name: "and arity", node: formAt("and", 2, boolAtom(true, 6)), want: "bad.su:1:2: and expects at least 2 arguments, got 1"},
		{name: "or arity", node: formAt("or", 2), want: "bad.su:1:2: or expects at least 2 arguments, got 0"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := compiler.Compile([]ast.Node{test.node})
			if err == nil {
				t.Fatalf("Compile() error = nil, want %q", test.want)
			}
			if got := err.Error(); got != test.want {
				t.Errorf("Compile() error = %q, want %q", got, test.want)
			}
			compileError, ok := errors.AsType[*compiler.Error](err)
			if !ok || compileError.Pos.File != "bad.su" {
				t.Errorf("Compile() error = %#v, want positioned *compiler.Error", err)
			}
		})
	}
}

func TestEveryCompiledOperandAndJumpIsInBounds(t *testing.T) {
	t.Parallel()

	nodes := []ast.Node{
		formAt("print", 2, formAt("if", 8,
			formAt("and", 12, boolAtom(true, 16), formAt("<", 21, intAtom(1, 23), floatAtom(2, 25))),
			formAt("+", 30, intAtom(1, 32), intAtom(2, 34), intAtom(3, 36)),
			formAt("not", 40, formAt("or", 44, boolAtom(false, 47), boolAtom(true, 53))),
		)),
	}
	program, err := compiler.Compile(nodes)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	assertValidProgram(t, program)
}

func assertCompile(t *testing.T, nodes []ast.Node, want vm.Program) {
	t.Helper()

	got, err := compiler.Compile(nodes)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Compile() =\n%s\nregs=%d\nwant:\n%s\nregs=%d", got.Ops, got.NumRegs, want.Ops, want.NumRegs)
	}
	assertValidProgram(t, got)
}

func assertValidProgram(t *testing.T, program vm.Program) {
	t.Helper()

	register := func(at int, reg vm.Reg) {
		t.Helper()
		if reg < 0 || int(reg) >= program.NumRegs {
			t.Fatalf("operation %d: register r%d outside %d registers", at, reg, program.NumRegs)
		}
	}
	target := func(at, destination int) {
		t.Helper()
		if destination < 0 || destination > len(program.Ops) {
			t.Fatalf("operation %d: jump target @%d outside %d operations", at, destination, len(program.Ops))
		}
	}

	for at, operation := range program.Ops {
		switch operation := operation.(type) {
		case vm.LoadConst:
			register(at, operation.Dst)
		case vm.Move:
			register(at, operation.Dst)
			register(at, operation.Src)
		case vm.Add:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Sub:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Mul:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Div:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Eq:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Lt:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Gt:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Le:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Ge:
			registerBinary(t, at, program.NumRegs, operation.Dst, operation.A, operation.B)
		case vm.Not:
			register(at, operation.Dst)
			register(at, operation.Src)
		case vm.Jump:
			target(at, operation.Target)
		case vm.JumpIfFalse:
			register(at, operation.Cond)
			target(at, operation.Target)
		case vm.Print:
			register(at, operation.Src)
		default:
			t.Fatalf("operation %d has unchecked type %T", at, operation)
		}
	}
}

func registerBinary(t *testing.T, at, count int, registers ...vm.Reg) {
	t.Helper()
	for _, register := range registers {
		if register < 0 || int(register) >= count {
			t.Fatalf("operation %d: register r%d outside %d registers", at, register, count)
		}
	}
}

func position(column int) tokens.Position {
	return tokens.Position{File: "bad.su", Line: 1, Column: column}
}

func sourceRange(column int) ast.PosRange {
	return ast.PosRange{From: position(column), To: position(column + 1)}
}

func intAtom(value int64, column int) ast.Atom[int64] {
	return ast.Atom[int64]{PosRange: sourceRange(column), Kind: tokens.TokenInt, Value: value}
}

func floatAtom(value float64, column int) ast.Atom[float64] {
	return ast.Atom[float64]{PosRange: sourceRange(column), Kind: tokens.TokenFloat, Value: value}
}

func stringAtom(value string, column int) ast.Atom[string] {
	return ast.Atom[string]{PosRange: sourceRange(column), Kind: tokens.TokenStr, Value: value}
}

func boolAtom(value bool, column int) ast.Atom[bool] {
	return ast.Atom[bool]{PosRange: sourceRange(column), Value: value}
}

func symbolAtom(value string, column int) ast.Atom[string] {
	return ast.Atom[string]{PosRange: sourceRange(column), Kind: tokens.TokenSymbol, Value: value}
}

func keywordAtom(value string, column int) ast.Atom[string] {
	return ast.Atom[string]{PosRange: sourceRange(column), Kind: tokens.TokenKeyword, Value: value}
}

func formAt(name string, column int, arguments ...ast.Node) ast.List {
	items := make([]ast.Node, 1, len(arguments)+1)
	items[0] = symbolAtom(name, column)
	items = append(items, arguments...)
	return ast.List{PosRange: sourceRange(max(1, column-1)), Items: items}
}

func listAt(column int, items ...ast.Node) ast.List {
	return ast.List{PosRange: sourceRange(column), Items: items}
}
