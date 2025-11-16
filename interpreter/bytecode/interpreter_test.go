package bytecode

import (
	"testing"

	"github.com/ninedraft/sulisp/language/object"
)

func TestJumpIfTrueSkipsFalseBranch(t *testing.T) {
	vm := NewVM([]Command{
		Const(true),
		JumpIfTrue(3),
		Const(int64(99)), // should be skipped
		Const(int64(42)),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("unexpected vm error: %v", vm.Err)
	}

	result, ok := vm.Stack.Peek()
	if !ok {
		t.Fatalf("stack is empty, expected result")
	}

	value, ok := result.(*object.Primitive[int64])
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}

	if value.Value != 42 {
		t.Fatalf("expected 42 after jump, got %d", value.Value)
	}
}

func TestCallLoadsArgumentAndReturns(t *testing.T) {
	vm := NewVM([]Command{
		Const(int64(3)),
		Call(5, 1),
		Const(int64(5)),
		Add,
		Jump(9), // skip function definition after main
		LoadLocal(0),
		Const(int64(2)),
		Mul,
		Return,
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("unexpected vm error: %v", vm.Err)
	}

	result, ok := vm.Stack.Pop()
	if !ok {
		t.Fatalf("stack empty after program")
	}

	value, ok := result.(*object.Primitive[int64])
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}

	if value.Value != 11 {
		t.Fatalf("expected return value 11, got %d", value.Value)
	}
}

func TestStoreAndLoadLocal(t *testing.T) {
	vm := NewVM([]Command{
		Const(int64(7)),
		StoreLocal(0),
		LoadLocal(0),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("unexpected vm error: %v", vm.Err)
	}

	result, ok := vm.Stack.Pop()
	if !ok {
		t.Fatalf("stack empty after store/load")
	}

	value, ok := result.(*object.Primitive[int64])
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}

	if value.Value != 7 {
		t.Fatalf("expected loaded value 7, got %d", value.Value)
	}

	if vm.Stack.Len() != 0 {
		t.Fatalf("stack should be empty after pop, got len %d", vm.Stack.Len())
	}
}

func TestComparisonCommands(t *testing.T) {
	vm := NewVM([]Command{
		Const(int64(2)),
		Const(int64(3)),
		Less(),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}

	result, _ := vm.Stack.Pop()
	value, ok := result.(*object.Primitive[bool])
	if !ok || !value.Value {
		t.Fatalf("expected true for 2 < 3, got %v", result)
	}
}

func TestLogicalCommands(t *testing.T) {
	vm := NewVM([]Command{
		Const(true),
		Const(false),
		Or(),
	})
	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}
	result, _ := vm.Stack.Pop()
	value, ok := result.(*object.Primitive[bool])
	if !ok || !value.Value {
		t.Fatalf("expected true from true || false, got %v", result)
	}
}

func TestAndCommand(t *testing.T) {
	vm := NewVM([]Command{
		Const(true),
		Const(false),
		And(),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}

	result, _ := vm.Stack.Pop()
	value, ok := result.(*object.Primitive[bool])
	if !ok || value.Value {
		t.Fatalf("expected false from true && false, got %v", result)
	}
}

func TestNotCommand(t *testing.T) {
	vm := NewVM([]Command{
		Const(false),
		Not(),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}

	result, _ := vm.Stack.Pop()
	value, ok := result.(*object.Primitive[bool])
	if !ok || !value.Value {
		t.Fatalf("expected true from !false, got %v", result)
	}
}

func TestProgramUsesEveryInstruction(t *testing.T) {
	commands := []Command{
		Const(int64(2)),           // 0
		Const(int64(10)),          // 1
		Div,                       // 2
		Const(int64(5)),           // 3
		Add,                       // 4
		Const(int64(3)),           // 5
		Mul,                       // 6
		Const(int64(5)),           // 7
		Sub,                       // 8
		Const(int64(-25)),         // 9
		Equal,                     // 10
		JumpIfTrue(14),            // 11
		Const(int64(0)),           // 12
		Jump(15),                  // 13
		Const(int64(256)),         // 14
		StoreLocal(0),             // 15
		LoadLocal(0),              // 16
		Const(int64(1)),           // 17
		Swap(),                    // 18
		Sub,                       // 19
		Dup(),                     // 20
		Pop(),                     // 21
		Const(int64(255)),         // 22
		Equal,                     // 23
		JumpIfFalse(25),           // 24
		Const(int64(5)),           // 25
		Const(int64(10)),          // 26
		Greater(),                 // 27
		Const(int64(3)),           // 28
		Const(int64(3)),           // 29
		LessEqual(),               // 30
		Const(int64(5)),           // 31
		Const(int64(2)),           // 32
		Less(),                    // 33
		Const(int64(1)),           // 34
		Const(int64(2)),           // 35
		GreaterEqual(),            // 36
		Or(),                      // 37
		And(),                     // 38
		Not(),                     // 39
		And(),                     // 40
		JumpIfFalse(42),           // 41
		ArrayNew,                  // 42
		Const(int64(7)),           // 43
		ArrayAppend,               // 44
		Const(int64(8)),           // 45
		ArrayAppend,               // 46
		ArrayShuffle,              // 47
		ArrayLen,                  // 48
		Pop(),                     // 49
		Const(int64(3)),           // 50
		Call(53, 1),               // 51
		Halt(),                    // 52
		LoadLocal(0),              // 53 (function start)
		Const(int64(4)),           // 54
		Add,                       // 55
		StoreLocal(0),             // 56
		LoadLocal(0),              // 57
		Return,                    // 58
	}

	vm := NewVM(commands)
	vm.Run()
	if vm.Err != nil {
		t.Fatalf("unexpected vm error: %v", vm.Err)
	}

	result, ok := vm.Stack.Pop()
	if !ok {
		t.Fatalf("stack empty after program")
	}

	value, ok := result.(*object.Primitive[int64])
	if !ok {
		t.Fatalf("expected integer result, got %T", result)
	}

	if value.Value != 7 {
		t.Fatalf("expected return value 7, got %d", value.Value)
	}

	if vm.Stack.Len() != 0 {
		t.Fatalf("stack should be empty after pop, got len %d", vm.Stack.Len())
	}
}

func TestStackUtilities(t *testing.T) {
	vm := NewVM([]Command{
		Const(int64(1)),
		Const(int64(2)),
		Swap(),
		Pop(),
		Dup(),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}

	result, _ := vm.Stack.Pop()
	value, ok := result.(*object.Primitive[int64])
	if !ok || value.Value != 2 {
		t.Fatalf("expected final value 2 after swap/pop/dup, got %v", result)
	}
}

func TestHaltStopsExecution(t *testing.T) {
	vm := NewVM([]Command{
		Const(int64(5)),
		Halt(),
		Const(int64(6)),
	})

	vm.Run()
	if vm.Err != nil {
		t.Fatalf("vm error: %v", vm.Err)
	}

	if vm.Stack.Len() != 1 {
		t.Fatalf("expected only one value after halt, got %d", vm.Stack.Len())
	}
}
