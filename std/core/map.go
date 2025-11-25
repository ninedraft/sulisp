package core

import (
	"hash/maphash"
)

const (
	hashBranch   = 32
	hashBits     = 5
	hashBitMask  = hashBranch - 1
	hashMaxDepth = (64 + hashBits - 1) / hashBits
)

// HashFunc computes a hash for the provided key using the supplied maphash.Hash.
type HashFunc[K any] func(key K, h *maphash.Hash)

type EqualFunc[K any] func(a, B K) bool

// Map is an immutable hash table that canonically hashes keys with a provided callback.
type Map[K any, V any] struct {
	root   *mapNode[K, V]
	size   int
	seed   maphash.Seed
	hashFn HashFunc[K]
	equal  EqualFunc[K]
}

type mapNode[K any, V any] struct {
	children []any
}

type mapEntry[K any, V any] struct {
	hash  uint64
	key   K
	value V
}

// NewMap constructs an empty PersistentHash that relies on the provided hashFn for key hashing.
func NewMap[K any, V any](hashFn HashFunc[K], equal EqualFunc[K]) *Map[K, V] {
	if hashFn == nil {
		panic("core: hash function is required")
	}

	return &Map[K, V]{
		seed:   maphash.MakeSeed(),
		hashFn: hashFn,
		equal:  equal,
	}
}

// Size returns the number of entries held by the hashmap.
func (hashmap *Map[K, V]) Size() int {
	if hashmap == nil {
		return 0
	}
	return hashmap.size
}

// Get looks up the value associated with the provided key.
func (hashmap *Map[K, V]) Get(key K) (V, bool) {
	var zero V
	if hashmap == nil || hashmap.root == nil {
		return zero, false
	}

	entryHash := hashmap.hashKey(key)
	node := hashmap.root
	for level := 0; node != nil; level++ {
		idx := hashIndexAtLevel(entryHash, level)
		child := node.children[idx]
		switch typed := child.(type) {
		case nil:
			return zero, false
		case []mapEntry[K, V]:
			for _, entry := range typed {
				if hashmap.equal(entry.key, key) {
					return entry.value, true
				}
			}
			return zero, false
		case *mapNode[K, V]:
			node = typed
			continue
		default:
			panic("core: unexpected node child")
		}
	}

	return zero, false
}

// Put returns a new hash with the provided key and value stored. If the key already existed, the value is replaced.
func (hashmap *Map[K, V]) Put(key K, value V) *Map[K, V] {
	if hashmap == nil {
		panic("core: Put cannot be used on a nil PersistentHash")
	}

	entry := mapEntry[K, V]{
		key:   key,
		value: value,
		hash:  hashmap.hashKey(key),
	}

	var newRoot *mapNode[K, V]
	var added bool
	if hashmap.root == nil {
		newRoot = newHashNode[K, V]()
		newRoot.children[hashIndexAtLevel(entry.hash, 0)] = []mapEntry[K, V]{entry}
		added = true
	} else {
		newRoot, added = hashmap.root.withPut(0, entry, hashmap.equal)
	}

	if !added && newRoot == hashmap.root {
		return hashmap
	}

	size := hashmap.size
	if added {
		size++
	}

	return &Map[K, V]{
		root:   newRoot,
		size:   size,
		seed:   hashmap.seed,
		hashFn: hashmap.hashFn,
		equal:  hashmap.equal,
	}
}

// Remove returns a new hash that does not contain the provided key.
func (hashmap *Map[K, V]) Remove(key K) *Map[K, V] {
	if hashmap == nil || hashmap.root == nil {
		return hashmap
	}

	entryHash := hashmap.hashKey(key)
	newRoot, removed := hashmap.root.withRemove(0, entryHash, key, hashmap.equal)
	if !removed {
		return hashmap
	}

	return &Map[K, V]{
		root:   newRoot,
		size:   hashmap.size - 1,
		seed:   hashmap.seed,
		hashFn: hashmap.hashFn,
		equal:  hashmap.equal,
	}
}

// All returns an iterator over all stored key/value pairs.
func (hashmap *Map[K, V]) All(yield func(K, V) bool) {
	if hashmap == nil || hashmap.root == nil {
		return
	}

	var visit func(node *mapNode[K, V]) bool
	visit = func(node *mapNode[K, V]) bool {
		for _, child := range node.children {
			switch typed := child.(type) {
			case nil:
				continue
			case []mapEntry[K, V]:
				for _, entry := range typed {
					if !yield(entry.key, entry.value) {
						return false
					}
				}
			case *mapNode[K, V]:
				if !visit(typed) {
					return false
				}
			default:
				panic("core: unexpected node child")
			}
		}
		return true
	}

	visit(hashmap.root)
}

func (hashmap *Map[K, V]) Equal(other *Map[K, V], equal func(a, b V) bool) bool {
	if hashmap == nil && other == nil {
		return true
	}

	if hashmap == nil || other == nil {
		return false
	}

	if hashmap.size != other.size {
		return false
	}

	for key, value := range hashmap.All {
		otherValue, ok := other.Get(key)
		if !ok {
			return false
		}

		if !equal(value, otherValue) {
			return false
		}
	}

	return true
}

