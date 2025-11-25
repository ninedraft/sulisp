package bytecode

import (
	"fmt"
	"hash/maphash"

	"github.com/ninedraft/sulisp/language/object"
)

type Continuation struct {
	ResumePC PC
}

func (cont *Continuation) Kind() object.Kind { return object.ObjContinuation }

func (cont *Continuation) Inspect() string {
	return fmt.Sprintf("<continuation resume=%d>", cont.ResumePC)
}

func (cont *Continuation) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, cont.Kind())
	maphash.WriteComparable(h, cont)
}

func (cont *Continuation) Compare(other object.Object) (int, bool) {
	o, ok := other.(*Continuation)
	if !ok {
		return 0, false
	}

	if cont.ResumePC == o.ResumePC {
		return 0, true
	}

	return 0, false
}

func PushContinuation() Command {
	return Command{
		Repr: "PushCont",
		Execute: func(vm *VM) {
			cont := &Continuation{
				ResumePC: vm.PC + 2,
			}
			vm.Stack.Push(cont)
		},
	}
}

func InvokeContinuation(argCount int) Command {
	return Command{
		Repr: fmt.Sprintf("InvokeCont(%d)", argCount),
		Execute: func(vm *VM) {
			contValue, ok := vm.Stack.Pop()
			if !ok {
				vm.Err = fmt.Errorf("%w: invoke continuation missing continuation", ErrBadStack)
				return
			}

			cont, ok := contValue.(*Continuation)
			if !ok {
				vm.Err = fmt.Errorf("%w: invoke continuation expected continuation", ErrBadStack)
				return
			}

			values := make([]object.Object, argCount)
			for i := argCount - 1; i >= 0; i-- {
				value, ok := vm.Stack.Pop()
				if !ok {
					vm.Err = fmt.Errorf("%w: invoke continuation args", ErrBadStack)
					return
				}
				values[i] = value
			}

			var result object.Object = object.Null{}
			if argCount > 0 {
				result = values[0]
			}

			vm.Stack.Push(result)
			if vm.CallStack.Len() == 0 {
				vm.Err = fmt.Errorf("%w: continuation without frame", ErrCallStackUnderflow)
				return
			}

			_, popped := vm.CallStack.Pop()
			if !popped {
				vm.Err = fmt.Errorf("%w: continuation pop", ErrCallStackUnderflow)
				return
			}

			vm.PC = cont.ResumePC - 1
		},
	}
}
