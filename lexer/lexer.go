package lexer

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	language "github.com/ninedraft/sulisp/language/tokens"
	scanner "github.com/ninedraft/sulisp/lexer/scanner"
)

const eof = scanner.EOF

type Lexer struct {
	file    string
	scanner *scanner.Scanner
}

func NewLexer(filename string, re io.RuneReader) *Lexer {
	sc := scanner.New(re)
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

func (lexer *Lexer) next() (scanned *language.Token, _ error) {
	ru := lexer.scanner.Current()

	// skipping spaces and commas
	for unicode.IsSpace(ru) || ru == ',' {
		ru = lexer.scanner.Scan()
	}

	switch {
	case ru == eof:
		tok := lexer.newToken(language.TokenEOF, "")
		lexer.scanner.Scan()
		return tok, lexer.scanner.Err()
	case ru == ';':
		return lexer.readComment()
	case ru == '.':
		tok := lexer.newToken(language.TokenPoint, ".")
		lexer.scanner.Scan()
		return tok, nil
	case ru == '\'':
		tok := lexer.newToken(language.TokenQuote, "'")
		lexer.scanner.Scan()
		return tok, nil
	case containsRune(brackets, ru):
		kind := language.TokenKind(ru)
		tok := lexer.newToken(kind, kind.String())
		lexer.scanner.Scan()
		return tok, nil
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
		if strings.ContainsRune("\x00\n", current) {
			break
		}

		comment.WriteRune(current)
	}

	return lexer.newToken(language.TokenComment, comment.String()), nil
}

var errBadStringLit = errors.New("bad string literal")

var strEscapes = map[[2]rune]rune{
	{'\\', '\\'}: '\\',
	{'\\', '"'}:  '"',
	{'\\', 'r'}:  'r',
	{'\\', 'n'}:  'n',
	{'\\', 't'}:  't',
	{'\\', 'x'}:  'x',
}

func (lexer *Lexer) readString() (*language.Token, error) {
	buf := &strings.Builder{}
	sc := lexer.scanner

	// already know that first rune is '"'
	buf.WriteRune('"')

scan:
	for current := sc.Scan(); ; current = sc.Scan() {
		window := [...]rune{current, sc.Peek()}
		escaped, isEscape := strEscapes[window]

		switch {
		case current == eof:
			return nil, errors.Join(errBadStringLit, sc.Err(), io.ErrUnexpectedEOF)
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
	const floaty = ".eE"
	const numbery = "+-_box" + floaty
	value := &strings.Builder{}
	sc := lexer.scanner

	kind := language.TokenInt

	for current := sc.Current(); ; current = sc.Scan() {
		if containsRune(".eE", current) {
			kind = language.TokenFloat
		}

		ok := unicode.IsDigit(current) || containsRune(numbery, current)
		if !ok {
			break
		}

		value.WriteRune(current)
	}

	v := value.String()

	if v == "+" || v == "-" {
		kind = language.TokenSymbol
	}

	return lexer.newToken(kind, v), nil
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

	return lexer.newToken(kind, value.String()), sc.Err()
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
