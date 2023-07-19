package language

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
	return t.Kind.String() + " " + t.Value
}

//go:generate stringer -type TokenKind -linecomment
type TokenKind int

const (
	TokenLBrace = TokenKind('(') // (
	TokenRBrace = TokenKind(')') // )

	TokenLBrack = TokenKind('[') // [
	TokenRBrack = TokenKind(']') // ]

	TokenLCurl = TokenKind('{') // {
	TokenRCurl = TokenKind('}') // }

	TokenQuote = TokenKind('\'') // '

	TokenSymbol  TokenKind = iota + 100 // symbol
	TokenKeyword                        // :keyword

	TokenInt   // integer
	TokenFloat // float
	TokenStr   // string
)

func (tk TokenKind) GoString() string {
	switch tk {
	case TokenLBrace:
		return "language.TokenLBrace"
	case TokenRBrace:
		return "language.TokenRBrace"
	case TokenLBrack:
		return "language.TokenLBrack"
	case TokenRBrack:
		return "language.TokenRBrack"
	case TokenLCurl:
		return "language.TokenLCurl"
	case TokenRCurl:
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
	default:
		return fmt.Sprintf("language.TokenKind(%d)", tk)
	}
}
