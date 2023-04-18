package core

type Collection[E Value] interface {
	Len() int
	Seq() Seq[E]
}

type Seq[E Value] interface {
	Empty() bool
	First() (E, bool)
	Next() Seq[E]
}
