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

const integerStart = "0123456789-+"

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
		case r == '\'':
			return lexer.newToken(language.TokenQuote, "'"), nil
		case strings.ContainsRune(brackets, r):
			kind := language.TokenKind(r)
			return lexer.newToken(kind, kind.String()), nil
		case r == '"':
			return lexer.readString()
		case strings.ContainsRune(integerStart, r):
			return lexer.readNumber(r)
		default:
			return lexer.readName(r)
		}
	}
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

// integers + hex + octal + binary
const numberRunes = "+-0123456789_." + "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// emits int, float or name (for +-) tokens
func (lexer *Lexer) readNumber(r rune) (*language.Token, error) {
	buf := &strings.Builder{}
	buf.WriteRune(r)

	// read until non-integer rune
scan:
	for {
		r, _, errRead := lexer.Input.ReadRune()
		switch {
		case errors.Is(errRead, io.EOF):
			break scan
		case errRead != nil:
			return nil, lexer.errPos(errRead)
		}
		lexer.updatePos(r)

		if !strings.ContainsRune(numberRunes, r) {
			lexer.Input.UnreadRune()
			break scan
		}
		buf.WriteRune(r)
	}

	lit := buf.String()

	_, errFloat := strconv.ParseFloat(lit, 64)
	if errFloat == nil {
		return lexer.newToken(language.TokenFloat, lit), nil
	}

	var errInt error

	switch {
	case strings.HasPrefix(lit, "0x"):
		_, errInt = strconv.ParseInt(lit, 16, 64)
	case strings.HasPrefix(lit, "0o"):
		_, errInt = strconv.ParseInt(lit, 8, 64)
	case strings.HasPrefix(lit, "0b"):
		_, errInt = strconv.ParseInt(lit, 2, 64)
	case lit == "+" || lit == "-":
		return lexer.newToken(language.TokenSymbol, lit), nil
	default:
		_, errInt = strconv.ParseInt(lit, 10, 64)
	}

	if errInt != nil {
		return nil, lexer.errPos(errInt)
	}

	return lexer.newToken(language.TokenInt, lit), nil
}

func isNameRune(ru rune) bool {
	return unicode.IsDigit(ru) || unicode.IsLetter(ru) || strings.ContainsRune("+-*/%&|!?:", ru)
}

// keywords + symbols
func (lexer *Lexer) readName(r rune) (*language.Token, error) {
	buf := &strings.Builder{}
	buf.WriteRune(r)

	// read until non-symbol rune
scan:
	for {
		r, _, errRead := lexer.Input.ReadRune()
		switch {
		case errors.Is(errRead, io.EOF):
			break scan
		case errRead != nil:
			return nil, lexer.errPos(errRead)
		}
		lexer.updatePos(r)

		if !isNameRune(r) {
			lexer.Input.UnreadRune()
			break scan
		}
		buf.WriteRune(r)
	}

	kind := language.TokenSymbol
	if r == ':' {
		kind = language.TokenKeyword
	}

	lit := buf.String()

	if lit == "true" || lit == "false" {
		kind = language.TokenBool
	}

	return lexer.newToken(kind, buf.String()), nil
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
