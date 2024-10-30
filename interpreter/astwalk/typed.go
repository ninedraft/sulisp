package astwalk

import (
	"errors"
	"fmt"

	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/object"
)

const (
	TypeAny    = object.ObjAny
	TypeInt    = object.ObjInteger
	TypeFloat  = object.ObjFloat64
	TypeString = object.ObjString
	TypeBool   = object.ObjBool
	TypeNull   = object.ObjNull
	TypeArray  = object.ObjArray
	TypeError  = object.ObjError
)

var errType = errors.New("type inference error")

// Example usage in the Eval function, or similar functions for evaluation
func EvalWithInference(node ast.Node, env *object.Env) object.Object {
	inferencer := NewTypeInferencer(env)
	_, err := inferencer.Infer(node)
	if err != nil {
		return fmtError(node.Pos(), "%w", err)
	}

	// Continue with evaluation if type inference succeeds
	return Eval(node, env)
}

func Infer(sexp *ast.SExp, env *object.Env, eval object.Eval) object.Object {
	if len(sexp.Items) == 0 {
		return fmtError(sexp.Pos(), "need at least on expression to infer types, got none")
	}

	ot, err := NewTypeInferencer(env).Infer(sexp.Items[0])
	if err != nil {
		return fmtError(sexp.Items[0].Pos(), "%w", err)
	}

	return ot
}

// TypeInferencer structure to handle type inference
type TypeInferencer struct {
	env *object.Env
}

// NewTypeInferencer initializes a TypeInferencer with a given environment
func NewTypeInferencer(env *object.Env) *TypeInferencer {
	return &TypeInferencer{env: env}
}

// Infer infers the type of an AST node
func (ti *TypeInferencer) Infer(node ast.Node) (*object.Type, error) {
	switch n := node.(type) {
	case *ast.Literal[int64]:
		return &object.Type{
			ObjKind: TypeInt,
		}, nil
	case *ast.Literal[float64]:
		return &object.Type{
			ObjKind: TypeFloat,
		}, nil
	case *ast.Literal[string]:
		return &object.Type{
			ObjKind: TypeString,
		}, nil
	case *ast.Literal[bool]:
		return &object.Type{
			ObjKind: TypeBool,
		}, nil
	case *ast.Symbol:
		obj, _ := ti.env.LookUp(n.Value)
		switch obj := obj.(type) {
		case nil:
			return object.TypeFor(object.ObjNull), nil
		case *object.Builtin:
			return obj.Type, nil
		default:
			return &object.Type{
				ObjKind: obj.Kind(),
			}, nil
		}
	case *ast.If:
		condType, err := ti.Infer(n.Cond)
		if err != nil {
			return nil, fmt.Errorf("cond: %w", err)
		}
		if condType.ObjKind != object.ObjBool {
			return nil, fmt.Errorf("if condition is not boolean: got %s", condType.Inspect())
		}

		thenType, err := ti.Infer(n.Then)
		if err != nil {
			return nil, err
		}

		if n.Else == nil {
			return &object.Type{
				ObjKind: object.ObjAny, // because thenType | nil
			}, nil
		}

		elseType, err := ti.Infer(n.Else)
		if err != nil {
			return nil, fmt.Errorf("else: %w", err)
		}

		if !thenType.Equal(elseType) {
			return &object.Type{
				ObjKind: object.ObjAny, // because thenType | nil
			}, nil
		}

		return thenType, nil
	case *ast.SExp:
		return ti.inferSExp(n)
	case *ast.SpecialOp:
		return ti.inferSpecialOp(n)
	}

	return nil, fmt.Errorf("%w: unexpected node %s %q", errType, node.Name(), node.String())
}

func (ti *TypeInferencer) inferSExp(sexp *ast.SExp) (*object.Type, error) {
	if len(sexp.Items) == 0 {
		return &object.Type{
			ObjKind: TypeAny,
		}, nil
	}

	fnType, err := ti.Infer(sexp.Items[0]) // Infer function type
	if err != nil {
		return nil, fmt.Errorf("head of SExp: %w", err)
	}

	if fnType.ObjKind == object.ObjArray {
		if len(fnType.Params) == 0 {
			return object.TypeFor(TypeArray, object.ObjAny), nil
		}
	}

	return fnType, nil
}

func (ti *TypeInferencer) inferSpecialOp(op *ast.SpecialOp) (*object.Type, error) {
	switch op.Op {
	case "+", "*":
		return ti.inferArithmetic(op.Items)
	default:
		return nil, errType
	}
}

// inferArithmetic ensures consistent types (int or float) across operands
func (ti *TypeInferencer) inferArithmetic(items []ast.Node) (*object.Type, error) {
	hasFloats := false
	for _, item := range items {
		itemType, err := ti.Infer(item)
		if err != nil {
			return nil, fmt.Errorf("number arguments: %w", err)
		}
		if itemType.ObjKind == object.ObjFloat64 || itemType.ObjKind == object.ObjAST {
			hasFloats = true
		} else if itemType.ObjKind != object.ObjInteger {
			return nil, errType
		}
	}

	if hasFloats {
		return object.TypeFor(object.ObjFloat64), nil
	}
	return object.TypeFor(object.ObjInteger), nil
}
