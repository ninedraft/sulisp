package bytecode

import (
	"fmt"

	"github.com/ninedraft/sulisp/internal/collections/stack"
	"github.com/ninedraft/sulisp/language/object"
)

type HostEffectFn func(ctx *CallCtx, args []object.Object)

type HostEffectRegistry struct {
	effects map[string]HostEffectFn
}

func NewHostEffectRegistry() *HostEffectRegistry {
	return &HostEffectRegistry{
		effects: map[string]HostEffectFn{},
	}
}

func (registry *HostEffectRegistry) Register(effect, operation string, handler HostEffectFn) error {
	if effect == "" || operation == "" {
		return fmt.Errorf("effect and operation required")
	}
	if handler == nil {
		return fmt.Errorf("host effect handler required")
	}
	if registry.effects == nil {
		registry.effects = map[string]HostEffectFn{}
	}

	key := hostEffectKey(effect, operation)
	if _, ok := registry.effects[key]; ok {
		return fmt.Errorf("host effect %s.%s already registered", effect, operation)
	}

	registry.effects[key] = handler
	return nil
}

func (registry *HostEffectRegistry) Lookup(effect, operation string) (HostEffectFn, bool) {
	if registry == nil || registry.effects == nil {
		return nil, false
	}

	handler, ok := registry.effects[hostEffectKey(effect, operation)]
	return handler, ok
}

func hostEffectKey(effect, operation string) string {
	return effect + ":" + operation
}

type CallCtx struct {
	Continuation *Continuation
	Stack        *stack.Stack[object.Object]
	VM           *VM
	Effect       string
	Operation    string

	invoked bool
}

func (ctx *CallCtx) Invoke(args ...object.Object) {
	for _, arg := range args {
		ctx.Stack.Push(arg)
	}
	ctx.Stack.Push(ctx.Continuation)
	ctx.invoked = true
	InvokeContinuation(len(args)).Execute(ctx.VM)
}

func CallHost(effect, operation string, argCount int) Command {
	return Command{
		Repr: fmt.Sprintf("CallHost(%s.%s, args=%d)", effect, operation, argCount),
		Execute: func(vm *VM) {
			handler, ok := vm.lookupHostEffect(effect, operation)
			if !ok || handler == nil {
				vm.Err = fmt.Errorf("host effect %s.%s not registered", effect, operation)
				return
			}

			contValue, ok := vm.Stack.Pop()
			if !ok {
				vm.Err = fmt.Errorf("%w: host effect %s.%s missing continuation", ErrBadStack, effect, operation)
				return
			}

			cont, ok := contValue.(*Continuation)
			if !ok {
				vm.Err = fmt.Errorf("%w: host effect %s.%s expected continuation", ErrBadStack, effect, operation)
				return
			}

			args := make([]object.Object, argCount)
			for i := argCount - 1; i >= 0; i-- {
				value, ok := vm.Stack.Pop()
				if !ok {
					vm.Err = fmt.Errorf("%w: host effect %s.%s args", ErrBadStack, effect, operation)
					return
				}
				args[i] = value
			}

			frame := &Frame{
				ReturnPC: vm.PC,
			}
			vm.CallStack.Push(frame)

			ctx := &CallCtx{
				Continuation: cont,
				Stack:        &vm.Stack,
				VM:           vm,
				Effect:       effect,
				Operation:    operation,
			}

			handler(ctx, args)

			if !ctx.invoked {
				if _, popped := vm.CallStack.Pop(); !popped {
					if vm.Err == nil {
						vm.Err = fmt.Errorf("%w: host effect %s.%s call stack", ErrCallStackUnderflow, effect, operation)
					}
					return
				}
				if vm.Err == nil {
					vm.Err = fmt.Errorf("host effect %s.%s did not resume continuation", effect, operation)
				}
			}
		},
	}
}
