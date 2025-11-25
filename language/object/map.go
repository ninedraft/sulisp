package object

import (
	"hash/maphash"
	"strings"

	"github.com/ninedraft/sulisp/std/core"
)

type Map struct {
	*core.Map[Object, Object]
}

func NewMap() *Map {
	hm := core.NewMap[Object, Object](
		Object.Hash,
		func(a, b Object) bool {
			ord, _ := a.Compare(b)
			return ord == 0
		},
	)

	return &Map{
		Map: hm,
	}
}

func (hashmap *Map) Compare(other Object) (int, bool) {
	otherMap, ok := other.(*Map)
	if !ok {
		return 0, false
	}

	return 0, hashmap.Map.Equal(otherMap.Map, func(a, b Object) bool {
		ord, ok := a.Compare(b)
		return ord == 0 && ok
	})
}

func (hashmap *Map) Kind() Kind {
	return ObjMap
}

func (hashmap *Map) Hash(h *maphash.Hash) {
	maphash.WriteComparable(h, hashmap.Kind())

	for key, value := range hashmap.All {
		key.Hash(h)
		value.Hash(h)
	}
}

func (hashmap *Map) Inspect() string {
	buf := &strings.Builder{}

	const prefix = "{"
	buf.WriteString(prefix)

	for key, value := range hashmap.All {
		if buf.Len() > len(prefix) {
			buf.WriteString(" ")
		}
		buf.WriteString(key.Inspect())
		buf.WriteString(": ")
		buf.WriteString(value.Inspect())
	}
	buf.WriteRune('}')
	return buf.String()
}
