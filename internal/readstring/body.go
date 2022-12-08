package readstring

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

var errUnexpected = errors.New("unexpected rune")

func isSpecial(ru rune) bool {
	return strings.ContainsRune("(){}`#;:", ru)
}

type runeWriter interface {
	WriteRune(ru rune) (int, error)
}

func readBody(re io.RuneScanner, dst runeWriter) error {
	for {
		ru, _, errRead := re.ReadRune()
		if errRead != nil {
			return errRead
		}
		if unicode.IsSpace(ru) {
			break
		}
		if isSpecial(ru) {
			_ = re.UnreadRune()
			break
		}
		if _, err := dst.WriteRune(ru); err != nil {
			return err
		}
	}
	return nil
}

func ru2str(ru rune) string {
	return string([]rune{ru})
}
