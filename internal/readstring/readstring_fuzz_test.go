package readstring_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/internal/readstring"
)

func FuzzRead(fuzz *testing.F) {
	fuzz.Add(`"abcd 123 _ собака"`)
	fuzz.Add(`"\r\s\t\"\\"`)
	fuzz.Add(`asdb"`)
	fuzz.Add(`"asdb`)
	fuzz.Add(``)
	fuzz.Add(`"first line
	second line"`)
	fuzz.Add(string([]byte{0, 0, 0, 0}))

	fuzz.Fuzz(func(_ *testing.T, input string) {
		_, _ = readstring.Read(strings.NewReader(input))
	})
}
