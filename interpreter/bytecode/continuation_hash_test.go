package bytecode

import (
	"hash/maphash"
	"testing"
)

func hashWithSeed(seed maphash.Seed, fn func(*maphash.Hash)) uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	fn(&h)
	return h.Sum64()
}

func TestContinuationHash(t *testing.T) {
	seed := maphash.MakeSeed()

	first := &Continuation{ResumePC: 10}
	second := &Continuation{ResumePC: 10}

	if hashWithSeed(seed, first.Hash) != hashWithSeed(seed, first.Hash) {
		t.Fatalf("expected continuation to hash consistently")
	}

	if hashWithSeed(seed, first.Hash) == hashWithSeed(seed, second.Hash) {
		t.Fatalf("expected distinct continuations to hash differently")
	}
}
