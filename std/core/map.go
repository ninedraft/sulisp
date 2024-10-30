package core

import (
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Hashable interface {
	Value
	comparable
}

type HashMap[K Hashable, V Value] map[K]V

func (HashMap[K, V]) Kind() Value {
	var key K
	var value V
	return newTypeSpec("core.HashMap", map[Keyword]Value{
		":key":   key.Kind(),
		":value": value.Kind(),
	})
}

func (m HashMap[K, V]) String() string {
	if len(m) == 0 {
		return "(hashmap)"
	}

	str := &strings.Builder{}
	str.WriteString("(hashmap")

	for key, value := range m {
		str.WriteByte(' ')
		str.WriteString(key.String())
		str.WriteByte(' ')
		str.WriteString(value.String())
	}

	str.WriteByte(')')

	return str.String()
}

func (m HashMap[K, V]) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m HashMap[K, V]) Keys() *List[K] {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return ListNew(keys...)
}

func (m HashMap[K, V]) Values() *List[V] {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return ListNew(values...)
}

func (m HashMap[K, V]) Get(key K) (V, bool) {
	if m == nil {
		var value V
		return value, false
	}

	value, ok := m[key]
	return value, ok
}

func (m HashMap[K, V]) Put(key K, value V) {
	m[key] = value
}

func (m HashMap[K, V]) Remove(key K) {
	if m != nil {
		delete(m, key)
	}
}

func (m HashMap[K, V]) Clear() {
	maps.Clear(m)
}

func (m HashMap[K, V]) Len() int {
	return len(m)
}

func (m HashMap[K, V]) Eq(other HashMap[K, V]) bool {
	if m == nil {
		return other == nil
	}
	return maps.EqualFunc(m, other, Eq[V])
}

func (m HashMap[K, V]) Contains(key K) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

type hashmapSeq[K Hashable, V Value] struct {
	keys []K
	hm   HashMap[K, V]
}

func (m HashMap[K, V]) Seq() Seq[*List[Value]] {
	keys := maps.Keys(m)
	return &hashmapSeq[K, V]{
		keys: keys,
		hm:   m,
	}
}

func (seq *hashmapSeq[K, V]) First() (*List[Value], bool) {
	if seq.Empty() {
		return nil, false
	}
	key := seq.keys[0]

	return ListNew[Value](key, seq.hm[key]), true
}

func (seq *hashmapSeq[K, V]) Next() Seq[*List[Value]] {
	if seq.Empty() {
		return seq
	}

	return &hashmapSeq[K, V]{
		keys: seq.keys[1:],
		hm:   seq.hm,
	}
}

func (seq *hashmapSeq[K, V]) Empty() bool {
	return seq == nil || len(seq.keys) == 0 || len(seq.hm) == 0
}

func (seq *hashmapSeq[K, V]) Len() int {
	return len(seq.keys)
}

func (seq *hashmapSeq[K, V]) Contains(pair *List[Value]) bool {
	if seq.Empty() {
		return false
	}

	k, ok := pair.First()
	if !ok {
		return false
	}
	key, ok := k.(K)
	if !ok {
		return false
	}

	v, ok := pair.Next().First()
	if !ok {
		return false
	}
	value, ok := v.(V)
	if !ok {
		return false
	}

	if slices.Contains(seq.keys, key) {
		return Eq(seq.hm[key], value)
	}

	return false
}
