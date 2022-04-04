package multierr

import "strings"

func Combine(a, b error) error {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	return &pair{a, b}
}

type pair [2]error

func (p *pair) Error() string {
	str := &strings.Builder{}
	str.WriteString(p[0].Error())
	str.WriteString("; ")
	str.WriteString(p[1].Error())
	return str.String()
}
