// Package parser is Pratt top-down parser for a lispy language.
package parser

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"

	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
)

type Parser struct {
	lexer Lexer

	errs      []error
	cur, next *tokens.Token
}

func New(lexer Lexer) *Parser {
	parser := &Parser{
		lexer: lexer,
	}

	parser.nextTok()
	parser.nextTok()

	return parser
}

type Lexer interface {
	Next() (*tokens.Token, error)
}

func (parser *Parser) Parse() (*ast.Package, error) {
	pkg := &ast.Package{}

	for len(parser.errs) == 0 && !parser.curIs(tokens.TokenEOF) {
		item := parser.parseNode()
		if item == nil {
			continue
		}

		pkg.Nodes = append(pkg.Nodes, item)

		parser.nextTok()
	}

	return pkg, errors.Join(parser.errs...)
}

func (parser *Parser) parseNode() ast.Node {
	switch parser.cur.Kind {
	case tokens.TokenLParen:
		return parser.parseApply()
	case tokens.TokenSymbol, tokens.TokenKeyword, tokens.TokenPoint:
		return parser.parseAtomBoolOrDot()
	case tokens.TokenInt, tokens.TokenFloat, tokens.TokenStr: // bool parsed in parseAtomBoolOrDot
		return parser.parseLiteral()
	default:
		parser.errorf("unexpected token: %s", parser.cur)
		return nil
	}
}

var defaultAtoms = []tokens.TokenKind{
	tokens.TokenSymbol,
	tokens.TokenKeyword,
	tokens.TokenPoint,
}

// parse symbol, keyword, bool literal or dot
func (parser *Parser) parseAtomBoolOrDot(expect ...tokens.TokenKind) ast.Node {
	if len(expect) == 0 {
		expect = defaultAtoms
	}

	if !parser.expectCurrentKind(expect...) {
		return nil
	}

	// crutch for bool literals
	if parser.cur.Value == "true" || parser.cur.Value == "false" {
		b, _ := strconv.ParseBool(parser.cur.Value)
		return &ast.Literal[bool]{PosRange: parser.posRange(), Value: b}
	}

	var node ast.Node

	switch parser.cur.Kind {
	case tokens.TokenSymbol, tokens.TokenPoint:
		node = &ast.Symbol{PosRange: parser.posRange(), Value: parser.cur.Value}
	case tokens.TokenKeyword:
		node = &ast.Keyword{PosRange: parser.posRange(), Value: parser.cur.Value}
	default:
		parser.errorf("unexpected token: %s", parser.cur)
		return nil
	}

	return node
}

func (parser *Parser) parseLiteral() ast.Node {
	if !parser.expectCurrentKind(tokens.TokenInt, tokens.TokenFloat, tokens.TokenStr) {
		return nil
	}

	var parsed ast.Node
	pos := parser.posRange()
	var errParse error

	value := parser.cur.Value
	switch parser.cur.Kind {
	case tokens.TokenInt:
		x, err := strconv.ParseInt(value, 0, 64)
		errParse = err
		parsed = &ast.Literal[int64]{PosRange: pos, Value: x}

	case tokens.TokenFloat:
		x, err := strconv.ParseFloat(value, 64)
		errParse = err
		parsed = &ast.Literal[float64]{PosRange: pos, Value: x}

	case tokens.TokenStr:
		parsed = &ast.Literal[string]{PosRange: pos, Value: value}
	}

	if errParse != nil {
		parser.errorf("cannot parse literal %s: %s", parser.cur, errParse)
		return nil
	}

	return parsed
}

// can return special forms
func (parser *Parser) parseApply() ast.Node {
	sexp := parser.parseSexp()
	if sexp == nil || len(sexp.Items) == 0 {
		return sexp
	}

	head := sexp.Items[0]

	symbol, _ := head.(*ast.Symbol)
	if symbol == nil {
		// not a special form
		return sexp
	}

	if isSpecial[symbol.Value] {
		return parser.buildSpecial(sexp)
	}

	return sexp
}

func (parser *Parser) parseSexp() *ast.SExp {
	if !parser.expectCurrentKind(tokens.TokenLParen) {
		return nil
	}

	parser.nextTok()

	sexp := &ast.SExp{}

	for !parser.curIs(tokens.TokenRParen, tokens.TokenEOF) {
		node := parser.parseNode()
		if node != nil {
			sexp.Items = append(sexp.Items, node)
		}
		parser.nextTok()
	}

	if !parser.expectCurrentKind(tokens.TokenRParen) {
		return nil
	}

	return sexp
}
func (parser *Parser) curIs(kinds ...tokens.TokenKind) bool {
	return parser.cur != nil && slices.Contains(kinds, parser.cur.Kind)
}

func (parser *Parser) expectCurrentKind(kinds ...tokens.TokenKind) bool {
	if parser.cur == nil {
		parser.errorf("no current token")
		return false
	}

	ok := slices.Contains(kinds, parser.cur.Kind)
	if !ok {
		parser.errorf("current: want tokens %s, got %q", kinds, parser.cur)
	}

	return ok
}

func (parser *Parser) expectNextKind(kinds ...tokens.TokenKind) bool {
	if parser.next == nil {
		parser.errorf("no next token")
		return false
	}

	ok := slices.Contains(kinds, parser.next.Kind)
	if !ok {
		parser.errorf("next: want tokens %s, got %s", kinds, parser.next.Kind)
	}

	return ok
}

type Error struct {
	Pos tokens.Position
	Err error
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s: %s", err.Pos, err.Err)
}

func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func (parser *Parser) errorf(msg string, args ...any) {
	err := fmt.Errorf(msg, args...)

	if parser.cur != nil {
		err = &Error{Pos: parser.cur.Pos, Err: err}
	}

	parser.errs = append(parser.errs, err)
}

func (parser *Parser) nextTok() {
	parser.cur = parser.next
	next, err := parser.lexer.Next()

	switch {
	case errors.Is(err, io.EOF):
		next = &tokens.Token{Kind: tokens.TokenEOF, Pos: parser.cur.Pos}
	case err != nil:
		parser.errs = append(parser.errs, err)
	}

	parser.next = next
}

func (parser *Parser) posRange() ast.PosRange {
	pos := ast.PosRange{}

	if parser.cur != nil {
		pos.From = parser.cur.Pos
	}

	if parser.next != nil {
		pos.To = parser.next.Pos
	}

	return pos
}
