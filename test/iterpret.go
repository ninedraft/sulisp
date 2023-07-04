package test

import . "github.com/ninedraft/sulisp/std/core"

func Values(values ...Value) *List[Value] {
	return ListNew[Value](values...)
}

func Symbols(symbols ...Symbol) *List[Symbol] {
	return ListNew[Symbol](symbols...)
}

func _() {
	/*
		(fn inc (x / Int) (=> Int)
			(+ x 1))
	*/
	fn := ListNew[Value](
		Symbol("fn"),
		Symbol("inc"),
		Symbols("x", "/", "Int"),
		Symbols("=>", "Int"),
		Values(Symbol("+"), Symbol("x"), Int(1)),
	)
	_ = fn
}

// match a sequence of values against a pattern.
// Pattern is a sequence of values, where:
// - Symbol("_") is a placeholder for any value
// - Symbol("/") followed by a type is a placeholder for a value of that type
//
// Example in sulisp pseudocode:
//
//	(match (list 1 2) (list _ 2)) 		 ; true
//	(match (list 1 2) (list 1 _))			 ; true
//	(match (list "string" 2) (list _ \ string _)) ; true
func match[E Value](seq Seq[E], pattern Seq[Value]) bool {
	const placeholder = Symbol("_")
	const typePrefix = Symbol("/")

	if seq.Empty() {
		return pattern.Empty()
	}
	if pattern.Empty() {
		return false
	}

	patternHead, _ := pattern.First()
	patternTail := pattern.Next()

	switch {
	case Eq[Value](patternHead, placeholder):
		return match(seq.Next(), patternTail)
	case Eq[Value](patternHead, typePrefix):
		if patternTail.Empty() {
			return false
		}
		patternType, _ := patternTail.First()
		patternTail = patternTail.Next()
		valueHead, _ := seq.First()
		return Eq(valueHead.Type(), patternType) && match(seq.Next(), patternTail)
	default:
		valueHead, _ := seq.First()
		return Eq(patternHead, Value(valueHead)) && match(seq.Next(), patternTail)
	}
}
