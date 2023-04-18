package core

import (
	"strconv"
	"sync/atomic"
)

type Chan[E Value] struct {
	closed atomic.Bool
	value  atomic.Value
	c      chan E
}

func NewChan[E Value](size int) *Chan[E] {
	return &Chan[E]{
		c: make(chan E, size),
	}
}

func NewChanOf[E Value](values ...E) *Chan[E] {
	c := make(chan E, len(values))
	for _, value := range values {
		c <- value
	}

	return &Chan[E]{c: c}
}

func (ch *Chan[E]) String() string {
	return "(chan" + strconv.Itoa(cap(ch.c)) + ")"
}

func (ch *Chan[E]) GoString() string {
	return "NewChan(" + strconv.Itoa(cap(ch.c)) + ")"
}

func (ch *Chan[E]) Seq() Seq[E] {
	return ch
}

func (ch *Chan[E]) Push(value E) {
	ch.c <- value
}

func (ch *Chan[E]) First() (E, bool) {
	if ch.Empty() {
		var empty E
		return empty, false
	}

	value, ok := ch.value.Load().(E)
	return value, ok
}

func (ch *Chan[E]) Next() Seq[E] {
	if ch.Empty() {
		return nil
	}
	value, ok := <-ch.c
	if !ok {
		ch.closed.Store(true)
		return nil
	}
	ch.value.Store(value)
	return ch
}

func (ch *Chan[E]) Close() {
	ch.closed.Store(true)
	close(ch.c)
}

func (ch *Chan[E]) Len() int {
	return len(ch.c)
}

func (ch *Chan[E]) Empty() bool {
	return ch.closed.Load()
}
