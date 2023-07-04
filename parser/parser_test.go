package parser_test

import (
	"bytes"
	"encoding/json"
	"testing"

	_ "embed"

	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

//go:embed testdata/valid.lisp
var validInput []byte

func TestParse(t *testing.T) {
	t.Parallel()
	t.Log(
		"Testing parser for simple cases",
	)

	lex := lexer.Lexer{
		File:  t.Name(),
		Input: bytes.NewReader(validInput),
	}

	if err := lex.Run(); err != nil {
		t.Error("lexing: unexpected error", err)
	}

	pr := &parser.Parser{
		Tokens: lex.Tokens,
	}

	root, errParse := pr.Parse()

	if errParse != nil {
		t.Error("parsing: unexpected error", errParse)
	}

	g, _ := json.Marshal(root)
	got := gjson.ParseBytes(g)

	// (defun test-function (arg1 arg2) ...)	// (defun test-function (arg1 arg2) ...)
	assert.ElementsMatch(t, []string{"arg1", "arg2"}, got.Get(`#(1="test-function").2`).Value())

	// (define-constant PI 3.14159)
	assert.EqualValues(t, 3.14159, got.Get(`#(1="PI").2.Value`).Value())
}
