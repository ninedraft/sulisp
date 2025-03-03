package bytecode

import (
	"github.com/ninedraft/sulisp/internal/collections/stack"
	"github.com/ninedraft/sulisp/language/object"
)

type VM struct {
	Stack stack.Stack[object.Object]
	Tape  []Command
	PC    PC
	Err   error
}

type PC int

type Command struct {
	Repr    string
	Execute func(vm *VM)
}

func (vm *VM) Run() {
	for vm.Err == nil && vm.PC < PC(len(vm.Tape)) {
		vm.Tape[vm.PC].Execute(vm)
		vm.PC++
	}
}

func StackPop[E object.Object](stack *stack.Stack[object.Object]) (E, bool) {
	var empty E
	if stack.Len() == 0 {
		return empty, false
	}

	v, ok := stack.Peek()
	if !ok {
		return empty, false
	}

	value, ok := v.(E)
	if !ok {
		return empty, false
	}

	stack.Pop()

	return value, true
}
