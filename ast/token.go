package ast

import "fmt"

type Token struct {
	Kind  TokenKind
	Flags TokenFlags
	Value string
	Pos   *Position
}

type Position struct {
	Filename string
	Line     int
	Column   int
}

func (pos *Position) String() string {
	if pos == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
}

func (token Token) IsEnd() bool { return token.Kind == TokenEnd }

func (token Token) String() string {
	return token.Kind.String() + " " + token.Value
}

//go:generate stringer -type=TokenKind -linecomment
type TokenKind int

const (
	TokenEnd    TokenKind = iota
	TokenString           // string
	TokenSymbol           // :symbol
	TokenIdent
	TokenInt   // int
	TokenFloat // float

	TokenComment     TokenKind = ';' // ;
	TokenLambda      TokenKind = '#' // #
	TokenQuote       TokenKind = '`' // `
	TokenLeftPar     TokenKind = '(' // (
	TokenRightPar    TokenKind = ')' // )
	TokenLeftSquare  TokenKind = '[' // [
	TokenRightSquare TokenKind = ']' // ]
)

type TokenFlags uint64
