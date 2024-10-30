package core

func Type(value Value) Symbol {
	switch spec := value.Kind().(type) {
	case interface{ Get(Keyword) (Value, bool) }:
		t, ok := spec.Get(":type")
		if !ok {
			return Any
		}
		name, ok := t.(Symbol)
		if !ok {
			return Any
		}
		return name
	case Symbol:
		return spec
	default:
		return Any
	}
}

func newTypeSpec(name string, params HashMap[Keyword, Value]) HashMap[Keyword, Value] {
	return HashMap[Keyword, Value]{
		Keyword(":type"):   Symbol(name),
		Keyword(":params"): params,
	}
}
