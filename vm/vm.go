// Package vm implements Sulisp's register virtual machine.
package vm

import (
	"cmp"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ninedraft/sulisp/language/tokens"
)

// Kind discriminates runtime value types.
type Kind uint8

const (
	KindNumber Kind = iota
	KindString
	KindBool
	KindKeyword
)

func (kind Kind) String() string {
	switch kind {
	case KindNumber:
		return "number"
	case KindString:
		return "string"
	case KindBool:
		return "bool"
	case KindKeyword:
		return "keyword"
	default:
		return "unknown"
	}
}

// Value is a value held in a VM register.
type Value interface {
	fmt.Stringer
	Kind() Kind
	isValue()
}

// Number represents either an integer or a floating-point number. IsFloat
// preserves which representation produced the value even though both share
// one runtime and future static type.
type Number struct {
	Integer int64
	Float   float64
	IsFloat bool
}

// Int constructs an integer-representation number.
func Int(value int64) Number {
	return Number{Integer: value}
}

// Float constructs a floating-point-representation number.
func Float(value float64) Number {
	return Number{Float: value, IsFloat: true}
}

func (Number) Kind() Kind { return KindNumber }
func (Number) isValue()   {}

func (number Number) String() string {
	if number.IsFloat {
		return strconv.FormatFloat(number.Float, 'g', -1, 64)
	}
	return strconv.FormatInt(number.Integer, 10)
}

func (number Number) float64() float64 {
	if number.IsFloat {
		return number.Float
	}
	return float64(number.Integer)
}

// String is a runtime string.
type String string

func (String) Kind() Kind { return KindString }
func (String) isValue()   {}
func (value String) String() string {
	return string(value)
}

// Bool is a runtime boolean.
type Bool bool

func (Bool) Kind() Kind { return KindBool }
func (Bool) isValue()   {}
func (value Bool) String() string {
	return strconv.FormatBool(bool(value))
}

// Keyword is a runtime keyword.
type Keyword string

func (Keyword) Kind() Kind { return KindKeyword }
func (Keyword) isValue()   {}
func (value Keyword) String() string {
	return string(value)
}

// Reg is an index in the current register file.
type Reg int

// Operation is one VM instruction.
type Operation interface {
	fmt.Stringer
	Execute(*VM) error
}

// Tape is a sequence of operations.
type Tape []Operation

func (tape Tape) String() string {
	lines := make([]string, len(tape))
	for index, operation := range tape {
		lines[index] = operation.String()
	}
	return strings.Join(lines, "\n")
}

// Program is a compiled unit ready for execution.
type Program struct {
	Ops     Tape
	NumRegs int
}

// VM executes Programs.
type VM struct {
	ops  Tape
	regs []Value
	pc   int
	out  io.Writer
}

// New creates a VM that writes print output to out.
func New(out io.Writer) *VM {
	if out == nil {
		out = io.Discard
	}
	return &VM{out: out}
}

// Run executes program from its first operation.
func (machine *VM) Run(program Program) error {
	if program.NumRegs < 0 {
		return fmt.Errorf("vm: negative register count %d", program.NumRegs)
	}

	machine.ops = program.Ops
	machine.regs = make([]Value, program.NumRegs)
	machine.pc = 0
	for machine.pc < len(machine.ops) {
		operation := machine.ops[machine.pc]
		machine.pc++
		if err := operation.Execute(machine); err != nil {
			return err
		}
	}
	return nil
}

// Error is a runtime failure at a source position.
type Error struct {
	Pos tokens.Position
	Msg string
}

func (runtimeError *Error) Error() string {
	return fmt.Sprintf("%s: %s", runtimeError.Pos, runtimeError.Msg)
}

// LoadConst stores Val in Dst.
type LoadConst struct {
	Pos tokens.Position
	Dst Reg
	Val Value
}

func (operation LoadConst) String() string {
	return fmt.Sprintf("r%d = const %s", operation.Dst, operation.Val)
}

func (operation LoadConst) Execute(machine *VM) error {
	machine.regs[operation.Dst] = operation.Val
	return nil
}

// Move copies Src into Dst.
type Move struct {
	Pos      tokens.Position
	Dst, Src Reg
}

func (operation Move) String() string {
	return fmt.Sprintf("r%d = r%d", operation.Dst, operation.Src)
}

func (operation Move) Execute(machine *VM) error {
	machine.regs[operation.Dst] = machine.regs[operation.Src]
	return nil
}

