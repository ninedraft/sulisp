package lexer_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/ninedraft/sulisp/language"
	"github.com/ninedraft/sulisp/lexer"
)

//go:embed testdata/valid.lisp
var validInput []byte

func TestLexer(t *testing.T) {
	t.Parallel()
	t.Log(
		"Testing parsing of lexem scanner",
	)

	lex := lexer.Lexer{
		File:  t.Name(),
		Input: bytes.NewReader(validInput),
	}

	err := lex.Run()

	if err != nil {
		t.Error("unexpected error", err)
	}

	for i, tok := range lex.Tokens {
		expect := expectedValidTokens[i]
		if tok.Kind != expect.Kind {
			t.Errorf("unexpected token kind at %d: got %s, expected %s", i, tok.Kind, expect.Kind)
		}
		if tok.Value != expect.Value {
			t.Errorf("unexpected token value at %d: got %s, expected %s", i, tok.Value, expect.Value)
		}
	}
}

var expectedValidTokens = []*language.Token{
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "defun"},
	{Kind: language.TokenSymbol, Value: "test-function"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "arg1"},
	{Kind: language.TokenSymbol, Value: "arg2"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "let"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "list"},
	{Kind: language.TokenQuote, Value: "'"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenFloat, Value: "1"},
	{Kind: language.TokenFloat, Value: "2"},
	{Kind: language.TokenFloat, Value: "3"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "string"},
	{Kind: language.TokenStr, Value: "\"Hello, World!\""},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "number"},
	{Kind: language.TokenFloat, Value: "42"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "boolean"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "nil-value"},
	{Kind: language.TokenSymbol, Value: "nil"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"List: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "list"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"String: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "string"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Number: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "number"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Boolean: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "boolean"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Nil value: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "nil-value"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "+"},
	{Kind: language.TokenSymbol, Value: "arg1"},
	{Kind: language.TokenSymbol, Value: "arg2"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "define-constant"},
	{Kind: language.TokenSymbol, Value: "PI"},
	{Kind: language.TokenFloat, Value: "3.14159"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "setf"},
	{Kind: language.TokenSymbol, Value: "variable"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "+"},
	{Kind: language.TokenFloat, Value: "10"},
	{Kind: language.TokenFloat, Value: "20"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "if"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "<"},
	{Kind: language.TokenSymbol, Value: "variable"},
	{Kind: language.TokenFloat, Value: "50"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Variable is less than 50.~%\""},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Variable is greater than or equal to 50.~%\""},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "do"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "i"},
	{Kind: language.TokenFloat, Value: "0"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "+"},
	{Kind: language.TokenSymbol, Value: "i"},
	{Kind: language.TokenFloat, Value: "1"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: ">"},
	{Kind: language.TokenSymbol, Value: "="},
	{Kind: language.TokenSymbol, Value: "i"},
	{Kind: language.TokenFloat, Value: "10"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenLBrace, Value: "("},
	{Kind: language.TokenSymbol, Value: "format"},
	{Kind: language.TokenSymbol, Value: "t"},
	{Kind: language.TokenStr, Value: "\"Iteration: ~a~%\""},
	{Kind: language.TokenSymbol, Value: "i"},
	{Kind: language.TokenRBrace, Value: ")"},
	{Kind: language.TokenRBrace, Value: ")"},
}
