package match_test

import (
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/language"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/ninedraft/sulisp/std/internal/match"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	t.Parallel()
	t.Log(
		`Check AST matching with Match function.`,
	)

	ast := testParse(t, `(x :- int => int)`)

	const typeSep = language.Keyword("-")
	const resultSep = language.Symbol("=>")

	var name, type1, type2 language.Symbol

	ok := match.Match(ast[0],
		match.P(&name), typeSep, match.P(&type1), resultSep, match.P(&type2))

	require.Truef(t, ok, "match failed")

	require.Equal(t, "x", name.String())
	require.Equal(t, "int", type1.String())
	require.Equal(t, "int", type2.String())
}

func TestMatch_Special(t *testing.T) {
	t.Parallel()
	t.Log(
		`Check special AST matching with Match function.`,
	)

	ast := testParse(t, `(+ y 1)`)

	ok := match.Match(ast[0], language.Symbol("+"), match.AnyMore)

	require.Truef(t, ok, "match failed")
}

func testParse(t *testing.T, src string) []language.Sexp {
	t.Helper()

	lex := &lexer.Lexer{
		File:  t.Name(),
		Input: strings.NewReader(src),
	}

	if err := lex.Run(); err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	pr := &parser.Parser{
		Tokens: lex.Tokens,
	}

	parsed, errParse := pr.Parse()
	if errParse != nil {
		t.Fatalf("parser error: %v", errParse)
	}

	return parsed
}
