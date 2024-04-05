package ast

import (
	"iter"

	"github.com/ninedraft/itermore"
)

// Walk traverses the AST in depth-first order.
// It emits each node and its context to the yield function starting from leaf nodes.
// It calls next for contexts starting from the root node - so it basically allows to traverse the tree in both directions.
func Walk[Ctx any](node Node, ctx Ctx, next func(Ctx, Node) Ctx) iter.Seq2[Node, Ctx] {
	return func(yield func(Node, Ctx) bool) {
		for child := range children(node) {
			ok := itermore.YieldFrom2(yield, Walk(child, next(ctx, child), next))
			if !ok {
				return
			}
		}

		yield(node, ctx)
	}
}

func children(node Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		switch n := node.(type) {
		case *Package:
			itermore.YieldFrom(yield, itermore.Slice(n.Nodes))
		case *ImportGo:
			itermore.YieldFrom(yield, itermore.Slice(n.Items))
		case *SExp:
			itermore.YieldFrom(yield, itermore.Slice(n.Items))
		case *FunctionDecl:
			yieldFunctionLiteral(yield, n)
		case *FunctionSpec:
			yieldFunctionSpec(yield, n)
		case *DotSelector:
			yieldDotSelector(yield, n)
		case *If:
			if yield(n.Cond) && yield(n.Then) && n.Else != nil {
				yield(n.Else)
			}
		case *SpecialOp:
			itermore.YieldFrom(yield, itermore.Slice(n.Items))
		}
	}
}

func yieldFunctionLiteral(yield func(Node) bool, fl *FunctionDecl) bool {
	return itermore.YieldFrom(yield, itermore.Items[Node](
		fl.Spec,
		fl.Body,
	))
}

func yieldFunctionSpec(yield func(Node) bool, fs *FunctionSpec) bool {
	for _, param := range fs.Params {
		if !yield(param.Type) {
			return false
		}
	}

	return itermore.YieldFrom(yield, itermore.Slice(fs.Ret))
}

func yieldDotSelector(yield func(Node) bool, ds *DotSelector) bool {
	return itermore.YieldFrom(yield, itermore.Items(
		ds.Left,
		ds.Right,
	))
}
