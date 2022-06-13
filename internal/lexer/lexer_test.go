package lexer_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/ast"
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

func FuzzLexer(fuzz *testing.F) {
	fuzz.Add(`
	(print "hello world")
	(print "multiline
	hello world")
	(let a (+ 1 2))
	()
	`)
	fuzz.Fuzz(func(test *testing.T, input string) {
		tokens, errTokens := parse(input)
		if errTokens != nil {
			return
		}
		parsed := joinTokens(tokens)

		reparsed, errReparsed := parse(parsed)
		if errReparsed != nil {
			test.Errorf("unexpected error: %v", errReparsed)
			return
		}
		assertEqualSlices(test, tokens, reparsed, "tokens")
	})
}

func assertEqualSlices[E comparable](test *testing.T, expected, got []E, format string, args ...any) {
	if len(expected) != len(got) {
		test.Errorf("%d tokens are expected, got %d", len(expected), len(got))
		return
	}
	for i, tok := range expected {
		if got[i] != tok {
			test.Errorf("token %d: %#v is expected, got %#v", i, got[i], tok)
		}
	}
}

func parse(input string) ([]ast.Token, error) {
	lex := lexer.New(strings.NewReader(input))
	var tokens []ast.Token
	for {
		tok := lex.Scan()
		if tok.IsEnd() {
			break
		}
		tokens = append(tokens, tok)
	}
	if err := lex.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func joinTokens(tokens []ast.Token) string {
	str := strings.Builder{}
	for _, value := range tokens {
		str.WriteString(value.Value)
	}
	return str.String()
}
