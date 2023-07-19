package lexer_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/language"
	"github.com/ninedraft/sulisp/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLexer_FuncSignature(t *testing.T) {
	t.Parallel()
	t.Log(
		"Testing lexer for function signature",
	)

	const input = `(x :- int => int)`

	lex := &lexer.Lexer{
		File:  t.Name(),
		Input: strings.NewReader(input),
	}

	if err := lex.Run(); err != nil {
		t.Fatal("lexing input", err)
	}

	assert := func(i int, kind language.TokenKind, value string) {
		t.Helper()
		tok := lex.Tokens[i]
		if tok.Kind != kind {
			t.Errorf("unexpected token kind at %d: got %s (%q), expected %s", i, tok.Kind, tok.Value, kind)
		}
		if tok.Value != value {
			t.Errorf("unexpected token value at %d: got %s, expected %s", i, tok.Value, value)
		}
	}

	if len(lex.Tokens) != 7 {
		t.Errorf("unexpected number of tokens: got %d, expected %d", len(lex.Tokens), 7)
	}

	assert(0, language.TokenLBrace, "(")
	assert(1, language.TokenSymbol, "x")
	assert(2, language.TokenKeyword, "-")
	assert(3, language.TokenSymbol, "int")
	assert(4, language.TokenSymbol, "=>")
	assert(5, language.TokenSymbol, "int")
	assert(6, language.TokenRBrace, ")")
}

func TestLexer_S(t *testing.T) {
	t.Parallel()

	tokens := lexString(t, `(>=)`)

	tok := tokens[1]
	if tok.Value != ">=" {
		t.Errorf("unexpected token value: got %s, expected %s", tok.Value, ">=")
	}
}

func TestLexer_Comments(t *testing.T) {
	t.Parallel()
	t.Log(`
		Comments have following format:
		
		; comment \n

		Last space character is not included in comment value.
	`)

	tokens := lexString(t, `
		; 1
		a
		b ; 2`)

	assert.Equal(t, language.TokenComment, tokens[0].Kind)
	assert.Equal(t, " 1", tokens[0].Value)

	assert.Equal(t, language.TokenComment, tokens[3].Kind)
	assert.Equal(t, " 2", tokens[3].Value)
}

func lexString(t *testing.T, input string) []*language.Token {
	l := &lexer.Lexer{
		File:  t.Name(),
		Input: strings.NewReader(input),
	}

	if err := l.Run(); err != nil {
		t.Fatalf("lexer.Run() failed: %v", err)
	}

	return l.Tokens
}
