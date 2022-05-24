package set

import "golang.org/x/exp/maps"

type Set[E comparable] map[E]struct{}

func From[E comparable](values []E) Set[E] {
	set := make(Set[E], len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func (set Set[E]) Put(values ...E) {
	for _, value := range values {
		set[value] = struct{}{}
	}
}

func (set Set[E]) Contains(value E) bool {
	_, ok := set[value]
	return ok
}

func (set Set[E]) IsSubset(other Set[E]) bool {
	if len(set) > len(other) {
		return false
	}
	for elem := range set {
		if !other.Contains(elem) {
			return false
		}
	}
	return true
}

func (set Set[E]) Values() []E {
	return maps.Keys(set)
}
