package ast

import "strings"

type List []Expr

func (List) IsExpr() {}

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
		str.WriteString(elem.String())
		if i > 0 {
			str.WriteByte(' ')
		}
	}
	str.WriteRune(')')
}
