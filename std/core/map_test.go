package core_test

import (
	"hash/maphash"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	core "github.com/ninedraft/sulisp/std/core"
)

func TestPersistentHashPutAndGet(t *testing.T) {
	hash := core.NewMap[int, string](hashInt, equal)

	updated := hash.Put(1, "one")
	assert.Equal(t, 1, updated.Size())

	value, ok := updated.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "one", value)

	assert.Equal(t, 0, hash.Size())
	_, ok = hash.Get(1)
	assert.False(t, ok)

	replaced := updated.Put(1, "uno")
	assert.Equal(t, 1, replaced.Size())
	value, ok = replaced.Get(1)
	require.True(t, ok)
	assert.Equal(t, "uno", value)
}

func TestPersistentHashRemove(t *testing.T) {
	hash := buildHash(t, 3)
	removed := hash.Remove(1)

	assert.Equal(t, hash.Size()-1, removed.Size())
	_, ok := removed.Get(1)
	assert.False(t, ok)

	value, ok := hash.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 10, value)

	same := hash.Remove(999)
	assert.Equal(t, hash.Size(), same.Size())
}

func TestPersistentHashIteration(t *testing.T) {
	hash := buildHash(t, 5)

	seen := make(map[int]int)
	count := 0
	for k, v := range hash.All {
		seen[k] = v
		count++
	}

	assert.Equal(t, 5, count)
	for i := 0; i < 5; i++ {
		value, ok := seen[i]
		assert.True(t, ok)
		assert.Equal(t, i*10, value)
	}

	var empty *core.Map[int, int]
	for range empty.All {
		t.Fatal("should not iterate nil map")
	}
}

func TestPersistentHashSizeNil(t *testing.T) {
	var hash *core.Map[int, int]
	assert.Equal(t, 0, hash.Size())
}

func hashInt(key int, h *maphash.Hash) {
	h.WriteString(strconv.Itoa(key))
}

func equal[K comparable](a, b K) bool {
	return a == b
}

func buildHash(t *testing.T, size int) *core.Map[int, int] {
	t.Helper()

	hash := core.NewMap[int, int](hashInt, equal)
	for i := 0; i < size; i++ {
		hash = hash.Put(i, i*10)
	}

	require.Equal(t, size, hash.Size())
	return hash
}
