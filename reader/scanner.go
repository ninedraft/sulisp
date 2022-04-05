package reader

import (
	"fmt"
	"io"
	"text/scanner"
	"unicode"

	"github.com/ninedraft/sulisp/ast"
	"github.com/ninedraft/sulisp/internal/multierr"
	"golang.org/x/exp/slices"
)

func Scan(src io.Reader, fname string) *Scanner {
	var re = &Scanner{}

	re.sc = &scanner.Scanner{}
	re.sc.Init(src)
	re.sc.Error = func(sc *scanner.Scanner, msg string) {
		var err = fmt.Errorf("%s: %s", sc.Position, msg)
		re.err = multierr.Combine(re.err, err)
	}
	re.sc.IsIdentRune = isIdentRune
	re.sc.Filename = fname

	return re
}

type Scanner struct {
	sc    *scanner.Scanner
	token ast.Token
	err   error
}

func (re *Scanner) Scan() bool {
	var tok = re.sc.Scan()
	if tok == scanner.EOF || re.err != nil {
		return false
	}
	var kind ast.TokenKind
	switch tok {
	case scanner.Ident:
		kind = ast.TokenAtom
	case scanner.String:
		kind = ast.TokenString
	case scanner.Int:
		kind = ast.TokenInt
	case scanner.Float:
		kind = ast.TokenFloat
	case scanner.Char:
		kind = ast.TokenChar
	case '(':
		kind = ast.TokenLeftParen
	case ')':
		kind = ast.TokenRightParen
	}
	re.token.Kind = kind
	re.token.Value = re.sc.TokenText()
	re.token.Pos = re.sc.Position
	return true
}

func (re *Scanner) Err() error { return re.err }

func (re *Scanner) Token() ast.Token { return re.token }

var readerRunes = []rune("!?{}[]()")
var quoteRunes = []rune("\"`'")

func isIdentRune(ru rune, i int) bool {
	var startsWithDigit = i == 0 && unicode.IsDigit(ru)
	var startsWithMinus = i == 0 && ru == '-'
	var isReader = slices.Contains(readerRunes, ru)
	var startOfQuoted = i == 0 && slices.Contains(quoteRunes, ru)
	var isSpace = unicode.IsSpace(ru)
	var isIdent = !startsWithDigit && !startsWithMinus &&
		!isReader && !startOfQuoted &&
		!isSpace
	return ru > 0 && isIdent
}
