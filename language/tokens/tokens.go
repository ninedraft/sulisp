package tokens

import (
	"fmt"
)

type Position struct {
	File   string
	Line   int
	Column int
}

func (pos Position) String() string {
	return fmt.Sprintf("%s:%d:%d", pos.File, pos.Line, pos.Column)
}

type Token struct {
	Kind  TokenKind
	Value string
	Pos   Position
}

func (t *Token) String() string {
	return "<" + t.Kind.String() + " " + t.Value + ">"
}

//go:generate stringer -type TokenKind -linecomment
type TokenKind int

func (kind TokenKind) Of(value string) *Token {
	return &Token{
		Kind:  kind,
		Value: value,
	}
}

const (
	TokenUndefined TokenKind = 0 // <undefined>

	TokenLParen = TokenKind('(') // (
	TokenRParen = TokenKind(')') // )

	TokenLBrack = TokenKind('[') // [
	TokenRBrack = TokenKind(']') // ]

	TokenLBrace = TokenKind('{') // {
	TokenRBrace = TokenKind('}') // }

	TokenQuote = TokenKind('\'') // '

	TokenPoint = TokenKind('.') // .

	TokenSymbol  TokenKind = iota + 100 // symbol
	TokenKeyword                        // :keyword

	TokenInt     // integer
	TokenFloat   // float
	TokenStr     // string
	TokenComment // ; comment

	TokenEOF       // <eof>
	TokenMalformed // malformed
)

func (tk TokenKind) GoString() string {
	switch tk {
	case TokenLParen:
		return "language.TokenLBrace"
	case TokenRParen:
		return "language.TokenRBrace"
	case TokenLBrack:
		return "language.TokenLBrack"
	case TokenRBrack:
		return "language.TokenRBrack"
	case TokenLBrace:
		return "language.TokenLCurl"
	case TokenRBrace:
		return "language.TokenRCurl"
	case TokenQuote:
		return "language.TokenQuote"
	case TokenSymbol:
		return "language.TokenSymbol"
	case TokenKeyword:
		return "language.TokenKeyword"
	case TokenInt:
		return "language.TokenInt"
	case TokenFloat:
		return "language.TokenFloat"
	case TokenStr:
		return "language.TokenStr"
	case TokenComment:
		return "language.TokenComment"
	default:
		return fmt.Sprintf("language.TokenKind(%d)", tk)
	}
}
