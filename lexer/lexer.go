package lexer

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/ninedraft/sulisp/language"
	scanner "github.com/ninedraft/sulisp/lexer/scanner"
)

const eof = scanner.EOF

type Lexer struct {
	file    string
	scanner *scanner.Scanner
}

func NewLexer(filename string, re io.RuneReader) *Lexer {
	sc := scanner.New(re)
	sc.Scan()
	return &Lexer{
		file:    filename,
		scanner: sc,
	}
}

func (lexer *Lexer) Next() (*language.Token, error) {
	tok, err := lexer.next()
	if err != nil {
		return nil, lexer.errPos(err)
	}

	return tok, nil
}

const brackets = "[](){}"

func (lexer *Lexer) next() (*language.Token, error) {
	ru := lexer.scanner.Current()

	// skipping spaces and commas
	for unicode.IsSpace(ru) || ru == ',' {
		ru = lexer.scanner.Scan()
		if ru == eof {
			return nil, lexer.scanner.Err()
		}
	}

	switch {
	case ru == ';':
		return lexer.readComment()
	case ru == '.':
		return lexer.newToken(language.TokenPoint, "."), nil
	case ru == '\'':
		return lexer.newToken(language.TokenQuote, "'"), nil
	case containsRune(brackets, ru):
		kind := language.TokenKind(ru)
		return lexer.newToken(kind, kind.String()), nil
	case ru == '"':
		return lexer.readString()
	case unicode.IsDigit(ru) || ru == '+' || ru == '-':
		return lexer.readNumber()
	case ru == ':':
		return lexer.readAtom(language.TokenKeyword)
	default:
		return lexer.readAtom(language.TokenSymbol)
	}
}

func (lexer *Lexer) readComment() (*language.Token, error) {
	comment := &strings.Builder{}
	sc := lexer.scanner

	for current := sc.Current(); ; current = sc.Scan() {
		if current == eof {
			break
		}

		comment.WriteRune(current)

		if current == '\n' {
			break
		}
	}

	return &language.Token{
		Kind:  language.TokenComment,
		Value: comment.String(),
	}, nil
}

var errBadStringLit = errors.New("bad string literal")

var strEscapes = map[[2]rune]rune{
	{'\\', '\\'}: '\\',
	{'\\', '"'}:  '"',
	{'\\', 'r'}:  'r',
	{'\\', 'n'}:  'n',
	{'\\', 't'}:  't',
}

func (lexer *Lexer) readString() (*language.Token, error) {
	buf := &strings.Builder{}
	sc := lexer.scanner

	// already know that first rune is '"'
	buf.WriteRune('"')

scan:
	for current := sc.Scan(); ; current = sc.Scan() {
		escaped, isEscape := strEscapes[[2]rune{current, sc.Peek()}]
		switch {
		case current == eof:
			return nil, errors.Join(errBadStringLit, sc.Err())
		case current == '"':
			buf.WriteRune(current)
			sc.Scan()
			break scan
		case isEscape:
			buf.WriteRune('\\')
			buf.WriteRune(escaped)
			sc.Scan()
		case current == '\\':
			return nil, fmt.Errorf("%w: unexpected escaped symbol %q", errBadStringLit, sc.Peek())
		case current == '\n':
			buf.WriteString("\\n")
		default:
			buf.WriteRune(current)
		}
	}

	return lexer.newToken(language.TokenStr, buf.String()), sc.Err()
}

// can read number or symbols + -
func (lexer *Lexer) readNumber() (*language.Token, error) {
	value := &strings.Builder{}
	sc := lexer.scanner

	kind := language.TokenInt

	for current := sc.Current(); ; current = sc.Scan() {
		if containsRune(",.eE", current) {
			kind = language.TokenFloat
		}

		ok := unicode.IsDigit(current) || containsRune("+-_box,.eE", current)
		if !ok {
			break
		}

		value.WriteRune(current)
	}

	v := value.String()

	if v == "+" || v == "-" {
		kind = language.TokenSymbol
	}

	return &language.Token{
		Kind:  kind,
		Value: v,
	}, nil
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
		Pos: lexer.pos(),
		Err: err,
	}
}

func (lexer *Lexer) pos() language.Position {
	line, column := lexer.scanner.Pos()
	return language.Position{
		Line:   line,
		Column: column,
		File:   lexer.file,
	}
}

func (lexer *Lexer) newToken(kind language.TokenKind, value string) *language.Token {
	return &language.Token{
		Kind:  kind,
		Value: value,
		Pos:   lexer.pos(),
	}
}

// read

// keywords + symbols - strings
func (lexer *Lexer) readAtom(kind language.TokenKind) (*language.Token, error) {
	value := &strings.Builder{}
	sc := lexer.scanner
	value.WriteRune(sc.Current())

	for {
		ru := sc.Scan()
		if ru == eof {
			return nil, sc.Err()
		}

		if !isAtomRune(ru) {
			break
		}

		value.WriteRune(ru)
	}

	return &language.Token{
		Kind:  kind,
		Value: value.String(),
	}, sc.Err()
}

func isAtomRune(ru rune) bool {
	if unicode.IsSpace(ru) || containsRune(brackets, ru) || ru == '.' || ru == ',' {
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

func containsRune(rr string, ru rune) bool {
	return strings.ContainsRune(rr, ru)
}
