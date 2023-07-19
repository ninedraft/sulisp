package parser

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ninedraft/sulisp/language"
)

type Parser struct {
	Tokens []*language.Token
	i      int
}

var ErrUnexpectedToken = errors.New("unexpected token")

// sexp -> '(' exp* ')'
// expr -> sexp | atom
// atom -> number | symbol | string | keyword | boolean
func (parser *Parser) Parse() ([]language.Sexp, error) {
	root := []language.Sexp{}
	for {
		sexp, err := parser.readSexp()
		switch {
		case errors.Is(err, io.EOF):
			return root, nil
		case err != nil:
			return nil, err
		}

		root = append(root, sexp)
	}
}

const quote = language.Symbol("quote")

func (parser *Parser) readExpr() (language.Expression, error) {
	tok, ok := parser.next()
	if !ok {
		return nil, io.EOF
	}
	switch tok.Kind {
	case language.TokenLBrace:
		parser.unread()
		return parser.readSexp()
	case language.TokenInt:
		return parseInt(tok)
	case language.TokenFloat:
		return parseFloat(tok)
	case language.TokenStr:
		return parseString(tok)
	case language.TokenSymbol:
		return language.Symbol(tok.Value), nil
	case language.TokenKeyword:
		return language.Keyword(tok.Value), nil
	case language.TokenQuote:
		return parser.readQuote()
	}
	return nil, fmt.Errorf("%s: %w %s", tok.Pos, ErrUnexpectedToken, tok.Kind)
}

func parseInt(tok *language.Token) (*language.Literal[int64], error) {
	var x int64
	var err error
	switch {
	case strings.HasPrefix(tok.Value, "0x"):
		x, err = strconv.ParseInt(tok.Value, 16, 64)
	case strings.HasPrefix(tok.Value, "0o"):
		x, err = strconv.ParseInt(tok.Value, 8, 64)
	case strings.HasPrefix(tok.Value, "0b"):
		x, err = strconv.ParseInt(tok.Value, 2, 64)
	default:
		x, err = strconv.ParseInt(tok.Value, 10, 64)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", tok.Pos, err)
	}

	return &language.Literal[int64]{Value: x}, nil
}

func parseFloat(tok *language.Token) (*language.Literal[float64], error) {
	x, err := strconv.ParseFloat(tok.Value, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tok.Pos, err)
	}
	return &language.Literal[float64]{Value: x}, nil
}

func parseString(tok *language.Token) (*language.Literal[string], error) {
	val, err := strconv.Unquote(tok.Value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tok.Pos, err)
	}
	return &language.Literal[string]{Value: val}, nil
}

func (parser *Parser) readSexp() (language.Sexp, error) {
	exps := []language.Expression{}
	head, ok := parser.next()
	if !ok {
		return nil, io.EOF
	}

	if head.Kind != language.TokenLBrace {
		return nil, fmt.Errorf("%s: %w %s", head.Pos, ErrUnexpectedToken, head.Kind)
	}

	for {
		tok, ok := parser.next()
		if !ok {
			return nil, io.EOF
		}
		switch tok.Kind {
		case language.TokenRBrace:
			return exps, nil
		default:
			parser.unread()
			exp, err := parser.readExpr()
			if err != nil {
				return nil, errors.Join(ErrUnexpectedToken, err)
			}
			exps = append(exps, exp)
		}
	}
}

func (parser *Parser) readQuote() (language.Sexp, error) {
	expr, errExpr := parser.readExpr()
	if errExpr != nil {
		return nil, errExpr
	}

	return language.Sexp{
		quote,
		expr,
	}, nil
}

func (parser *Parser) unread() {
	parser.i--
}

func (parser *Parser) next() (*language.Token, bool) {
	if parser.i >= len(parser.Tokens) {
		return nil, false
	}
	tok := parser.Tokens[parser.i]
	parser.i++
	return tok, true
}
