package lexer

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/ninedraft/sulisp/ast"
)

type Lexer struct {
	src io.RuneScanner
	err error

	scan  scanFn
	width int

	tokens chan ast.Token
}

var endTok = ast.Token{Kind: ast.TokenEnd}

func New(src io.RuneScanner) *Lexer {
	lex := &Lexer{
		src:    src,
		tokens: make(chan ast.Token, 2),
	}
	lex.scan = lex.lexToken
	return lex
}

func (lex *Lexer) Err() error {
	return lex.err
}

func (lex *Lexer) Scan() ast.Token {
	for {
		select {
		case tok := <-lex.tokens:
			return tok
		default:
			lex.tryScan()
		}
	}
	panic("never reached")
}

func (lex *Lexer) tryScan() {
	if lex.scan == nil {
		close(lex.tokens)
		return
	}
	lex.scan = lex.scan()
}

type scanFn func() scanFn

func (lex *Lexer) lexToken() scanFn {
	for {
		ru := lex.read()
		if ru == eof {
			lex.tokens <- endTok
			return nil
		}
		if unicode.IsSpace(ru) {
			continue
		}
		lex.backup()

		if lex.expect(eq("(")) {
			lex.tokens <- ast.Token{Kind: ast.TokenLeftPar, Value: "("}
			return lex.lexToken
		}
		if lex.expect(eq(")")) {
			lex.tokens <- ast.Token{Kind: ast.TokenRightPar, Value: ")"}
			return lex.lexToken
		}
		if isAtom(lex.peek()) {
			return lex.lexAtom
		}
		if lex.expect(eq(`"`)) {
			return lex.lexString
		}
		return lex.errrorf("token: unexpected symbol %q", ru)
	}
}

func (lex *Lexer) lexAtom() scanFn {
	value := strings.Builder{}
	for {
		ru := lex.read()
		if ru == eof {
			return nil
		}
		if isAtom(ru) {
			value.WriteRune(ru)
			continue
		}
		lex.backup()

		v := value.String()
		lex.tokens <- ast.Token{Kind: ast.TokenAtom, Value: v, Flags: tokenFlags(v)}
		return lex.lexToken
	}
}

const escapeRune = '\\'

func (lex *Lexer) lexString() scanFn {
	value := strings.Builder{}
	for {
		ru := lex.read()
		if ru == eof {
			return nil
		}
		if ru == '"' {
			lex.tokens <- ast.Token{Kind: ast.TokenAtom, Value: value.String()}
			return lex.lexToken
		}
		switch {
		case ru == escapeRune && lex.expect(eq(`"`)):
			value.WriteRune('"')
			continue
		default:
			value.WriteRune(ru)
			continue
		}
	}
}

func isAtom(ru rune) bool {
	return unicode.IsLetter(ru) ||
		unicode.IsNumber(ru) ||
		strings.ContainsRune("_-+?!", ru)
}

func eq(set string) func(rune) bool {
	return func(ru rune) bool { return strings.ContainsRune(set, ru) }
}

func (lex *Lexer) expect(fns ...func(ru rune) bool) bool {
	ru := lex.read()
	for _, fn := range fns {
		if fn(ru) {
			return true
		}
	}
	lex.backup()
	return false
}

func (lex *Lexer) peek() rune {
	ru := lex.read()
	if ru != eof {
		lex.backup()
	}
	return ru
}

const eof = -1

func (lex *Lexer) read() rune {
	ru, width, err := lex.src.ReadRune()
	lex.width = width
	switch {
	case errors.Is(err, io.EOF):
		return eof
	case err != nil:
		lex.err = err
	}
	return ru
}

func (lex *Lexer) backup() {
	if lex.err != nil {
		return
	}
	lex.err = lex.src.UnreadRune()
}

func (lex *Lexer) errrorf(format string, args ...any) scanFn {
	if lex.err == nil {
		lex.err = fmt.Errorf(format, args...)
	}
	return nil
}

var (
	isInteger = regexp.MustCompile(`[+-]?[0-9]+`).MatchString
	isFloat   = regexp.MustCompile(`[+-]?[0-9]+\.[0-9]`).MatchString
)

func tokenFlags(value string) ast.TokenFlags {
	var flags ast.TokenFlags
	if strings.HasPrefix(value, ":") {
		flags |= ast.FSymbol
	}
	if isInteger(value) {
		flags |= ast.FInt
	}
	if isFloat(value) {
		flags |= ast.FFloat
	}
	return flags
}
