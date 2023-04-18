package core

import (
	"bytes"
	"strings"
)

type List[E Value] struct {
	head *listNode[E]
}

func ListNew[E Value](values ...E) *List[E] {
	head, _ := listAlloc(values)
	return &List[E]{
		head: head,
	}
}

func (*List[E]) Type() Value {
	var elem E
	return newTypeSpec("core.List", map[Keyword]Value{
		":elem": elem.Type(),
	})
}

func (list *List[E]) Seq() Seq[E] {
	return list
}

func (list List[E]) String() string {
	if list.head == nil {
		return "(list)"
	}

	dst := &strings.Builder{}
	_, _ = dst.WriteString("(list")

	for node := list.head; node != nil; node = node.next {
		_, _ = dst.WriteString(" ")
		_, _ = dst.WriteString(node.value.String())
	}

	dst.WriteString(")")

	return dst.String()
}

func (list List[E]) MarshalText() ([]byte, error) {
	if list.head == nil {
		return []byte("(list)"), nil
	}

	buf := &bytes.Buffer{}
	buf.WriteString("(list")

	for node := list.head; node != nil; node = node.next {
		buf.WriteByte(' ')

		bb, err := node.value.MarshalText()
		if err != nil {
			return nil, err
		}
		buf.Write(bb)
	}

	buf.WriteByte(')')

	return buf.Bytes(), nil
}

func (list *List[E]) Contains(value E) bool {
	if list == nil || list.head == nil {
		return false
	}

	for node := list.head; node != nil; node = node.next {
		if Eq(node.value, value) {
			return true
		}
	}

	return false
}

type listNode[E Value] struct {
	next *listNode[E]

	value E
}

func (l *List[E]) Len() int {
	if l == nil || l.head == nil {
		return 0
	}

	n := 0
	for node := l.head; node != nil; node = node.next {
		n++
	}
	return n
}

func (l *List[E]) First() (E, bool) {
	if l.Empty() {
		var empty E
		return empty, false
	}

	return l.head.value, true
}

func (l *List[E]) Next() Seq[E] {
	if l.Empty() {
		return nil
	}

	return &List[E]{head: l.head.next}
}

func (l *List[E]) Empty() bool {
	return l == nil || l.head == nil
}

func (l *List[E]) Eq(other *List[E]) bool {
	switch {
	case l.Empty():
		return other.Empty()
	case other.Empty():
		return false
	}

	a, b := l.head, other.head
	for a != nil && b != nil {
		if !Eq(a.value, b.value) {
			return false
		}

		a, b = a.next, b.next
	}

	return true
}

func (l *List[E]) Conj(values ...E) *List[E] {
	if len(values) == 0 {
		return l
	}

	head, tail := listAlloc(values)
	tail.next = l.head

	return &List[E]{head: head}
}

func listAlloc[E Value](values []E) (head, tail *listNode[E]) {
	n := len(values)
	if n == 0 {
		return nil, nil
	}

	nodes := make([]listNode[E], n)
	nodes[0].value = values[0]

	for i, v := range values[1:] {
		nodes[i].next = &nodes[i+1]
		nodes[i+1].value = v
	}

	return &nodes[0], &nodes[n-1]
}
