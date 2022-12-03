package readstring_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/internal/readstring"
)

var _BenchmarkRead string

func BenchmarkRead(bench *testing.B) {
	re := &strings.Reader{}
	const input = `"123 456 \r \s \n
					cos muas клёцки
					solipsum fish language"`

	bench.ReportAllocs()
	for i := 0; i < bench.N; i++ {
		re.Reset(input)
		_BenchmarkRead, _ = readstring.Read(re)
	}
}
