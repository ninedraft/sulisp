package ast

//go:generate stringer -type=TokenKind -linecomment
type TokenKind rune

const (
	TokenAtom   TokenKind = iota + 1 // atom
	TokenInt                         // int
	TokenFloat                       // float
	TokenString                      // string
	TokenChar                        // char

	TokenLeftRB  = TokenKind('(') // (
	TokenRightRB = TokenKind(')') // )
)

type Token struct {
	Kind  TokenKind
	Value string
}

func NewToken(kind TokenKind, value string) Token {
	return Token{
		Kind:  kind,
		Value: value,
	}
}

func (Token) IsListElem() {}
