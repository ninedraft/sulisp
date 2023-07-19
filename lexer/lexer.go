package lexer

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/ninedraft/sulisp/language"
)

type Lexer struct {
	File   string
	Input  io.RuneScanner
	Tokens []*language.Token

	pos language.Position
}

func (lexer *Lexer) Run() error {
	for {
		token, err := lexer.Next()
		if err != nil {
			return err
		}
		if token == nil {
			break
		}
		lexer.Tokens = append(lexer.Tokens, token)
	}
	return nil
}

const brackets = "[](){}"

func (lexer *Lexer) Next() (*language.Token, error) {
	for {
		r, _, errRead := lexer.Input.ReadRune()
		switch {
		case errors.Is(errRead, io.EOF):
			return nil, nil
		case errRead != nil:
			return nil, lexer.errPos(errRead)
		}
		lexer.updatePos(r)

		switch {
		case unicode.IsSpace(r) || r == ',':
			continue
		case r == ';':
			return lexer.readComment()
		case r == '\'':
			return lexer.newToken(language.TokenQuote, "'"), nil
		case strings.ContainsRune(brackets, r):
			kind := language.TokenKind(r)
			return lexer.newToken(kind, kind.String()), nil
		case r == '"':
			return lexer.readString()
		default:
			return lexer.readAtom(r)
		}
	}
}

func (lexer *Lexer) readComment() (*language.Token, error) {
	comment, err := readUntil(lexer.Input, func(ru rune) bool {
		return ru != '\n'
	})
	if err != nil {
		err = fmt.Errorf("reading comment: %w", err)
		return nil, lexer.errPos(err)
	}

	comment = strings.TrimRightFunc(comment, unicode.IsSpace)

	return lexer.newToken(language.TokenComment, comment), nil
}

const escapable = `\"nrts`

var errBadStringLit = errors.New("bad string literal")

func (lexer *Lexer) readString() (*language.Token, error) {
	buf := &strings.Builder{}

	// already know that first rune is '"'
	buf.WriteRune('"')

	const (
		stateScan = iota
		stateEscape
	)
	var state = stateScan

scan:
	for {
		r, _, errRead := lexer.Input.ReadRune()
		switch {
		case errors.Is(errRead, io.EOF):
			return nil, nil
		case errRead != nil:
			return nil, lexer.errPos(errRead)
		}
		lexer.updatePos(r)

		switch {
		case state == stateScan && r == '\\':
			state = stateEscape
			buf.WriteByte('\\')
		case state == stateEscape && strings.ContainsRune(escapable, r):
			state = stateScan
			buf.WriteRune(r)
		case state == stateScan && r == '"':
			buf.WriteRune('"')
			break scan
		case state == stateScan:
			buf.WriteRune(r)
		default:
			return nil, lexer.errPos(errBadStringLit)
		}
	}

	return lexer.newToken(language.TokenStr, buf.String()), nil
}

func isAtomRune(ru rune) bool {
	if unicode.IsSpace(ru) || strings.ContainsRune(brackets, ru) {
		return false
	}

	return unicode.In(ru,
		unicode.Letter,
		unicode.Digit,
		unicode.Mark,
		unicode.Other,
		unicode.Symbol,
		unicode.Punct,
	)
}

// keywords + symbols - strings
// Yes, it's not a canonical atom definition, but it's enough for now.
func (lexer *Lexer) readAtom(r rune) (*language.Token, error) {
	atom, errAtom := readUntil(lexer.Input, isAtomRune, r)
	if errAtom != nil {
		return nil, lexer.errPos(errAtom)
	}

	if value, ok := strings.CutPrefix(atom, ":"); ok {
		return lexer.newToken(language.TokenKeyword, value), nil
	}

	_, errInt := strconv.ParseInt(atom, 10, 64)
	if errInt == nil {
		return lexer.newToken(language.TokenInt, atom), nil
	}

	_, errFloat := strconv.ParseFloat(atom, 64)
	if errFloat == nil {
		return lexer.newToken(language.TokenFloat, atom), nil
	}

	return lexer.newToken(language.TokenSymbol, atom), nil
}

type Error struct {
	Pos language.Position
	Err error
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s: %v", err.Pos, err.Err)
}

func (err *Error) Unwrap() error {
	return err.Err
}

func (lexer *Lexer) errPos(err error) error {
	return &Error{
		Pos: lexer.pos,
		Err: err,
	}
}

func (lexer *Lexer) updatePos(r rune) {
	if lexer.pos.File == "" {
		lexer.pos.File = lexer.File
		lexer.pos.Line = 1
	}
	lexer.pos.Column++
	if r == '\n' {
		lexer.pos.Line++
		lexer.pos.Column = 0
	}
}

func (lexer *Lexer) newToken(kind language.TokenKind, value string) *language.Token {
	return &language.Token{
		Kind:  kind,
		Value: value,
		Pos:   lexer.pos,
	}
}

func readUntil(re io.RuneScanner, fn func(ru rune) bool, prepend ...rune) (string, error) {
	buf := &strings.Builder{}
	for _, ru := range prepend {
		buf.WriteRune(ru)
	}

	for {
		r, _, errRead := re.ReadRune()
		switch {
		case errors.Is(errRead, io.EOF):
			return buf.String(), nil
		case errRead != nil:
			return buf.String(), errRead
		}
		if !fn(r) {
			re.UnreadRune()
			return buf.String(), nil
		}
		buf.WriteRune(r)
	}
}
