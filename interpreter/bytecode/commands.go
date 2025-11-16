package bytecode

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/ninedraft/sulisp/language/object"
)

var (
	ErrBadStack           = errors.New("bad values on stack, or stack is empty")
	ErrNoFrame            = errors.New("no frame available for locals")
	ErrCallStackUnderflow = errors.New("call stack underflow")
)

func Const[E object.PrimitiveTypes](value E) Command {
	return Command{
		Repr: fmt.Sprintf("Const(%v)", value),
		Execute: func(vm *VM) {
			vm.Stack.Push(object.PrimitiveOf(value))
		},
	}
}

var Null = Command{
	Repr: "Null",
	Execute: func(vm *VM) {
		vm.Stack.Push(&object.Null{})
	},
}

var Add = Command{
	Repr: "Add",
	Execute: func(vm *VM) {
		left, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: add: left operand", ErrBadStack)
			return
		}

		right, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: add: right operand", ErrBadStack)
			return
		}

		vm.Stack.Push(object.PrimitiveOf(left.Value + right.Value))
	},
}

var Sub = Command{
	Repr: "Sub",
	Execute: func(vm *VM) {
		right, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: sub: right operand", ErrBadStack)
			return
		}
		left, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: sub: left operand", ErrBadStack)
			return
		}
		vm.Stack.Push(object.PrimitiveOf(left.Value - right.Value))
	},
}

var Mul = Command{
	Repr: "Mul",
	Execute: func(vm *VM) {
		right, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: mul: right operand", ErrBadStack)
			return
		}
		left, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: mul: left operand", ErrBadStack)
			return
		}
		vm.Stack.Push(object.PrimitiveOf(left.Value * right.Value))
	},
}

var Div = Command{
	Repr: "Div",
	Execute: func(vm *VM) {
		right, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: div: right operand", ErrBadStack)
			return
		}
		if right.Value == 0 {
			vm.Err = errors.New("division by zero")
			return
		}
		left, ok := StackPop[*object.Primitive[int64]](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: div: left operand", ErrBadStack)
			return
		}
		vm.Stack.Push(object.PrimitiveOf(left.Value / right.Value))
	},
}

var Equal = Command{
	Repr: "Equal",
	Execute: func(vm *VM) {
		right, ok := StackPop[object.Object](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: equal: right operand", ErrBadStack)
			return
		}
		left, ok := StackPop[object.Object](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: equal: left operand", ErrBadStack)
			return
		}
		if ord, ok := left.(object.Ordered); ok {
			if cmp, ok := ord.Compare(right); ok {
				vm.Stack.Push(object.PrimitiveOf(cmp == 0))
				return
			}
		}
		vm.Stack.Push(object.PrimitiveOf(false))
	},
}

var ArrayAppend = Command{
	Repr: "ArrayAppend",
	Execute: func(vm *VM) {
		value, ok := StackPop[object.Object](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: array append: value", ErrBadStack)
			return
		}

		array, ok := StackPop[*object.Array](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: array append: array", ErrBadStack)
			return
		}

		array.Elements = append(array.Elements, value)
		vm.Stack.Push(array)
	},
}

var ArrayNew = Command{
	Repr: "ArrayNew",
	Execute: func(vm *VM) {
		vm.Stack.Push(&object.Array{Elements: make([]object.Object, 0)})
	},
}

var ArrayLen = Command{
	Repr: "ArrayLen",
	Execute: func(vm *VM) {
		arr, ok := StackPop[*object.Array](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: array length", ErrBadStack)
			return
		}
		vm.Stack.Push(object.PrimitiveOf(int64(len(arr.Elements))))
	},
}

// Bonus: Array shuffle command using Fisher-Yates algorithm
var ArrayShuffle = Command{
	Repr: "ArrayShuffle",
	Execute: func(vm *VM) {
		arr, ok := StackPop[*object.Array](&vm.Stack)
		if !ok {
			vm.Err = fmt.Errorf("%w: array shuffle", ErrBadStack)
			return
		}
		rand.Shuffle(len(arr.Elements), func(i, j int) {
			arr.Elements[i], arr.Elements[j] = arr.Elements[j], arr.Elements[i]
		})
		vm.Stack.Push(arr)
	},
}

// Bonus: Command to create a range of numbers [start, end)
func Range(start, end int64) Command {
	return Command{
		Repr: fmt.Sprintf("Range(%d, %d)", start, end),
		Execute: func(vm *VM) {
			elements := make([]object.Object, 0, end-start)
			for i := start; i < end; i++ {
				elements = append(elements, object.PrimitiveOf(i))
			}
			vm.Stack.Push(&object.Array{Elements: elements})
		},
	}
}

func Jump(target PC) Command {
	return Command{
		Repr: fmt.Sprintf("Jump(%d)", target),
		Execute: func(vm *VM) {
			vm.jump(target)
		},
	}
}

func JumpIfTrue(target PC) Command {
	return Command{
		Repr: fmt.Sprintf("JumpIfTrue(%d)", target),
		Execute: func(vm *VM) {
			value, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: jump if true", ErrBadStack)
				return
			}
			if value.Value {
				vm.jump(target)
			}
		},
	}
}

func JumpIfFalse(target PC) Command {
	return Command{
		Repr: fmt.Sprintf("JumpIfFalse(%d)", target),
		Execute: func(vm *VM) {
			value, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: jump if false", ErrBadStack)
				return
			}
			if !value.Value {
				vm.jump(target)
			}
		},
	}
}

