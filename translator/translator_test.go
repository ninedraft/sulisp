package translator_test

import (
	"go/format"
	"go/token"
	"strings"
	"testing"

	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
	"github.com/ninedraft/sulisp/translator"
)

func TestTranslate(t *testing.T) {
	t.Parallel()
	t.Log(
		"Testing translator for simple cases",
	)

	input := strings.NewReader(`
		(package main)
		(import fmt strconv)
		
		(var 
			x 10
		 	y 20 
			x (float64 (+ x y 100)))

		(defn inc (x :- int => int)
			(+ x 1))

		(defn main ()  
			(go fmt.Println "Hello, world!"))
	`)

	lex := &lexer.Lexer{Input: input}
	if err := lex.Run(); err != nil {
		t.Fatal("lexing input", err)
	}

	p := &parser.Parser{Tokens: lex.Tokens}
	ast, errParse := p.Parse()
	if errParse != nil {
		t.Fatal("parsing input", errParse)
	}

	file, errTranslate := translator.TranslateFile(ast)
	if errTranslate != nil {
		t.Fatal("translating input", errTranslate)
	}

	result := &strings.Builder{}
	fset := token.NewFileSet()
	if err := format.Node(result, fset, file); err != nil {
		t.Fatal("formatting translated code", err)
	}
}
