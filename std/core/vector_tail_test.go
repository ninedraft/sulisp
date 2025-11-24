package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorTailHoldsRecentValues(t *testing.T) {
	var vector *Vector[int]
	for i := 0; i < vectorB-1; i++ {
		vector = vector.Append(i)
	}

	require.NotNil(t, vector)
	require.Equal(t, vectorB-1, vector.Size())
	assert.Nil(t, vector.root, "all elements should stay in tail while it fits")
	assert.Len(t, vector.tail, vectorB-1)

	for i := 0; i < vector.Size(); i++ {
		value, ok := vector.Get(i)
		require.True(t, ok)
		assert.Equal(t, i, value)
	}
}

func TestVectorFlushesTailWhenOverflowing(t *testing.T) {
	var vector *Vector[int]
	for i := 0; i < vectorB; i++ {
		vector = vector.Append(i)
	}

	require.Len(t, vector.tail, vectorB)
	require.Nil(t, vector.root, "tail should hold the first block")

	vector = vector.Append(999)

	require.Equal(t, vectorB+1, vector.Size())
	require.Len(t, vector.tail, 1, "new tail starts with the overflow value")
	require.NotNil(t, vector.root, "full tail should be stored in the tree")

	for i := 0; i < vectorB; i++ {
		value, ok := vector.Get(i)
		require.True(t, ok)
		assert.Equal(t, i, value)
	}

	value, ok := vector.Get(vectorB)
	require.True(t, ok)
	assert.Equal(t, 999, value)
}

func TestVectorPopRestoresTailFromTree(t *testing.T) {
	var vector *Vector[int]
	for i := 0; i < vectorB+1; i++ {
		vector = vector.Append(i)
	}

	require.Len(t, vector.tail, 1, "last element should live in tail")
	require.NotNil(t, vector.root, "full tail gets flushed into the root")

	vector = vector.Pop()

	require.Equal(t, vectorB, vector.Size())
	assert.Len(t, vector.tail, vectorB)
	assert.Nil(t, vector.root, "tree is empty when only tail remains")

	for i := 0; i < vectorB; i++ {
		value, ok := vector.Get(i)
		require.True(t, ok)
		assert.Equal(t, i, value)
	}
}
