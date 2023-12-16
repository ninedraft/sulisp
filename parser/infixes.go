package parser

import (
	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
)

type infixOp func(node ast.Node) ast.Node

func (parser *Parser) addInfix(tok tokens.TokenKind, op infixOp) {
	if op == nil {
		panic("[addInfix]: nil op")
	}

	if parser.infixes == nil {
		parser.infixes = map[tokens.TokenKind]infixOp{}
	}

	parser.infixes[tok] = op
}

func (parser *Parser) parseDotSelector(left ast.Node) ast.Node {
	if !parser.expectCurrentKind(tokens.TokenPoint) {
		return nil
	}

	parser.nextTok()

	right := parser.parseNode()

	if right == nil {
		return nil
	}

	return &ast.DotSelector{
		PosRange: parser.posRange(),
		Left:     left,
		Right:    right,
	}
}
