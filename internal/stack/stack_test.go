package stack_test

import (
	"testing"

	"github.com/ninedraft/sulisp/internal/stack"
)

func TestStackPushPop(test *testing.T) {
	s := &stack.Stack[int]{}
	s.Push(100)

	value, okValue := s.Pop()
	if !okValue {
		test.Errorf("s.Pop() expected to return true")
	}
	if value != 100 {
		test.Errorf("s.Pop() expected to return 100")
	}

	_, okEmpty := s.Pop()
	if okEmpty {
		test.Errorf("s.Pop() expected to return false")
	}
}

func TestStackClone(test *testing.T) {
	s := &stack.Stack[int]{}
	s.Push(1, 2, 3, 4, 5, 6)

	cp := s.Clone()

	for s.Len() > 0 {
		expected, _ := s.Pop()
		got, gotOk := cp.Pop()
		if !gotOk {
			test.Fatalf("got unexpected copy.Pop()==false")
		}
		if got != expected {
			test.Fatalf("%d is expected, got %d", expected, got)
		}
	}
}

func TestStackPeek(test *testing.T) {
	s := &stack.Stack[int]{}
	for i := 0; i < 10; i++ {
		s.Push(i)
		got, ok := s.Peek()
		if !ok {
			test.Fatalf("got unexpected copy.Pop()==false")
		}
		if got != i {
			test.Fatalf("%d is expected, got %d", i, got)
		}
	}
}
