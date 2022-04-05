package reader

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ninedraft/sulisp/ast"
)

func Read(src io.Reader, fname string) (ast.Expr, error) {
	var sc = Scan(src, fname)
	sc.Scan()
	if err := sc.Err(); err != nil {
		return nil, err
	}
	var tok = sc.Token()
	if tok.Kind != ast.TokenLeftParen {
		return parseExpr(tok)
	}
	return readList(sc)
}

func readList(sc *Scanner) (ast.List, error) {
	var list ast.List
scan:
	for sc.Scan() {
		var token = sc.Token()
		switch token.Kind {
		case ast.TokenLeftParen:
			var l, errList = readList(sc)
			if errList != nil {
				return list, errList
			}
			list = append(list, l)
		case ast.TokenRightParen:
			break scan
		default:
			var expr, errExpr = parseExpr(token)
			if errExpr != nil {
				return list, errExpr
			}
			list = append(list, expr)
		}
	}
	return list, sc.Err()
}

var errUnexpectedToken = errors.New("unexpected token")

func parseExpr(token ast.Token) (ast.Expr, error) {
	switch token.Kind {
	case ast.TokenString, ast.TokenAtom:
		return &ast.Literal[string]{Kind: token.Kind, Value: token.Value}, nil
	case ast.TokenInt:
		var x, err = strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", token.Pos, err)
		}
		return &ast.Literal[int64]{Kind: token.Kind, Value: x}, nil
	case ast.TokenFloat:
		var x, err = strconv.ParseFloat(token.Value, 10)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", token.Pos, err)
		}
		return &ast.Literal[float64]{Kind: token.Kind, Value: x}, nil
	}
	return nil, errUnexpectedToken
}
