package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/ninedraft/sulisp/std/core"
)

func TestVectorAppendAndGet(t *testing.T) {
	t.Run("grows beyond root and keeps inserted elements", func(t *testing.T) {
		var vector *Vector[int]
		for i := 0; i < 70; i++ {
			vector = vector.Append(i)
			assert.Equal(t, i+1, vector.Size())
			val, ok := vector.Get(i)
			assert.True(t, ok)
			assert.Equal(t, i, val)
		}
	})

	t.Run("returns false for out of range indexes", func(t *testing.T) {
		var vector *Vector[int]
		vector = vector.Append(42)

		for _, idx := range []int{-1, 1, 5} {
			_, ok := vector.Get(idx)
			assert.False(t, ok, "index %d should be invalid", idx)
		}
	})
}

func TestVectorAssoc(t *testing.T) {
	var base *Vector[int]
	for i := 0; i < 3; i++ {
		base = base.Append(i * 10)
	}

	t.Run("updates a single index without mutating the original vector", func(t *testing.T) {
		updated := base.Assoc(1, 99)
		assert.Equal(t, 3, updated.Size())
		assert.Equal(t, 3, base.Size(), "original vector should stay intact")

		val, ok := updated.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 99, val)

		origVal, ok := base.Get(1)
		assert.True(t, ok)
		assert.Equal(t, 10, origVal)
	})
}

func TestVectorPop(t *testing.T) {
	var vector *Vector[int]
	for i := 0; i < 10; i++ {
		vector = vector.Append(i)
	}

	t.Run("drops the last element and keeps previous values", func(t *testing.T) {
		popped := vector.Pop()
		assert.Equal(t, 9, popped.Size())

		for i := 0; i < popped.Size(); i++ {
			val, ok := popped.Get(i)
			assert.True(t, ok)
			assert.Equal(t, i, val)
		}

		_, ok := popped.Get(9)
		assert.False(t, ok)
	})

	t.Run("returns nil-safe vector when popping empty or single element instance", func(t *testing.T) {
		var single *Vector[int]
		single = single.Append(1)
		single = single.Pop()
		assert.Equal(t, 0, single.Size())
		val, ok := single.Get(0)
		assert.False(t, ok)
		assert.Equal(t, 0, val)
	})
}

func TestVectorAll(t *testing.T) {
	t.Run("iterates in order and exposes indexes", func(t *testing.T) {
		var vector *Vector[int]
		const total = 12
		for i := 0; i < total; i++ {
			vector = vector.Append(i * 2)
		}

		var seen []int
		vector.All(func(idx int, value int) bool {
			seen = append(seen, value)
			assert.Equal(t, len(seen)-1, idx)
			return true
		})

		assert.Equal(t, total, len(seen))
		for i := 0; i < total; i++ {
			assert.Equal(t, i*2, seen[i])
		}
	})

	t.Run("stops when callback returns false", func(t *testing.T) {
		var vector *Vector[int]
		for i := 0; i < 5; i++ {
			vector = vector.Append(i)
		}

		counter := 0
		vector.All(func(int, int) bool {
			counter++
			return counter < 3
		})

		assert.Equal(t, 3, counter)
	})
}