func (machine *VM) numberPair(position tokens.Position, name string, a, b Reg) (Number, Number, error) {
	left, leftOK := machine.regs[a].(Number)
	right, rightOK := machine.regs[b].(Number)
	if !leftOK || !rightOK {
		return Number{}, Number{}, &Error{
			Pos: position,
			Msg: fmt.Sprintf("%s expects number operands, got %s and %s",
				name, kindOf(machine.regs[a]), kindOf(machine.regs[b])),
		}
	}
	return left, right, nil
}

func kindOf(value Value) Kind {
	if value == nil {
		return Kind(255)
	}
	return value.Kind()
}

func binaryNumber(
	machine *VM,
	position tokens.Position,
	name string,
	dst, a, b Reg,
	integers func(int64, int64) int64,
	floats func(float64, float64) float64,
) error {
	left, right, err := machine.numberPair(position, name, a, b)
	if err != nil {
		return err
	}
	if left.IsFloat || right.IsFloat {
		machine.regs[dst] = Float(floats(left.float64(), right.float64()))
	} else {
		machine.regs[dst] = Int(integers(left.Integer, right.Integer))
	}
	return nil
}

// Add stores A + B in Dst.
type Add struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Add) String() string {
	return fmt.Sprintf("r%d = add r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Add) Execute(machine *VM) error {
	return binaryNumber(machine, operation.Pos, "+", operation.Dst, operation.A, operation.B,
		func(a, b int64) int64 { return a + b },
		func(a, b float64) float64 { return a + b },
	)
}

// Sub stores A - B in Dst.
type Sub struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Sub) String() string {
	return fmt.Sprintf("r%d = sub r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Sub) Execute(machine *VM) error {
	return binaryNumber(machine, operation.Pos, "-", operation.Dst, operation.A, operation.B,
		func(a, b int64) int64 { return a - b },
		func(a, b float64) float64 { return a - b },
	)
}

// Mul stores A * B in Dst.
type Mul struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Mul) String() string {
	return fmt.Sprintf("r%d = mul r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Mul) Execute(machine *VM) error {
	return binaryNumber(machine, operation.Pos, "*", operation.Dst, operation.A, operation.B,
		func(a, b int64) int64 { return a * b },
		func(a, b float64) float64 { return a * b },
	)
}

// Div stores A / B in Dst. Integer-only division truncates toward zero.
type Div struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Div) String() string {
	return fmt.Sprintf("r%d = div r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Div) Execute(machine *VM) error {
	left, right, err := machine.numberPair(operation.Pos, "/", operation.A, operation.B)
	if err != nil {
		return err
	}
	if (!right.IsFloat && right.Integer == 0) || (right.IsFloat && right.Float == 0) {
		return &Error{Pos: operation.Pos, Msg: "division by zero"}
	}
	if left.IsFloat || right.IsFloat {
		machine.regs[operation.Dst] = Float(left.float64() / right.float64())
	} else {
		machine.regs[operation.Dst] = Int(left.Integer / right.Integer)
	}
	return nil
}

func equalValues(left, right Value) (bool, bool) {
	if leftNumber, ok := left.(Number); ok {
		rightNumber, rightOK := right.(Number)
		if !rightOK {
			return false, false
		}
		if leftNumber.IsFloat || rightNumber.IsFloat {
			return leftNumber.float64() == rightNumber.float64(), true
		}
		return leftNumber.Integer == rightNumber.Integer, true
	}
	if left.Kind() != right.Kind() {
		return false, false
	}
	switch left := left.(type) {
	case String:
		return left == right.(String), true
	case Bool:
		return left == right.(Bool), true
	case Keyword:
		return left == right.(Keyword), true
	default:
		return false, false
	}
}

// Eq stores whether A and B are equal in Dst. Numeric equality crosses
// integer and floating-point representations.
type Eq struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Eq) String() string {
	return fmt.Sprintf("r%d = eq r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Eq) Execute(machine *VM) error {
	left, right := machine.regs[operation.A], machine.regs[operation.B]
	equal, comparable := equalValues(left, right)
	if !comparable {
		return &Error{
			Pos: operation.Pos,
			Msg: fmt.Sprintf("= expects operands of the same type, got %s and %s",
				kindOf(left), kindOf(right)),
		}
	}
	machine.regs[operation.Dst] = Bool(equal)
	return nil
}

func compareOrdered(left, right Value) (int, bool) {
	switch left := left.(type) {
	case Number:
		right, ok := right.(Number)
		if !ok {
			return 0, false
		}
		if !left.IsFloat && !right.IsFloat {
			return cmp.Compare(left.Integer, right.Integer), true
		}
		return cmp.Compare(left.float64(), right.float64()), true
	case String:
		right, ok := right.(String)
		if !ok {
			return 0, false
		}
		return strings.Compare(string(left), string(right)), true
	case Keyword:
		right, ok := right.(Keyword)
		if !ok {
			return 0, false
		}
		return strings.Compare(string(left), string(right)), true
	default:
		return 0, false
	}
}

func compare(
	machine *VM,
	position tokens.Position,
	name string,
	dst, a, b Reg,
	predicate func(int) bool,
) error {
	left, right := machine.regs[a], machine.regs[b]
	ordering, ok := compareOrdered(left, right)
	if !ok {
		return &Error{
			Pos: position,
			Msg: fmt.Sprintf("%s expects ordered operands, got %s and %s",
				name, kindOf(left), kindOf(right)),
		}
	}
	machine.regs[dst] = Bool(predicate(ordering))
	return nil
}

// Lt stores A < B in Dst.
type Lt struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Lt) String() string {
	return fmt.Sprintf("r%d = lt r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Lt) Execute(machine *VM) error {
	return compare(machine, operation.Pos, "<", operation.Dst, operation.A, operation.B,
		func(ordering int) bool { return ordering < 0 })
}

// Gt stores A > B in Dst.
type Gt struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Gt) String() string {
	return fmt.Sprintf("r%d = gt r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Gt) Execute(machine *VM) error {
	return compare(machine, operation.Pos, ">", operation.Dst, operation.A, operation.B,
		func(ordering int) bool { return ordering > 0 })
}

// Le stores A <= B in Dst.
type Le struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Le) String() string {
	return fmt.Sprintf("r%d = le r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Le) Execute(machine *VM) error {
	return compare(machine, operation.Pos, "<=", operation.Dst, operation.A, operation.B,
		func(ordering int) bool { return ordering <= 0 })
}

// Ge stores A >= B in Dst.
type Ge struct {
	Pos       tokens.Position
	Dst, A, B Reg
}

func (operation Ge) String() string {
	return fmt.Sprintf("r%d = ge r%d, r%d", operation.Dst, operation.A, operation.B)
}

func (operation Ge) Execute(machine *VM) error {
	return compare(machine, operation.Pos, ">=", operation.Dst, operation.A, operation.B,
		func(ordering int) bool { return ordering >= 0 })
}

// Not stores the boolean negation of Src in Dst.
type Not struct {
	Pos      tokens.Position
	Dst, Src Reg
}

func (operation Not) String() string {
	return fmt.Sprintf("r%d = not r%d", operation.Dst, operation.Src)
}

func (operation Not) Execute(machine *VM) error {
	value, ok := machine.regs[operation.Src].(Bool)
	if !ok {
		return &Error{
			Pos: operation.Pos,
			Msg: fmt.Sprintf("not expects bool operand, got %s", kindOf(machine.regs[operation.Src])),
		}
	}
	machine.regs[operation.Dst] = !value
	return nil
}

// Jump transfers execution to Target.
type Jump struct {
	Pos    tokens.Position
	Target int
}

func (operation Jump) String() string {
	return fmt.Sprintf("jump @%d", operation.Target)
}

func (operation Jump) Execute(machine *VM) error {
	machine.pc = operation.Target
	return nil
}

// JumpIfFalse transfers execution to Target if Cond contains false.
type JumpIfFalse struct {
	Pos    tokens.Position
	Cond   Reg
	Target int
}

func (operation JumpIfFalse) String() string {
	return fmt.Sprintf("jump_if_false r%d, @%d", operation.Cond, operation.Target)
}

func (operation JumpIfFalse) Execute(machine *VM) error {
	condition, ok := machine.regs[operation.Cond].(Bool)
	if !ok {
		return &Error{
			Pos: operation.Pos,
			Msg: fmt.Sprintf("condition expects bool, got %s", kindOf(machine.regs[operation.Cond])),
		}
	}
	if !condition {
		machine.pc = operation.Target
	}
	return nil
}

// Print writes Src followed by a newline.
type Print struct {
	Pos tokens.Position
	Src Reg
}

func (operation Print) String() string {
	return fmt.Sprintf("print r%d", operation.Src)
}

func (operation Print) Execute(machine *VM) error {
	if _, err := fmt.Fprintln(machine.out, machine.regs[operation.Src]); err != nil {
		return &Error{Pos: operation.Pos, Msg: fmt.Sprintf("print: %v", err)}
	}
	return nil
}