func Call(target PC, locals int) Command {
	return Command{
		Repr: fmt.Sprintf("Call(%d, locals=%d)", target, locals),
		Execute: func(vm *VM) {
			if locals < 0 {
				vm.Err = fmt.Errorf("call: invalid locals %d", locals)
				return
			}

			frame := &Frame{
				ReturnPC: vm.PC,
				Locals:   make([]object.Object, locals),
			}

			for i := locals - 1; i >= 0; i-- {
				value, ok := vm.Stack.Pop()
				if !ok {
					vm.Err = fmt.Errorf("%w: call args", ErrBadStack)
					return
				}
				frame.Locals[i] = value
			}

			vm.CallStack.Push(frame)
			vm.jump(target)
		},
	}
}

var Return = Command{
	Repr: "Return",
	Execute: func(vm *VM) {
		if vm.CallStack.Len() <= 1 {
			vm.Err = fmt.Errorf("%w: return without call", ErrCallStackUnderflow)
			return
		}
		frame, ok := vm.CallStack.Pop()
		if !ok {
			vm.Err = fmt.Errorf("%w: return pop", ErrCallStackUnderflow)
			return
		}

		vm.PC = frame.ReturnPC
	},
}

func StoreLocal(slot int) Command {
	return Command{
		Repr: fmt.Sprintf("StoreLocal(%d)", slot),
		Execute: func(vm *VM) {
			frame, ok := vm.currentFrame()
			if !ok {
				vm.Err = ErrNoFrame
				return
			}

			value, ok := vm.Stack.Pop()
			if !ok {
				vm.Err = fmt.Errorf("%w: store local", ErrBadStack)
				return
			}

			frame.ensureSlot(slot)
			frame.Locals[slot] = value
		},
	}
}

func LoadLocal(slot int) Command {
	return Command{
		Repr: fmt.Sprintf("LoadLocal(%d)", slot),
		Execute: func(vm *VM) {
			frame, ok := vm.currentFrame()
			if !ok {
				vm.Err = ErrNoFrame
				return
			}

			if slot < 0 || slot >= len(frame.Locals) {
				vm.Err = fmt.Errorf("load local: slot %d empty", slot)
				return
			}
			value := frame.Locals[slot]
			vm.Stack.Push(value)
		},
	}
}

func Halt() Command {
	return Command{
		Repr: "Halt",
		Execute: func(vm *VM) {
			vm.PC = PC(len(vm.Tape))
		},
	}
}

func Pop() Command {
	return Command{
		Repr: "Pop",
		Execute: func(vm *VM) {
			if _, ok := vm.Stack.Pop(); !ok {
				vm.Err = fmt.Errorf("%w: pop", ErrBadStack)
			}
		},
	}
}

func Dup() Command {
	return Command{
		Repr: "Dup",
		Execute: func(vm *VM) {
			value, ok := vm.Stack.Peek()
			if !ok {
				vm.Err = fmt.Errorf("%w: dup", ErrBadStack)
				return
			}
			vm.Stack.Push(value)
		},
	}
}

func Swap() Command {
	return Command{
		Repr: "Swap",
		Execute: func(vm *VM) {
			first, ok := vm.Stack.Pop()
			if !ok {
				vm.Err = fmt.Errorf("%w: swap first", ErrBadStack)
				return
			}
			second, ok := vm.Stack.Pop()
			if !ok {
				vm.Err = fmt.Errorf("%w: swap second", ErrBadStack)
				vm.Stack.Push(first)
				return
			}
			vm.Stack.Push(first)
			vm.Stack.Push(second)
		},
	}
}

func Not() Command {
	return Command{
		Repr: "Not",
		Execute: func(vm *VM) {
			value, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: not", ErrBadStack)
				return
			}
			vm.Stack.Push(object.PrimitiveOf(!value.Value))
		},
	}
}

func And() Command {
	return Command{
		Repr: "And",
		Execute: func(vm *VM) {
			right, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: and right", ErrBadStack)
				return
			}
			left, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: and left", ErrBadStack)
				return
			}
			vm.Stack.Push(object.PrimitiveOf(left.Value && right.Value))
		},
	}
}

func Or() Command {
	return Command{
		Repr: "Or",
		Execute: func(vm *VM) {
			right, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: or right", ErrBadStack)
				return
			}
			left, ok := StackPop[*object.Primitive[bool]](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: or left", ErrBadStack)
				return
			}
			vm.Stack.Push(object.PrimitiveOf(left.Value || right.Value))
		},
	}
}

func Less() Command {
	return comparisonCommand("Less", func(cmp int) bool { return cmp < 0 })
}

func LessEqual() Command {
	return comparisonCommand("LessEqual", func(cmp int) bool { return cmp <= 0 })
}

func Greater() Command {
	return comparisonCommand("Greater", func(cmp int) bool { return cmp > 0 })
}

func GreaterEqual() Command {
	return comparisonCommand("GreaterEqual", func(cmp int) bool { return cmp >= 0 })
}

func comparisonCommand(name string, fn func(int) bool) Command {
	return Command{
		Repr: name,
		Execute: func(vm *VM) {
			right, ok := StackPop[object.Object](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: %s right operand", ErrBadStack, name)
				return
			}
			left, ok := StackPop[object.Object](&vm.Stack)
			if !ok {
				vm.Err = fmt.Errorf("%w: %s left operand", ErrBadStack, name)
				return
			}

			cmp, ok := compareObjects(left, right)
			if !ok {
				vm.Err = fmt.Errorf("%w: %s unsupported operands", ErrBadStack, name)
				return
			}

			vm.Stack.Push(object.PrimitiveOf(fn(cmp)))
		},
	}
}

func compareObjects(left, right object.Object) (int, bool) {
	if ord, ok := left.(object.Ordered); ok {
		if cmp, ok := ord.Compare(right); ok {
			return cmp, true
		}
	}
	if ord, ok := right.(object.Ordered); ok {
		if cmp, ok := ord.Compare(left); ok {
			return -cmp, true
		}
	}
	return 0, false
}
