package ast

type Token struct {
	Kind  TokenKind
	Flags TokenFlags
	Value string
}

func (token Token) IsEnd() bool { return token.Kind == TokenEnd }

//go:generate stringer -type=TokenKind -linecomment
type TokenKind int

const (
	TokenEnd      TokenKind = iota
	TokenLeftPar            // (
	TokenRightPar           // )
	TokenString             // string
	TokenAtom               // atom
)

type TokenFlags uint64

const (
	FInt TokenFlags = 1 << iota
	FFloat
	FSymbol
)
