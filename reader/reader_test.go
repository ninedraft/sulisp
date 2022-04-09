package reader_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/reader"
)

func TestReader(test *testing.T) {
	var t = func(name string, input, expectedAST string, err bool) {
		test.Run(name, func(test *testing.T) {
			var got, errRead = reader.Read(strings.NewReader(input), test.Name())

			switch {
			case errRead != nil && !err:
				test.Errorf("unexpected error: %s", errRead)
			case errRead == nil && err:
				test.Errorf("an error is expected, got nil")
			case errRead == nil && !err:
				assertEq(test, got.String(), expectedAST, "unexpected AST")
			}
		})
	}
	t("hello world",
		`
		(+ 1 2)
		(def hello (who)
			(print "hello," who))
	`,
		"(((atom +)(int 1) (int 2) )"+
			"((atom def)(atom hello) ((atom who)) "+
			"((atom print)(string \"hello,\") "+
			"(atom who) ) ) )",
		false,
	)
}

func assertEq[E comparable](t testing.TB, got, expected E, msg string, args ...any) {
	t.Helper()
	if got != expected {
		t.Errorf(msg, args...)
		t.Logf("Expected:")
		t.Logf("\t%v", expected)
		t.Logf("Got:")
		t.Logf("\t%v", got)
	}
}
