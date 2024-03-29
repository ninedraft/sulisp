package lexer_test

import (
	"strings"
	"testing"

	language "github.com/ninedraft/sulisp/language/tokens"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLex_Comment(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
	; comment ending with newline
	; comment ending with semicolon
	;
	`)

	want := []language.Token{
		{Kind: language.TokenComment, Value: "; comment ending with newline"},
		{Kind: language.TokenComment, Value: "; comment ending with semicolon"},
		{Kind: language.TokenComment, Value: ";"},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", i, got.Pos)
	}
}

func TestLex_Numbers(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
		1 2
		3.5 1e1
	`)

	want := []language.Token{
		{Kind: language.TokenInt, Value: "1"},
		{Kind: language.TokenInt, Value: "2"},
		{Kind: language.TokenFloat, Value: "3.5"},
		{Kind: language.TokenFloat, Value: "1e1"},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", got.Pos)
	}
}

func TestLex_Strings(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
		"string without newline"
		
		"string with new
line"

		"string with e\\scapes\n"

		""
	`)

	want := []language.Token{
		{Kind: language.TokenStr, Value: `"string without newline"`},
		{Kind: language.TokenStr, Value: `"string with new\nline"`},
		{Kind: language.TokenStr, Value: `"string with e\\scapes\n"`},
		{Kind: language.TokenStr, Value: `""`},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", i, got.Pos)
	}
}

func TestLex_Strings_BadEscape(t *testing.T) {
	t.Parallel()

	lex := lexer.NewLexer(t.Name(), strings.NewReader(`"\g"`))

	_, err := lex.Next()

	assert.Error(t, err)
}

func TestLex_Strings_UnexpectedEOF(t *testing.T) {
	t.Parallel()

	lex := lexer.NewLexer(t.Name(), strings.NewReader(`"sasd`))

	_, err := lex.Next()

	assert.Error(t, err)
}

func TestLex_Keyword(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
		:a,:b
		:bvasd
		:_das
		:123
	`)

	want := []language.Token{
		{Kind: language.TokenKeyword, Value: `:a`},
		{Kind: language.TokenKeyword, Value: `:b`},
		{Kind: language.TokenKeyword, Value: `:bvasd`},
		{Kind: language.TokenKeyword, Value: `:_das`},
		{Kind: language.TokenKeyword, Value: ":123"},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", i, got.Pos)
	}
}

func TestLex_Symbol(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
		a, b,
		c_, d5
		+ -
	`)

	want := []language.Token{
		{Kind: language.TokenSymbol, Value: `a`},
		{Kind: language.TokenSymbol, Value: `b`},
		{Kind: language.TokenSymbol, Value: `c_`},
		{Kind: language.TokenSymbol, Value: `d5`},

		{Kind: language.TokenSymbol, Value: `+`},
		{Kind: language.TokenSymbol, Value: `-`},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", i, got.Pos)
	}
}

func TestLex_SExp(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `(applorange)`)

	want := []language.Token{
		{Kind: language.TokenLParen, Value: `(`},
		{Kind: language.TokenSymbol, Value: `applorange`},
		{Kind: language.TokenRParen, Value: `)`},
	}

	require.Len(t, tokens, len(want), "len(tokens)==len(want)")

	for i, expect := range want {
		got := tokens[i]

		assert.EqualValues(t, expect.Kind, got.Kind, "[%d] %s token kind", i, got.Pos)
		assert.EqualValues(t, expect.Value, got.Value, "[%d] %s token value", i, got.Pos)
	}
}

func readTokens(t *testing.T, input string) []*language.Token {
	lex := lexer.NewLexer(t.Name(), strings.NewReader(input))

	tokens := []*language.Token{}

	for {
		tok, errTok := lex.Next()
		if tok == nil || tok.Kind == language.TokenEOF {
			break
		}

		require.NoError(t, errTok, "lexer error")

		tokens = append(tokens, tok)
	}

	return tokens
}
