package readstring

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/exp/maps"
)

// ErrUnexpectedRune means that reader got a strange rune at wrong position
var ErrUnexpectedRune = errors.New("unexpected rune")

func Read(re io.RuneReader) (string, error) {
	start, _, errStart := re.ReadRune()
	if errStart != nil {
		return "", fmt.Errorf("unexpected error: %w", errStart)
	}
	if start != '"' {
		return "", fmt.Errorf("%w %q: \" is expected", ErrUnexpectedRune, start)
	}

	str := &strings.Builder{}
scan:
	for state := stateBody; ; {
		ru, n, errRune := re.ReadRune()
		if errRune != nil {
			return str.String(), fmt.Errorf("unexpected error: %w", errRune)
		}
		if n == 0 {
			return "", fmt.Errorf("%w %q", ErrUnexpectedRune, strconv.QuoteRune(ru))
		}
		switch {
		case ru == unicode.ReplacementChar:
			return str.String(), ErrUnexpectedRune
		case state == stateBody && ru == '"':
			break scan
		case state == stateBody && ru == '\\':
			state = stateEscape
		case state == stateEscape && strings.ContainsRune(escapeRunes, ru):
			str.WriteRune(escapeMappings[ru])
			state = stateBody
		case state == stateEscape:
			return str.String(), fmt.Errorf("%w %q, one of %s is expected", ErrUnexpectedRune, strconv.QuoteRune(ru), escapeRunes)
		case state == stateBody:
			str.WriteRune(ru)
		default:
			panic("unexpected state " + state.String() + "and rune" + strconv.QuoteRune(ru))
		}
	}
	return str.String(), nil
}

var (
	escapeRunes    = string(maps.Keys(escapeMappings))
	escapeMappings = map[rune]rune{
		'r': '\r',
		's': ' ',
		'n': '\n',
		't': '\t',

		'\\': '\\',
		'"':  '"',
	}
)

type readerState int

const (
	stateBody readerState = iota + 1
	stateEscape
)

func (state readerState) String() string {
	switch state {
	case stateBody:
		return "body"
	case stateEscape:
		return "escape"
	default:
		return fmt.Sprintf("unexpected state value %d", state)
	}
}
