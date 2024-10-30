package seq

import (
	"fmt"
	"iter"
	"slices"
)

func CollectErr[E any](seq iter.Seq2[E, error]) ([]E, error) {
	var items []E
	for item, err := range seq {
		if err != nil {
			return items, err
		}

		items = append(items, item)
	}
	return items, nil
}

func SlicePairs[E any](items []E) iter.Seq2[E, E] {
	if len(items)%2 != 0 {
		panic(fmt.Sprintf("seq.SlicePairs: want an event number of items, got %d", len(items)))
	}

	return func(yield func(E, E) bool) {
		for pair := range slices.Chunk(items, 2) {
			if !yield(pair[0], pair[1]) {
				return
			}
		}
	}
}
