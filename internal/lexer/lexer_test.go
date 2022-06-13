package lexer_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/internal/lexer"
)

func TestLexer(test *testing.T) {
	lex := lexer.New(strings.NewReader(`
	(print "hello world")
	(print "multiline
	hello world")
	(let a (+ 1 2))
	()
	`))
	for {
		tok := lex.Scan()
		if tok.IsEnd() {
			break
		}
		test.Logf("%q", tok.Value)
	}
	if err := lex.Err(); err != nil {
		test.Log("error:", err)
	}
}
