package set_test

import (
	"testing"

	"github.com/ninedraft/sulisp/internal/collections/set"
)

func TestSetPutContains(test *testing.T) {
	var expected []int
	for i := 0; i < 10; i++ {
		expected = append(expected, i)
	}
	set := set.Set[int]{}
	set.Put(expected...)

	for _, x := range expected {
		if !set.Contains(x) {
			test.Errorf("value %d expected to be in set", x)
		}
	}
}

func TestSetSubset(test *testing.T) {
	var expected []int
	for i := 0; i < 10; i++ {
		expected = append(expected, i)
	}

	a := set.From(expected)
	b := set.From(expected[:len(expected)/2])
	c := set.From([]int{100, 200, 300})

	if !b.IsSubset(a) {
		test.Errorf("b%v is expected to be a subset of a%v", b.Values(), a.Values())
	}
	if c.IsSubset(a) || c.IsSubset(b) {
		test.Errorf("c%v is not expected to be a subset of a%v or b%v", c.Values(), a.Values(), b.Values())
	}
}

func TestSetFromValues(test *testing.T) {
	values := []int{1, 2, 3, 4, 5, 6, 6, 8, 9, 10}
	a := set.From(values)
	b := set.From(a.Values())

	if !a.IsSubset(b) || !b.IsSubset(a) {
		test.Errorf("a%v expected to be equal to b%v", a.Values(), b.Values())
	}
}
