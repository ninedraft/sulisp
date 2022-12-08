package reader

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"unicode"

	"github.com/ninedraft/sulisp/ast"
	"github.com/ninedraft/sulisp/internal/readstring"
)

func Read(input io.RuneScanner, filename string) ([]ast.Token, error) {
	var tokens []ast.Token
	var line = 1
	var column = 0
	pos := func() *ast.Position {
		return &ast.Position{
			Filename: filename,
			Line:     line,
			Column:   column,
		}
	}
	for {
		ru, _, errRune := input.ReadRune()
		switch {
		case errors.Is(errRune, io.EOF):
			return tokens, nil
		case errRune != nil:
			return tokens, errRune
		}

		if unicode.IsSpace(ru) {
			continue
		}

		token := ast.Token{Pos: pos()}
	match:
		switch ru {
		case '"':
			_ = input.UnreadRune()
			value, errString := readstring.Read(input)
			if errString != nil {
				return tokens, fmt.Errorf("%s: ident: %s", pos(), errString)
			}
			token.Kind = ast.TokenString
			token.Value = value
		case ':':
			_ = input.UnreadRune()
			value, errSymbol := readstring.Symbol(input)
			if errSymbol != nil {
				return tokens, fmt.Errorf("%s: ident: %s", pos(), errSymbol)
			}
			token.Kind = ast.TokenSymbol
			token.Value = value
		case '(', ')', '[', ']', '`', ';', '#':
			token.Value = string([]rune{ru})
			token.Kind = ast.TokenKind(ru)
		case '-', '+', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			_ = input.UnreadRune()
			value, errScan := readstring.Ident(input)
			if errScan != nil {
				return tokens, fmt.Errorf("%s: %v", pos(), errScan)
			}
			if value == "+" || value == "-" {
				token.Kind = ast.TokenIdent
				token.Value = value
				break match
			}
			_, errNumber := strconv.ParseFloat(value, 64)
			if errNumber != nil && !errors.Is(errNumber, strconv.ErrRange) {
				return tokens, fmt.Errorf("%s: bad number litera: %v", pos(), errNumber)
			}
			token.Value = value
			token.Kind = ast.TokenFloat
			if isInteger(value) {
				token.Kind = ast.TokenInt
			}
		default:
			_ = input.UnreadRune()
			value, errIdent := readstring.Ident(input)
			if errIdent != nil {
				return tokens, fmt.Errorf("%s: ident: %s", pos(), errIdent)
			}
			token.Kind = ast.TokenIdent
			token.Value = value
		}

		tokens = append(tokens, token)
		switch ru {
		case '\n':
			line++
			column = 1
		default:
			column++
		}
	}
}

var isInteger = regexp.MustCompile(`^[+-]?[0-9]+$`).MatchString
