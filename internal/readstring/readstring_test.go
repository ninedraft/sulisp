package readstring_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/internal/readstring"
)

func TestRead(test *testing.T) {
	t := func(name string, input, expected string, err error) {
		test.Run(name, func(test *testing.T) {
			re := strings.NewReader(input)

			got, errRead := readstring.Read(re)

			if !errors.Is(errRead, err) {
				test.Errorf("got error:      %v", errRead)
				test.Errorf("expected error: %v", err)
			}
			if got != expected {
				test.Errorf("got:      %q", got)
				test.Errorf("expected: %q", expected)
			}
		})
	}

	t("simple literal",
		`"abcd 123 _ собака"`,
		`abcd 123 _ собака`, nil)

	t("escapes",
		`"\r\s\t\"\\"`,
		"\r \t\"\\", nil)

	t("bad escape",
		`"\_"`,
		"", readstring.ErrUnexpectedRune)

	t("missing first quote",
		`asdb"`,
		"", readstring.ErrUnexpectedRune)

	t("missing last quote",
		`"asdb`,
		`asdb`, io.EOF)

	t("empty input",
		``,
		``, io.EOF)

	t("multiline string",
		`"first line
		second line"`,
		"first line\n		second line", nil)
}
