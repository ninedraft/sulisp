package bytecode

import (
	"github.com/ninedraft/sulisp/internal/collections/stack"
	"github.com/ninedraft/sulisp/language/object"
)

type PC int

type Frame struct {
	ReturnPC PC
	Locals   []object.Object
}

type VM struct {
	Stack     stack.Stack[object.Object]
	Tape      []Command
	PC        PC
	Err       error
	CallStack stack.Stack[*Frame]
}

type Command struct {
	Repr    string
	Execute func(vm *VM)
}

func NewVM(tape []Command) *VM {
	vm := &VM{
		Tape: tape,
	}
	vm.CallStack.Push(&Frame{})
	return vm
}

func (vm *VM) Run() {
	for vm.Err == nil && vm.PC < PC(len(vm.Tape)) {
		vm.Tape[vm.PC].Execute(vm)
		vm.PC++
	}
}

func (vm *VM) currentFrame() (*Frame, bool) {
	frame, ok := vm.CallStack.Peek()
	return frame, ok
}

func (vm *VM) jump(target PC) {
	vm.PC = target - 1
}

func (frame *Frame) ensureSlot(slot int) {
	if slot < 0 {
		return
	}
	if slot < len(frame.Locals) {
		return
	}

	newLocals := make([]object.Object, slot+1)
	copy(newLocals, frame.Locals)
	frame.Locals = newLocals
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
