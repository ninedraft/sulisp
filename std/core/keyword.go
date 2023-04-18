package core

import (
	"errors"
	"regexp"
	"strconv"
)

// Keyword is a string that starts with a colon.
// Examples:
//
//	:foo
//	:bar?
//	:baz_123
//
// It resolves to itself, so you can use it as a map key.
type Keyword string

func (keyword Keyword) Type() Value {
	return newTypeSpec("core.KeyWord", map[Keyword]Value{
		":name": keyword,
	})
}

func (keyword Keyword) Eq(other Keyword) bool {
	return keyword == other
}

func (keyword Keyword) Name() string {
	return string(keyword)[1:]
}

func (keyword Keyword) GoString() string {
	q := strconv.Quote(keyword.String())
	return "Keyword(" + q + ")"
}

func (keyword Keyword) String() string {
	return string(keyword)
}

func (keyword Keyword) MarshalText() ([]byte, error) {
	return []byte(keyword), nil
}

var (
	errKeywordMalformed = errors.New("malformed keyword")
	keywordRe           = regexp.MustCompile(`^:[^\s]+$`)
)

func (keyword *Keyword) UnmarshalText(bb []byte) error {
	if !keywordRe.Match(bb) {
		return errKeywordMalformed
	}

	*keyword = Keyword(bb)

	return nil
}
