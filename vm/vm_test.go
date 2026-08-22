package vm

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/language/tokens"
)

func TestRuntimeValues(t *testing.T) {
	t.Parallel()

	integer := Int(3)
	if integer.Kind() != KindNumber || integer.IsFloat {
		t.Fatalf("Int(3) = %#v, want integer number", integer)
	}
	floating := Float(3)
	if floating.Kind() != KindNumber || !floating.IsFloat {
		t.Fatalf("Float(3) = %#v, want floating-point number", floating)
	}

	values := []struct {
		value Value
		kind  Kind
		text  string
	}{
		{value: integer, kind: KindNumber, text: "3"},
		{value: floating, kind: KindNumber, text: "3"},
		{value: String("hello"), kind: KindString, text: "hello"},
		{value: Bool(true), kind: KindBool, text: "true"},
		{value: Keyword(":hello"), kind: KindKeyword, text: ":hello"},
	}
	for _, test := range values {
		if got := test.value.Kind(); got != test.kind {
			t.Errorf("%T.Kind() = %v, want %v", test.value, got, test.kind)
		}
		if got := test.value.String(); got != test.text {
			t.Errorf("%T.String() = %q, want %q", test.value, got, test.text)
		}
	}
}

func TestRunPrimitives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		prog Program
		want string
	}{
		{
			name: "print every value kind",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Int(42)}, Print{Src: 0},
					LoadConst{Dst: 0, Val: Float(3.5)}, Print{Src: 0},
					LoadConst{Dst: 0, Val: String("hello")}, Print{Src: 0},
					LoadConst{Dst: 0, Val: Bool(true)}, Print{Src: 0},
					LoadConst{Dst: 0, Val: Keyword(":ok")}, Print{Src: 0},
				},
				NumRegs: 1,
			},
			want: "42\n3.5\nhello\ntrue\n:ok\n",
		},
		{
			name: "integer arithmetic",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Int(20)},
					LoadConst{Dst: 1, Val: Int(5)},
					Add{Dst: 0, A: 0, B: 1},
					Sub{Dst: 0, A: 0, B: 1},
					Mul{Dst: 0, A: 0, B: 1},
					Div{Dst: 0, A: 0, B: 1},
					Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "20\n",
		},
		{
			name: "integer division truncates toward zero",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Int(-7)},
					LoadConst{Dst: 1, Val: Int(2)},
					Div{Dst: 0, A: 0, B: 1},
					Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "-3\n",
		},
		{
			name: "mixed arithmetic promotes to float",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Int(7)},
					LoadConst{Dst: 1, Val: Float(2)},
					Div{Dst: 0, A: 0, B: 1},
					Print{Src: 0},
					Add{Dst: 0, A: 0, B: 1},
					Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "3.5\n5.5\n",
		},
		{
			name: "comparisons",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Int(2)},
					LoadConst{Dst: 1, Val: Float(2)},
					Eq{Dst: 2, A: 0, B: 1}, Print{Src: 2},
					Lt{Dst: 2, A: 0, B: 1}, Print{Src: 2},
					Le{Dst: 2, A: 0, B: 1}, Print{Src: 2},
					Gt{Dst: 2, A: 0, B: 1}, Print{Src: 2},
					Ge{Dst: 2, A: 0, B: 1}, Print{Src: 2},
				},
				NumRegs: 3,
			},
			want: "true\nfalse\ntrue\nfalse\ntrue\n",
		},
		{
			name: "string ordering",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: String("a")},
					LoadConst{Dst: 1, Val: String("b")},
					Lt{Dst: 0, A: 0, B: 1}, Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "true\n",
		},
		{
			name: "keyword ordering",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Keyword(":z")},
					LoadConst{Dst: 1, Val: Keyword(":a")},
					Gt{Dst: 0, A: 0, B: 1}, Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "true\n",
		},
		{
			name: "boolean equality and not",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Bool(true)},
					LoadConst{Dst: 1, Val: Bool(false)},
					Eq{Dst: 0, A: 0, B: 1}, Print{Src: 0},
					Not{Dst: 0, Src: 0}, Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "false\ntrue\n",
		},
		{
			name: "string and keyword equality",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: String("same")},
					LoadConst{Dst: 1, Val: String("same")},
					Eq{Dst: 0, A: 0, B: 1}, Print{Src: 0},
					LoadConst{Dst: 0, Val: Keyword(":left")},
					LoadConst{Dst: 1, Val: Keyword(":right")},
					Eq{Dst: 0, A: 0, B: 1}, Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "true\nfalse\n",
		},
		{
			name: "move",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 1, Val: String("copied")},
					Move{Dst: 0, Src: 1},
					Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "copied\n",
		},
		{
			name: "conditional true branch",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Bool(true)},
					JumpIfFalse{Cond: 0, Target: 4},
					LoadConst{Dst: 0, Val: String("then")},
					Jump{Target: 5},
					LoadConst{Dst: 0, Val: String("else")},
					Print{Src: 0},
				},
				NumRegs: 1,
			},
			want: "then\n",
		},
		{
			name: "conditional false branch",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Bool(false)},
					JumpIfFalse{Cond: 0, Target: 4},
					LoadConst{Dst: 0, Val: String("then")},
					Jump{Target: 5},
					LoadConst{Dst: 0, Val: String("else")},
					Print{Src: 0},
				},
				NumRegs: 1,
			},
			want: "else\n",
		},
		{
			name: "jump provides short circuit",
			prog: Program{
				Ops: Tape{
					LoadConst{Dst: 0, Val: Bool(false)},
					JumpIfFalse{Cond: 0, Target: 6},
					LoadConst{Dst: 0, Val: Int(1)},
					LoadConst{Dst: 1, Val: Int(0)},
					Div{Dst: 0, A: 0, B: 1},
					Jump{Target: 6},
					Print{Src: 0},
				},
				NumRegs: 2,
			},
			want: "false\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := New(&output).Run(test.prog); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
			if got := output.String(); got != test.want {
				t.Errorf("output = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRunErrorsRetainSourcePositions(t *testing.T) {
	t.Parallel()

	at := tokens.Position{File: "bad.su", Line: 3, Column: 5}
	tests := []struct {
		name string
		ops  Tape
		regs int
		want string
	}{
		{
			name: "arithmetic type error",
			ops: Tape{
				LoadConst{Dst: 0, Val: Int(1)},
				LoadConst{Dst: 1, Val: String("x")},
				Add{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: + expects number operands, got number and string",
		},
		{
			name: "integer division by zero",
			ops: Tape{
				LoadConst{Dst: 0, Val: Int(1)},
				LoadConst{Dst: 1, Val: Int(0)},
				Div{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: division by zero",
		},
		{
			name: "float division by zero",
			ops: Tape{
				LoadConst{Dst: 0, Val: Float(1)},
				LoadConst{Dst: 1, Val: Float(0)},
				Div{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: division by zero",
		},
		{
			name: "equality type error",
			ops: Tape{
				LoadConst{Dst: 0, Val: Bool(true)},
				LoadConst{Dst: 1, Val: String("true")},
				Eq{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: = expects operands of the same type, got bool and string",
		},
		{
			name: "ordering boolean error",
			ops: Tape{
				LoadConst{Dst: 0, Val: Bool(false)},
				LoadConst{Dst: 1, Val: Bool(true)},
				Lt{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: < expects ordered operands, got bool and bool",
		},
		{
			name: "ordering different types",
			ops: Tape{
				LoadConst{Dst: 0, Val: String("a")},
				LoadConst{Dst: 1, Val: Keyword(":a")},
				Ge{Pos: at, Dst: 0, A: 0, B: 1},
			},
			regs: 2,
			want: "bad.su:3:5: >= expects ordered operands, got string and keyword",
		},
		{
			name: "condition type error",
			ops: Tape{
				LoadConst{Dst: 0, Val: Int(1)},
				JumpIfFalse{Pos: at, Cond: 0, Target: 2},
			},
			regs: 1,
			want: "bad.su:3:5: condition expects bool, got number",
		},
		{
			name: "not type error",
			ops: Tape{
				LoadConst{Dst: 0, Val: Keyword(":x")},
				Not{Pos: at, Dst: 0, Src: 0},
			},
			regs: 1,
			want: "bad.su:3:5: not expects bool operand, got keyword",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := New(io.Discard).Run(Program{Ops: test.ops, NumRegs: test.regs})
			if err == nil {
				t.Fatalf("Run() error = nil, want %q", test.want)
			}
			if got := err.Error(); got != test.want {
				t.Errorf("Run() error = %q, want %q", got, test.want)
			}
			runtimeError, ok := errors.AsType[*Error](err)
			if !ok || runtimeError.Pos != at {
				t.Errorf("Run() error = %#v, want positioned *Error", err)
			}
		})
	}
}

func TestPrintErrorRetainsSourcePosition(t *testing.T) {
	t.Parallel()

	at := tokens.Position{File: "output.su", Line: 8, Column: 2}
	err := New(failingWriter{}).Run(Program{
		Ops: Tape{
			LoadConst{Dst: 0, Val: String("hello")},
			Print{Pos: at, Src: 0},
		},
		NumRegs: 1,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want output error")
	}
	if got, want := err.Error(), "output.su:8:2: print: write failed"; got != want {
		t.Errorf("Run() error = %q, want %q", got, want)
	}
}

func TestTapeString(t *testing.T) {
	t.Parallel()

	tape := Tape{
		LoadConst{Dst: 0, Val: Int(42)},
		Move{Dst: 1, Src: 0},
		Add{Dst: 1, A: 1, B: 0},
		JumpIfFalse{Cond: 1, Target: 5},
		Jump{Target: 6},
		Print{Src: 1},
	}
	want := strings.Join([]string{
		"r0 = const 42",
		"r1 = r0",
		"r1 = add r1, r0",
		"jump_if_false r1, @5",
		"jump @6",
		"print r1",
	}, "\n")
	if got := tape.String(); got != want {
		t.Errorf("Tape.String() = %q, want %q", got, want)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
