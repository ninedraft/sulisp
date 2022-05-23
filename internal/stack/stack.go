package stack

import "golang.org/x/exp/slices"

type Stack[T any] struct {
	values []T
}

func (stack *Stack[T]) Len() int {
	if stack == nil {
		return 0
	}
	return len(stack.values)
}

func (stack *Stack[T]) Push(values ...T) {
	stack.values = append(stack.values, values...)
}

func (stack *Stack[T]) Pop() (T, bool) {
	var empty T
	if stack == nil || len(stack.values) == 0 {
		return empty, false
	}
	n := len(stack.values) - 1
	value := stack.values[n]
	stack.values[n] = empty
	stack.values = stack.values[:n]
	return value, true
}

func (stack *Stack[T]) Peek() (T, bool) {
	if stack == nil || len(stack.values) == 0 {
		var empty T
		return empty, false
	}
	n := len(stack.values) - 1
	return stack.values[n], true
}

func (stack *Stack[T]) Clone() *Stack[T] {
	return &Stack[T]{
		values: slices.Clone(stack.values),
	}
}
