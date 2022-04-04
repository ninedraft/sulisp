package ast

import "strings"

type ListElem interface {
	IsListElem()
}

type List []ListElem

func (List) IsListElem() {}

func (list *List) AppendToken(kind TokenKind, value string) {
	*list = append(*list, NewToken(kind, value))
}

func (list *List) AppendList(l List) {
	*list = append(*list, l)
}

func (list List) String() string {
	if len(list) == 0 {
		return "()"
	}
	str := &strings.Builder{}
	list.writeStr(str)
	return str.String()
}

func (list List) writeStr(str *strings.Builder) {
	str.WriteByte('(')
	for i, elem := range list {
		switch elem := elem.(type) {
		case List:
			elem.writeStr(str)
		case Token:
			str.WriteString(elem.Value)
		}
		if i > 0 {
			str.WriteByte(' ')
		}
	}
	str.WriteRune(')')
}
