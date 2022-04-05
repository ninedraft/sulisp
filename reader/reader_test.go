package reader_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/reader"
)

func TestReader(test *testing.T) {
	var input = strings.NewReader(`
	(
		(+ 1 2)
		(def hello (who)
			(print "hello," who))
	)
	`)
	var a, errRead = reader.Read(input, "testdata")
	if errRead != nil {
		test.Errorf("unexpected error: %v", errRead)
		return
	}
	test.Log(a)
}
