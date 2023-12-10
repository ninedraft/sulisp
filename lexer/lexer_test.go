package lexer_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/language"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLex_Comment(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(`
	; comment ending with newline
	; comment ending with semicolon
	;
	`)

	lex := lexer.NewLexer(t.Name(), input)

	comments := []string{}

	for {
		tok, errTok := lex.Next()
		if tok == nil {
			break
		}

		require.NoError(t, errTok, "lexer error")
		require.Equal(t, language.TokenComment, tok.Kind, "tok: %v", tok)

		comments = append(comments, tok.Value)
	}

	require.ElementsMatch(t, []string{
		"; comment ending with newline\n",
		"; comment ending with semicolon\n",
		";\n",
	}, comments, "got comments")
}

func TestLex_Numbers(t *testing.T) {
	t.Parallel()

	tokens := readTokens(t, `
		1 2
		3.5 1e1
	`)

	t.Log(tokens)

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

func readTokens(t *testing.T, input string) []*language.Token {
	lex := lexer.NewLexer(t.Name(), strings.NewReader(input))

	tokens := []*language.Token{}

	for {
		tok, errTok := lex.Next()
		if tok == nil {
			break
		}

		require.NoError(t, errTok, "lexer error")

		tokens = append(tokens, tok)
	}

	return tokens
}
