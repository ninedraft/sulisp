package core

import (
	"bytes"
	"errors"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Symbol is a string that resolves to a value.
// It can be used as a map key if quoted.
type Symbol string

func (symbol Symbol) Kind() Value {
	return newTypeSpec("core.Symbol", map[Keyword]Value{
		":name": symbol,
	})
}

func (symbol Symbol) Eq(other Symbol) bool {
	return symbol == other
}

func (symbol Symbol) Name() string {
	return string(symbol)
}

func (symbol Symbol) GoString() string {
	q := strconv.Quote(symbol.String())
	return "Symbol(" + q + ")"
}

func (symbol Symbol) String() string {
	return string(symbol)
}

func (symbol Symbol) MarshalText() ([]byte, error) {
	return []byte(symbol), nil
}

var (
	errSymbol = errors.New("malformed symbol")
)

func (symbol *Symbol) UnmarshalText(bb []byte) error {
	if isIdent(bb) {
		return errSymbol
	}

	*symbol = Symbol(bb)

	return nil
}

var identDot = []byte(".")

// Valid identifiers:
//
//	foo
//	foo.bar
//	.foo // method call
func isIdent(name []byte) bool {
	if len(name) == 0 {
		return false
	}

	// special case for method calls
	name = bytes.TrimPrefix(name, identDot)

	for len(name) != 0 {
		field, rest, _ := bytes.Cut(name, identDot)
		if !isGoIdent(field) {
			return false
		}
		name = rest
	}

	return true
}

func isGoIdent(name []byte) bool {
	if len(name) == 0 {
		return false
	}

	for i := 0; len(name) > 0; i++ {
		ru, n := utf8.DecodeRune(name)
		if !unicode.IsLetter(ru) && ru != '_' && (i == 0 || !unicode.IsDigit(ru)) {
			return false
		}
		name = name[n:]
	}

	return true
}
