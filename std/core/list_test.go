package core_test

import (
	"testing"

	"github.com/ninedraft/sulisp/std/core"
)

func TestList(t *testing.T) {
	a, b := core.ListNew[core.Int](1, 2, 3), core.ListNew[core.Int](1, 3)

	t.Error(core.Eq(a, b))
}
