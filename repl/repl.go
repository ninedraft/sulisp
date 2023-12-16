package repl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ninedraft/sulisp/lexer"
	"github.com/ninedraft/sulisp/parser"
)

type Signal struct {
	Kind SignalKind
}

type SignalKind int

const (
	SignalHistoryPrev SignalKind = iota + 1
	SignalBackspace              // delete last char
)

func Run(out io.Writer, in io.Reader, signals <-chan Signal) error {
	buf := &bytes.Buffer{}

	sc := bufio.NewScanner(in)

	fmt.Fprintf(out, ">> ")

	handle := func(signal Signal) {
		switch signal.Kind {
		case SignalHistoryPrev:
			// pass
		case SignalBackspace:
			fmt.Fprintf(out, "\b \b")
		}
	}

	for sc.Scan() {
		select {
		case signal := <-signals:
			handle(signal)
		default:
			// pass
		}

		if len(bytes.TrimSpace(sc.Bytes())) == 0 {
			fmt.Fprintf(out, ">> ")
			continue
		}

		buf.Write(sc.Bytes())

		lex := lexer.NewLexer("repl", bufReader(buf))
		par := parser.New(lex)

		pkg, errParse := par.Parse()

		switch {
		case errors.Is(errParse, io.EOF),
			errors.Is(errParse, io.ErrUnexpectedEOF):
			buf.WriteRune('\n')

			continue

		case errParse != nil:
			fmt.Fprintf(out, "ERROR:\n%s\n", errParse)
		default:
			fmt.Fprintf(out, "\n\n%s\n", pkg)
		}

		buf.Reset()
		fmt.Fprintf(out, ">> ")
	}

	return sc.Err()
}

func bufReader(buf *bytes.Buffer) *bytes.Reader {
	return bytes.NewReader(buf.Bytes())
}
