package repl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ninedraft/sulisp/interpreter/astwalk"
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

	handle := func(signal Signal) {
		switch signal.Kind {
		case SignalHistoryPrev:
			// pass
		case SignalBackspace:
			fmt.Fprintf(out, "\b \b")
		}
	}

	prompt := func() {
		prompt := ">> "
		if buf.Len() > 0 {
			prompt = ".. "
		}
		_, _ = io.WriteString(out, prompt)
	}

	env := astwalk.DefaultEnv()

	for prompt(); sc.Scan(); prompt() {
		select {
		case signal := <-signals:
			handle(signal)
		default:
			// pass
		}

		cmd := strings.TrimSpace(sc.Text())
		switch cmd {
		case ":q", ":quit":
			fmt.Fprintf(out, "bye!\n")
			return nil
		case ":h", ":help":
			fmt.Fprintf(out, "help\n")
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
			fmt.Fprintf(out, "\n\n%s\n", astwalk.Eval(pkg, env).Inspect())
		}

		buf.Reset()
	}

	return sc.Err()
}

func bufReader(buf *bytes.Buffer) *bytes.Reader {
	return bytes.NewReader(buf.Bytes())
}
