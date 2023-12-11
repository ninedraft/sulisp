package scanner_test

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/lexer/scanner"
	"github.com/stretchr/testify/assert"
)

func TestScanner(t *testing.T) {
	t.Parallel()

	sc := scanner.New(strings.NewReader(input))

	got := &strings.Builder{}
	got.WriteRune(sc.Current())

	for {
		ru := sc.Scan()
		if ru == scanner.EOF {
			break
		}

		got.WriteRune(ru)
	}

	assert.Equal(t, input, got.String(), "scanned")
}

const input = `
abnode
nohoe
`

func TestScanner_Error(t *testing.T) {
	t.Parallel()

	const want = "a"

	in := &errReader{
		re: strings.NewReader(want),
	}

	sc := scanner.New(bufio.NewReader(in))

	got := &strings.Builder{}
	got.WriteRune(sc.Current())
	for {
		ru := sc.Scan()
		if ru == scanner.EOF {
			break
		}

		got.WriteRune(ru)
	}

	t.Logf("got: %x", got.String())

	assert.Equal(t, want, got.String(), "scanned")
	assert.ErrorIs(t, sc.Err(), errTest, "got error")
}

var errTest = errors.New("test error")

type errReader struct {
	re io.Reader
}

func (er *errReader) Read(dst []byte) (int, error) {
	n, err := er.re.Read(dst)
	if err != nil {
		err = errTest
	}
	return n, err
}
