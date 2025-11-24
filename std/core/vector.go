package core

import (
	"fmt"
	"iter"
	"slices"
	"strings"
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
	tail  []E
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

func vectorCapacity(depth int) int {
	capacity := 1
	for i := 0; i <= depth; i++ {
		capacity *= vectorB
	}
	return capacity
}

func (vector *Vector[E]) tailOffset() int {
	if vector == nil {
		return 0
	}

	if vector.size < vectorB {
		return 0
	}

	return vector.size - len(vector.tail)
}

func (vector *Vector[E]) Get(index int) (E, bool) {
	if vector == nil || index < 0 || index >= vector.size {
		var zero E
		return zero, false
	}

	tailOffset := vector.tailOffset()
	if index >= tailOffset {
		return vector.tail[index-tailOffset], true
	}

	node := vector.root
	if node == nil {
		var zero E
		return zero, false
	}

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
	if index < 0 {
		return vector
	}

	if vector == nil {
		if index == 0 {
			return vector.Append(value)
		}
		return vector
	}

	if index == vector.size {
		return vector.Append(value)
	}

	if index > vector.size {
		return vector
	}

	tailOffset := vector.tailOffset()
	if index >= tailOffset {
		newTail := slices.Clone(vector.tail)
		newTail[index-tailOffset] = value
		return &Vector[E]{
			size:  vector.size,
			depth: vector.depth,
			root:  vector.root,
			tail:  newTail,
		}
	}

	root := vector.assocNode(vector.root, vector.depth, index, value)
	return &Vector[E]{
		size:  vector.size,
		depth: vector.depth,
		root:  root,
		tail:  vector.tail,
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
		return &Vector[E]{
			size: 1,
			tail: []E{value},
		}
	}

	if len(vector.tail) < vectorB {
		newTail := append(slices.Clip(vector.tail), value)
		return &Vector[E]{
			size:  vector.size + 1,
			depth: vector.depth,
			root:  vector.root,
			tail:  newTail,
		}
	}

	root, depth := vector.pushTailToTree()

	return &Vector[E]{
		size:  vector.size + 1,
		depth: depth,
		root:  root,
		tail:  []E{value},
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

func (vector *Vector[E]) pushValueToTree(root *vectorNode[E], depth int, index int, value E) (*vectorNode[E], int) {
	if root == nil {
		return vector.newPath(0, index, value).(*vectorNode[E]), 0
	}

	vecCap := vectorCapacity(depth)
	if index == vecCap {
		newRoot := &vectorNode[E]{
			children: make([]any, vectorB),
		}
		newRoot.children[0] = root
		newRoot.children[1] = vector.newPath(depth, index, value)
		return newRoot, depth + 1
	}

	if index > vecCap {
		panic(fmt.Sprintf("vector is broken, index %d exceeded capacity %d", index, vecCap))
	}

	return vector.pushNode(root, depth, index, value), depth
}

func (vector *Vector[E]) pushTailToTree() (*vectorNode[E], int) {
	root := vector.root
	depth := vector.depth
	offset := vector.tailOffset()

	for i, value := range vector.tail {
		index := offset + i
		root, depth = vector.pushValueToTree(root, depth, index, value)
	}

	return root, depth
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
		return &Vector[E]{}
	}

	if len(vector.tail) > 1 {
		newTail := slices.Clone(vector.tail[:len(vector.tail)-1])
		return &Vector[E]{
			size:  vector.size - 1,
			depth: vector.depth,
			root:  vector.root,
			tail:  newTail,
		}
	}

	if len(vector.tail) == 1 {
		if vector.tailOffset() == 0 {
			return &Vector[E]{}
		}

		newTail, newRoot, newDepth := vector.popTailFromTree()
		return &Vector[E]{
			size:  vector.size - 1,
			depth: newDepth,
			root:  newRoot,
			tail:  newTail,
		}
	}

	newRoot := vector.popNode(vector.root, vector.depth, vector.size-1)
	newDepth := vector.depth
	for newDepth > 0 && vector.hasOnlyOneChild(newRoot, 0) {
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
		tail:  vector.tail,
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

func (vector *Vector[E]) popTailFromTree() ([]E, *vectorNode[E], int) {
	tailOffset := vector.tailOffset()
	start := tailOffset - vectorB

	newTail := make([]E, 0, vectorB)
	for i := range vectorB {
		value, ok := vector.Get(start + i)
		if !ok {
			break
		}
		newTail = append(newTail, value)
	}

	newRoot := vector.root
	newDepth := vector.depth

	for removed := 0; removed < vectorB && tailOffset-1-removed >= 0; removed++ {
		index := tailOffset - 1 - removed
		newRoot = vector.popNode(newRoot, newDepth, index)

		for newDepth > 0 && vector.hasOnlyOneChild(newRoot, 0) {
			r, ok := newRoot.children[0].(*vectorNode[E])
			if !ok {
				panic(fmt.Sprintf("vector is brokent, unexpected child node %T at index=0", newRoot.children[0]))
			}
			newRoot = r
			newDepth--
		}
	}

	if start == 0 {
		newRoot = nil
		newDepth = 0
	}

	return newTail, newRoot, newDepth
}

func (vector *Vector[E]) hasOnlyOneChild(node *vectorNode[E], idx int) bool {
	if node == nil {
		return false
	}

	if idx < 0 || idx >= len(node.children) {
		return false
	}

	for childIndex, child := range node.children {
		if childIndex == idx {
			continue
		}

		if child != nil {
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

	if vector.root != nil {
		if !visitNode(vector.root, vector.depth, 0) {
			return
		}
	}

	tailOffset := vector.tailOffset()
	for i, value := range vector.tail {
		if !yield(tailOffset+i, value) {
			return
		}
	}
}

func (vector *Vector[E]) AllReversed(yield func(int, E) bool) {
	if vector == nil {
		return
	}

	var visitNode func(node *vectorNode[E], level int, indexBase int) bool

	visitNode = func(node *vectorNode[E], level int, indexBase int) bool {
		if level == 0 {
			for i := len(node.children) - 1; i >= 0; i-- {
				child := node.children[i]
				if child == nil {
					continue
				}

				if !yield(indexBase+i, child.(E)) {
					return false
				}
			}

			return true
		}

		for i, child := range slices.Backward(node.children) {
			if child == nil {
				continue
			}

			if !visitNode(child.(*vectorNode[E]), level-1, indexBase+i*vectorB) {
				return false
			}
		}

		return true
	}

	tailOffset := vector.tailOffset()
	for i := len(vector.tail) - 1; i >= 0; i-- {
		if !yield(tailOffset+i, vector.tail[i]) {
			return
		}
	}

	if vector.root != nil {
		visitNode(vector.root, vector.depth, 0)
	}
}

func (vector *Vector[E]) AllValues(yield func(E) bool) {
	if vector == nil {
		return
	}

	var visitNode func(node *vectorNode[E], level int, indexBase int) bool

	visitNode = func(node *vectorNode[E], level int, indexBase int) bool {
		if level == 0 {
			for _, child := range node.children {
				if child == nil {
					continue
				}

				if !yield(child.(E)) {
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

	if vector.root != nil {
		if !visitNode(vector.root, vector.depth, 0) {
			return
		}
	}

	for _, value := range vector.tail {
		if !yield(value) {
			return
		}
	}
}

func (vector *Vector[E]) Size() int {
	if vector == nil {
		return 0
	}
	return vector.size
}

func (vector *Vector[E]) String() string {
	str := &strings.Builder{}
	str.WriteString("(vector")

	for _, value := range vector.All {
		str.WriteByte(' ')
		fmt.Fprint(str, value)
	}

	str.WriteByte(')')
	return str.String()
}

func (vector *Vector[E]) EqualFn(other *Vector[E], equal func(a, b E) bool) bool {
	if vector == other {
		return true
	}

	if vector == nil || other == nil {
		return false
	}

	if vector.Size() != other.Size() {
		return false
	}

	next, stop := iter.Pull2(vector.All)
	defer stop()

	for _, otherValue := range other.All {
		_, value, ok := next()
		if !ok {
			return false
		}

		if !equal(value, otherValue) {
			return false
		}
	}

	return true
}
