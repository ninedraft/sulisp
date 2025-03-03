package bytecode

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/ninedraft/sulisp/language/object"
)

var ErrBadStack = errors.New("bad values on stack, or stack is empty")

func Const[E object.PrimitiveTypes](value E) Command {
	return Command{
		Repr: "Const",
		Execute: func(vm *VM) {
			vm.Stack.Push(object.PrimitiveOf(value))
		},
	}
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
