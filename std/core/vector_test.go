package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/ninedraft/sulisp/std/core"
)

const (
	nestedVectorDepth = 40
	deepVectorSize    = 100
	shrinkVectorSize  = 33
	deepAppendSize    = 1025
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

	t.Run("acts as append when index equals size", func(t *testing.T) {
		appended := base.Assoc(base.Size(), 888)
		assert.Equal(t, base.Size()+1, appended.Size())
		val, ok := appended.Get(base.Size())
		assert.True(t, ok)
		assert.Equal(t, 888, val)

		_, ok = base.Get(base.Size())
		assert.False(t, ok, "original vector should not expose the new value")
	})
}

func TestVectorAssocNilReceiver(t *testing.T) {
	t.Run("appends at zero index", func(t *testing.T) {
		var vector *Vector[int]
		appended := vector.Assoc(0, 99)
		require.NotNil(t, appended)
		assert.Equal(t, 1, appended.Size())
		val, ok := appended.Get(0)
		assert.True(t, ok)
		assert.Equal(t, 99, val)
	})

	t.Run("ignores indexes larger than size", func(t *testing.T) {
		var vector *Vector[int]
		assert.Nil(t, vector.Assoc(5, 1))
	})
}

func TestVectorAssocBounds(t *testing.T) {
	vector := buildVectorWithSize(t, 5)

	t.Run("negative index returns original vector", func(t *testing.T) {
		assert.Same(t, vector, vector.Assoc(-1, 1))
	})

	t.Run("indexes beyond size return original vector", func(t *testing.T) {
		assert.Same(t, vector, vector.Assoc(vector.Size()+1, 1))
	})
}

func TestVectorAssocDeep(t *testing.T) {
	vector := buildVectorWithSize(t, deepVectorSize)
	updated := vector.Assoc(deepVectorSize-1, 4242)
	val, ok := updated.Get(deepVectorSize - 1)
	assert.True(t, ok)
	assert.Equal(t, 4242, val)

	origVal, ok := vector.Get(deepVectorSize - 1)
	assert.True(t, ok)
	assert.Equal(t, deepVectorSize-1, origVal)
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

func TestVectorPopShrinksRoot(t *testing.T) {
	vector := buildVectorWithSize(t, shrinkVectorSize)
	popped := vector.Pop()
	assert.Equal(t, shrinkVectorSize-1, popped.Size())

	val, ok := popped.Get(shrinkVectorSize - 2)
	assert.True(t, ok)
	assert.Equal(t, shrinkVectorSize-2, val)
}

func TestVectorDeepAppendTriggersSecondLevel(t *testing.T) {
	vector := buildVectorWithSize(t, deepAppendSize)
	assert.Equal(t, deepAppendSize, vector.Size())

	val, ok := vector.Get(deepAppendSize - 1)
	require.True(t, ok)
	assert.Equal(t, deepAppendSize-1, val)
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

	t.Run("does nothing when vector is nil", func(t *testing.T) {
		var vector *Vector[int]
		called := false

		vector.All(func(int, int) bool {
			called = true
			return true
		})

		assert.False(t, called)
	})

	t.Run("pop on nil returns nil", func(t *testing.T) {
		var vector *Vector[int]
		assert.Nil(t, vector.Pop())
	})

	t.Run("size on nil returns zero", func(t *testing.T) {
		var vector *Vector[int]
		assert.Equal(t, 0, vector.Size())
	})

	t.Run("iterates nested nodes when depth exceeds one level", func(t *testing.T) {
		vector := buildVectorWithSize(t, nestedVectorDepth)

		var seen []int
		vector.All(func(int, int) bool {
			seen = append(seen, 0)
			return true
		})

		assert.Equal(t, vector.Size(), len(seen))
	})

	t.Run("stops when nested callback returns false", func(t *testing.T) {
		vector := buildVectorWithSize(t, nestedVectorDepth)

		counter := 0
		stopAt := 3
		vector.All(func(int, int) bool {
			counter++
			return counter < stopAt
		})

		assert.Equal(t, stopAt, counter)
	})
}

func buildVectorWithSize(t *testing.T, size int) *Vector[int] {
	t.Helper()
	var vector *Vector[int]
	for i := 0; i < size; i++ {
		vector = vector.Append(i)
	}

	require.Equal(t, size, vector.Size())
	return vector
}
