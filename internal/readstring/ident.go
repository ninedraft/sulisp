package readstring

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func Ident(re io.RuneScanner) (string, error) {
	ident := &strings.Builder{}

	head, _, errHead := re.ReadRune()
	switch {
	case errors.Is(io.EOF, errHead):
		return ident.String(), nil
	case errHead != nil:
		return ident.String(), errHead
	}

	if isSpecial(head) {
		return "", fmt.Errorf("%w %q", errUnexpected, ru2str(head))
	}

	_ = re.UnreadRune()
	errBody := readBody(re, ident)
	switch {
	case errors.Is(io.EOF, errBody):
		return ident.String(), nil
	case errBody != nil:
		return ident.String(), errBody
	}

	return ident.String(), nil
}
