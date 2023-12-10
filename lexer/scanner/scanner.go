package scanner

import (
	"io"
)

const EOF rune = 0

type Scanner struct {
	re           io.RuneReader
	line, column int
	current      rune
	next         rune
	err          error
}

func New(re io.RuneReader) *Scanner {
	sc := &Scanner{
		next:    -1,
		current: -1,
		re:      re,
	}

	sc.Scan()

	return sc
}

func (sc *Scanner) Scan() rune {
	if sc.next == EOF && sc.err != nil {
		return EOF
	}

	sc.current = sc.next

	sc.updatePos(sc.current)

	sc.next, _, sc.err = sc.re.ReadRune()

	return sc.current
}

func (sc *Scanner) updatePos(ru rune) {
	sc.column++
	if ru == '\n' {
		sc.line++
		sc.column = 0
	}
}

func (sc *Scanner) Peek() rune {
	return sc.next
}

func (sc *Scanner) Current() rune {
	return sc.current
}

func (sc *Scanner) Err() error {
	return sc.err
}

func (sc *Scanner) Pos() (line, column int) {
	return sc.line, sc.column
}
