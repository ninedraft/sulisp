package scanner

import (
	"errors"
	"io"
)

const EOF rune = 0

type Scanner struct {
	re            io.RuneReader
	line, column  int
	current, next rune
	err           error
}

func New(re io.RuneReader) *Scanner {
	sc := &Scanner{
		next:    -1,
		current: -1,
		re:      re,
	}

	current, _, errCurrent := sc.re.ReadRune()
	next, _, errNext := sc.re.ReadRune()

	err := errors.Join(errCurrent, errNext)
	if err != nil {
		sc.err = err
		sc.next, sc.current = EOF, EOF
	}

	sc.current = current
	sc.next = next

	return sc
}

func (sc *Scanner) Scan() rune {
	if sc.next == EOF {
		sc.current = sc.next
		return sc.next
	}

	sc.current = sc.next

	sc.updatePos(sc.current)

	next, _, err := sc.re.ReadRune()

	if err != nil && !errors.Is(err, io.EOF) {
		sc.err = err
		sc.next, sc.current = EOF, EOF
		return EOF
	}

	sc.next = next

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
