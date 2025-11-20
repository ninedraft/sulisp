package core

import (
	"fmt"
	"math"
	"slices"
)

const (
	vectorB       = 32
	vectorBits    = 5
	vectorBitMask = vectorB - 1
)

type Vector[E any] struct {
	size  int
	depth int
	root  *vectorNode[E]
}

type vectorNode[E any] struct {
	children []any
}

func (node *vectorNode[E]) Clone() *vectorNode[E] {
	if node == nil {
		return nil
	}

	return &vectorNode[E]{
		children: slices.Clone(node.children),
	}
}

func indexAtLevel(index int, level int) int {
	return (index >> (level * vectorBits)) & vectorBitMask
}

func (vector *Vector[E]) Get(index int) (E, bool) {
	if index < 0 || index >= vector.size {
		var zero E
		return zero, false
	}

	node := vector.root

	for level := vector.depth; level > 0; level-- {
		idx := indexAtLevel(index, level)
		child, ok := node.children[idx].(*vectorNode[E])
		if !ok {
			var zero E
			return zero, false
		}
		node = child
	}

	idx := indexAtLevel(index, 0)
	value, ok := node.children[idx].(E)
	return value, ok
}

func (vector *Vector[E]) Assoc(index int, value E) *Vector[E] {
	if index < 0 || index >= vector.size {
		return vector
	}

	root := vector.assocNode(vector.root, vector.depth, index, value)
	return &Vector[E]{
		size:  vector.size,
		depth: vector.depth,
		root:  root,
	}
}

func (vector *Vector[E]) assocNode(node *vectorNode[E], level int, index int, value E) *vectorNode[E] {
	root := node.Clone()

	i := indexAtLevel(index, level)
	if level == 0 {
		root.children[i] = value
		return root
	}

	child, ok := root.children[i].(*vectorNode[E])
	if !ok {
		panic(fmt.Sprintf("vector is broken, got unexpected child node %T at level=%d index=%d", root.children[i], level, i))
	}
	root.children[i] = vector.assocNode(child, level-1, index, value)

	return root
}

func (vector *Vector[E]) Append(value E) *Vector[E] {
	if vector == nil {
		newRoot := vector.newPath(0, 0, value).(*vectorNode[E])
		return &Vector[E]{
			size:  1,
			depth: 0,
			root:  newRoot,
		}
	}

	vecCap := int(math.Pow(vectorB, float64(vector.depth+1)))
	if vector.size < vecCap {
		root := vector.pushNode(vector.root, vector.depth, vector.size, value)

		return &Vector[E]{
			size:  vector.size + 1,
			depth: vector.depth,
			root:  root,
		}
	}

	root := &vectorNode[E]{
		children: make([]any, vectorB),
	}
	root.children[0] = vector.root
	root.children[1] = vector.newPath(vector.depth, vector.size, value)

	return &Vector[E]{
		size:  vector.size + 1,
		depth: vector.depth + 1,
		root:  root,
	}
}

func (vector *Vector[E]) pushNode(node *vectorNode[E], level int, index int, value E) *vectorNode[E] {
	newNode := node.Clone()

	i := indexAtLevel(index, level)

	if level == 0 {
		newNode.children[i] = value
		return newNode
	}

	child, ok := newNode.children[i].(*vectorNode[E])
	if !ok {
		newNode.children[i] = vector.newPath(level-1, index, value)
		return newNode
	}

	newNode.children[i] = vector.pushNode(child, level-1, index, value)

	return newNode
}

func (vector *Vector[E]) newPath(depth int, i int, value E) any {
	node := &vectorNode[E]{
		children: make([]any, vectorB),
	}

	if depth == 0 {
		idx := indexAtLevel(i, 0)
		node.children[idx] = value
		return node // as leaf
	}

	idx := indexAtLevel(i, depth)
	node.children[idx] = vector.newPath(depth-1, i, value)

	return node
}

func (vector *Vector[E]) Pop() *Vector[E] {
	if vector == nil {
		return nil
	}

	if vector.size <= 1 {
		return &Vector[E]{
			root: &vectorNode[E]{},
		}
	}

	newRoot := vector.popNode(vector.root, vector.depth, vector.size-1)

	newDepth := vector.depth
	if vector.depth > 0 && vector.hasOnlyOneChild(newRoot, 0) {
		r, ok := newRoot.children[0].(*vectorNode[E])
		if !ok {
			panic(fmt.Sprintf("vector is brokent, unexpected child node %T at index=0", newRoot.children[0]))
		}
		newRoot = r
		newDepth--
	}

	return &Vector[E]{
		size:  vector.size - 1,
		depth: newDepth,
		root:  newRoot,
	}
}

func (vector *Vector[E]) popNode(node *vectorNode[E], level, index int) *vectorNode[E] {
	newNode := node.Clone()
	idx := indexAtLevel(index, level)

	if level == 0 {
		newNode.children[idx] = nil
		return newNode
	}

	child, ok := newNode.children[idx].(*vectorNode[E])
	if ok {
		newNode.children[idx] = vector.popNode(child, level-1, index)
	}

	return newNode
}

func (vector *Vector[E]) hasOnlyOneChild(node *vectorNode[E], idx int) bool {
	if node == nil {
		return false
	}

	for k := range vectorB - 1 {
		if k == idx {
			continue
		}

		if node.children[k] != nil {
			return false
		}
	}

	return node.children[idx] != nil
}

func (vector *Vector[E]) All(yield func(int, E) bool) {
	if vector == nil {
		return
	}

	var visitNode func(node *vectorNode[E], level int, indexBase int) bool

	visitNode = func(node *vectorNode[E], level int, indexBase int) bool {
		if level == 0 {
			for i, child := range node.children {
				if child == nil {
					continue
				}

				if !yield(indexBase+i, child.(E)) {
					return false
				}
			}

			return true
		}

		for i, child := range node.children {
			if child == nil {
				continue
			}

			if !visitNode(child.(*vectorNode[E]), level-1, indexBase+i*vectorB) {
				return false
			}
		}

		return true
	}

	visitNode(vector.root, vector.depth, 0)
}

func (vector *Vector[E]) Size() int {
	if vector == nil {
		return 0
	}
	return vector.size
}
