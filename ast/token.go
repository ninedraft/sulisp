package ast

import "text/scanner"

//go:generate stringer -type=TokenKind -linecomment
type TokenKind rune

const (
	TokenAtom   TokenKind = iota + 1 // atom
	TokenInt                         // int
	TokenFloat                       // float
	TokenString                      // string
	TokenChar                        // char

	TokenLeftParen  = TokenKind('(') // (
	TokenRightParen = TokenKind(')') // )
)

type Pos = scanner.Position

type Token struct {
	Pos   Pos
	Kind  TokenKind
	Value string
}

func NewToken(kind TokenKind, value string) Token {
	return Token{
		Kind:  kind,
		Value: value,
	}
}