func (hashmap *Map[K, V]) Keys(yield func(K) bool) {
	if hashmap == nil || hashmap.root == nil {
		return
	}

	var visit func(node *mapNode[K, V]) bool
	visit = func(node *mapNode[K, V]) bool {
		for _, child := range node.children {
			switch typed := child.(type) {
			case nil:
				continue
			case []mapEntry[K, V]:
				for _, entry := range typed {
					if !yield(entry.key) {
						return false
					}
				}
			case *mapNode[K, V]:
				if !visit(typed) {
					return false
				}
			default:
				panic("core: unexpected node child")
			}
		}
		return true
	}

	visit(hashmap.root)
}

func (hashmap *Map[K, V]) Values(yield func(V) bool) {
	if hashmap == nil || hashmap.root == nil {
		return
	}

	var visit func(node *mapNode[K, V]) bool
	visit = func(node *mapNode[K, V]) bool {
		for _, child := range node.children {
			switch typed := child.(type) {
			case nil:
				continue
			case []mapEntry[K, V]:
				for _, entry := range typed {
					if !yield(entry.value) {
						return false
					}
				}
			case *mapNode[K, V]:
				if !visit(typed) {
					return false
				}
			default:
				panic("core: unexpected node child")
			}
		}
		return true
	}

	visit(hashmap.root)
}

func (hashmap *Map[K, V]) hashKey(key K) uint64 {
	var h maphash.Hash
	h.SetSeed(hashmap.seed)
	hashmap.hashFn(key, &h)
	return h.Sum64()
}

func newHashNode[K any, V any]() *mapNode[K, V] {
	return &mapNode[K, V]{
		children: make([]any, hashBranch),
	}
}

func (node *mapNode[K, V]) clone() *mapNode[K, V] {
	if node == nil {
		return nil
	}

	dup := &mapNode[K, V]{
		children: make([]any, len(node.children)),
	}
	copy(dup.children, node.children)
	return dup
}

func (node *mapNode[K, V]) isEmpty() bool {
	for _, child := range node.children {
		if child != nil {
			return false
		}
	}
	return true
}

func (node *mapNode[K, V]) withPut(level int, entry mapEntry[K, V], equal EqualFunc[K]) (*mapNode[K, V], bool) {
	idx := hashIndexAtLevel(entry.hash, level)
	child := node.children[idx]

	switch typed := child.(type) {
	case nil:
		newNode := node.clone()
		newNode.children[idx] = []mapEntry[K, V]{entry}
		return newNode, true
	case []mapEntry[K, V]:
		for i, existing := range typed {
			if equal(existing.key, entry.key) {
				entries := append([]mapEntry[K, V]{}, typed...)
				entries[i] = entry

				clone := node.clone()
				clone.children[idx] = entries
				return clone, false
			}
		}

		if level >= hashMaxDepth-1 {
			entries := append([]mapEntry[K, V]{}, typed...)
			entries = append(entries, entry)

			clone := node.clone()
			clone.children[idx] = entries
			return clone, true
		}

		sub := newHashNode[K, V]()
		for _, existing := range typed {
			sub, _ = sub.withPut(level+1, existing, equal)
		}
		sub, _ = sub.withPut(level+1, entry, equal)

		clone := node.clone()
		clone.children[idx] = sub
		return clone, true
	case *mapNode[K, V]:
		newChild, added := typed.withPut(level+1, entry, equal)
		if !added && newChild == typed {
			return node, false
		}

		clone := node.clone()
		clone.children[idx] = newChild
		return clone, added
	default:
		panic("core: unexpected node child")
	}
}

func (node *mapNode[K, V]) withRemove(level int, entryHash uint64, key K, equal EqualFunc[K]) (*mapNode[K, V], bool) {
	idx := hashIndexAtLevel(entryHash, level)
	child := node.children[idx]
	if child == nil {
		return node, false
	}

	switch typed := child.(type) {
	case []mapEntry[K, V]:
		for i, entry := range typed {
			if equal(entry.key, key) {
				if len(typed) == 1 {
					clone := node.clone()
					clone.children[idx] = nil
					if clone.isEmpty() {
						return nil, true
					}
					return clone, true
				}

				entries := append([]mapEntry[K, V]{}, typed[:i]...)
				entries = append(entries, typed[i+1:]...)

				clone := node.clone()
				clone.children[idx] = entries
				return clone, true
			}
		}
		return node, false
	case *mapNode[K, V]:
		newChild, removed := typed.withRemove(level+1, entryHash, key, equal)
		if !removed {
			return node, false
		}

		clone := node.clone()
		clone.children[idx] = newChild
		if newChild == nil && clone.isEmpty() {
			return nil, true
		}
		return clone, true
	default:
		panic("core: unexpected node child")
	}
}

func hashIndexAtLevel(hash uint64, level int) int {
	shift := uint(level * hashBits)
	if shift >= 64 {
		return 0
	}
	return int((hash >> shift) & hashBitMask)
}
