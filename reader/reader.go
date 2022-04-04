package reader

import (
	"fmt"
	"io"
	"text/scanner"
	"unicode"

	"github.com/ninedraft/sulisp/ast"
	"github.com/ninedraft/sulisp/internal/multierr"
)

func Scan(src io.Reader, fname string) *Scanner {
	var re = &Scanner{}
	re.sc = &scanner.Scanner{
		Error: func(sc *scanner.Scanner, msg string) {
			var err = fmt.Errorf("%s: %s", sc.Position, msg)
			re.err = multierr.Combine(re.err, err)
		},
		IsIdentRune: isIdentRune,
	}
	re.sc.Init(src)
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
		kind = ast.TokenLeftRB
	case ')':
		kind = ast.TokenRightRB
	}
	re.token.Kind = kind
	re.token.Value = re.sc.TokenText()
	return true
}

func (re *Scanner) Err() error { return re.err }

func (re *Scanner) Token() ast.Token { return re.token }

func isIdentRune(ru rune, i int) bool {
	return unicode.IsLetter(ru) ||
		(i > 0 && unicode.IsDigit(ru)) ||
		(i > 0 && ru == ':') ||
		(ru == '#') ||
		(ru == '@')
}
