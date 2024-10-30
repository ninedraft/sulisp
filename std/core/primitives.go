package core

import (
	"encoding"
	"strconv"
)

const Any = Symbol("core.Any")

type Value interface {
	String() string
	Kind() Value

	encoding.TextMarshaler
}

type ValueEq[E Value] interface {
	Value
	Eq(E) bool
}

type Int int

func (Int) Kind() Value {
	return newTypeSpec("core.Int", map[Keyword]Value{})
}

func (i Int) As() int {
	return int(i)
}

func (i Int) Eq(other Int) bool {
	return i == other
}

func (i Int) String() string {
	return strconv.Itoa(int(i))
}

func (i Int) GoString() string {
	return "Int(" + strconv.Itoa(int(i)) + ")"
}

func (i Int) MarshalText() ([]byte, error) {
	return []byte(strconv.Itoa(int(i))), nil
}

func (i *Int) UnmarshalText(data []byte) error {
	v, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}

	*i = Int(v)
	return nil
}

type Float float64

func (Float) Kind() Value {
	return newTypeSpec("core.Float", map[Keyword]Value{})
}

func (f Float) As() float64 {
	return float64(f)
}

func (f Float) Eq(other Float) bool {
	return f == other
}

func (f Float) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 64)), nil
}

func (f *Float) UnmarshalText(data []byte) error {
	v, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}

	*f = Float(v)
	return nil
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) GoString() string {
	return "Float(" + f.String() + ")"
}

type String string

func (String) Kind() Value {
	return newTypeSpec("core.String", map[Keyword]Value{})
}

func (s String) As() string {
	return string(s)
}

func (s String) Eq(other String) bool {
	return s == other
}

func (s String) String() string {
	return string(s)
}

func (s String) GoString() string {
	return "String(" + strconv.Quote(s.String()) + ")"
}

func (s String) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

func (s *String) UnmarshalText(data []byte) error {
	*s = String(data)
	return nil
}

type Bool bool

const (
	True  Bool = true
	False Bool = false
)

func (b Bool) Kind() Value {
	return newTypeSpec("core.Bool", map[Keyword]Value{})
}

func (b Bool) As() bool {
	return bool(b)
}

func (b Bool) Eq(other Bool) bool {
	return b == other
}

func (b Bool) String() string {
	return strconv.FormatBool(bool(b))
}

func (b Bool) GoString() string {
	return "Bool(" + b.String() + ")"
}

func (b Bool) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *Bool) UnmarshalText(data []byte) error {
	v, err := strconv.ParseBool(string(data))
	if err != nil {
		return err
	}

	*b = Bool(v)
	return nil
}
