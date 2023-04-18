package core

import (
	"reflect"
	"strings"
)

func Eq[E Value](a, b E) bool {
	if a, ok := Value(a).(ValueEq[E]); ok {
		return a.Eq(b)
	}

	if b, ok := Value(b).(ValueEq[E]); ok {
		return b.Eq(a)
	}

	if reflect.TypeOf(a).Comparable() {
		return any(a) == any(b)
	}

	return reflect.DeepEqual(a, b)
}

func Contains[E Value](collection Value, value E) bool {
	switch collection := collection.(type) {
	case interface{ Contains(E) bool }:
		return collection.Contains(value)
	case Collection[E]:
		return seqContains(collection.Seq(), value)
	case String:
		return stringContains(collection, value)
	default:
		panic("Contais: unsupported type: " + Type(collection).String())
	}
}

func seqContains[E Value](seq Seq[E], value E) bool {
	for !seq.Empty() {
		first, ok := seq.First()
		if !ok {
			return false
		}

		if Eq(first, value) {
			return true
		}

		seq = seq.Next()
	}

	return false
}

func stringContains(str String, value Value) bool {
	if value, ok := value.(String); ok {
		return strings.Contains(str.As(), value.As())
	}

	return false
}
