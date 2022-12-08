package readstring

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func Symbol(re io.RuneScanner) (string, error) {
	symbol := &strings.Builder{}

	head, _, errHead := re.ReadRune()
	switch {
	case errors.Is(io.EOF, errHead):
		return symbol.String(), nil
	case errHead != nil:
		return symbol.String(), errHead
	}

	if head != ':' {
		return "", fmt.Errorf("%w %q", errUnexpected, ru2str(head))
	}

	_ = re.UnreadRune()
	errBody := readBody(re, symbol)
	switch {
	case errors.Is(io.EOF, errBody):
		return symbol.String(), nil
	case errBody != nil:
		return symbol.String(), errBody
	}

	return symbol.String(), nil
}
