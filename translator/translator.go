package translator

import (
	"errors"
	"fmt"
	goast "go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/ninedraft/sulisp/language"
	"golang.org/x/exp/slices"
)

func TranslateFile(root []language.Sexp) (*goast.File, error) {
	tr := &translator{
		target: &goast.File{},
	}

	for _, decl := range root {
		if err := tr.translateDecl(decl); err != nil {
			return nil, err
		}
	}

	return tr.target, nil
}

var ErrInvalidDecl = errors.New("invalid declaration")

var ErrInvalidExpr = errors.New("invalid expression")

type translator struct {
	target       *goast.File
	identCounter int64
}

func (t *translator) translateDecl(in language.Sexp) error {
	if len(in) == 0 {
		return ErrInvalidDecl
	}

	head := in[0]
	var decl goast.Decl
	var err error
	switch {
	case head.Equal(language.Symbol("package")):
		return t.translatePackage(in)
	case head.Equal(language.Symbol("import")):
		decl, err = t.translateImport(in)
	case head.Equal(language.Symbol("defn")):
		decl, err = t.translateFunc(in)
	case head.Equal(language.Symbol("var")):
		decl, err = t.translateVar(in)
	default:
		return fmt.Errorf("%s: %w", in, ErrInvalidDecl)
	}

	if err != nil {
		return err
	}

	if decl != nil {
		t.target.Decls = append(t.target.Decls, decl)
	}
	return nil
}

func (t *translator) translatePackage(in language.Sexp) error {
	if len(in) != 2 {
		return fmt.Errorf("%s: %w: want (package package/name)", in, ErrInvalidDecl)
	}

	name, ok := in[1].(language.Symbol)
	if !ok {
		return fmt.Errorf("%s: %w", in, ErrInvalidDecl)
	}

	t.target.Name = goast.NewIdent(name.String())
	return nil
}

/*
	(import
		fmt
		io
		strconv)

translates to

	import (
		"fmt"
		"io"
		"strconv"
	)
*/
func (t *translator) translateImport(in language.Sexp) (*goast.GenDecl, error) {
	decl := &goast.GenDecl{
		Tok: token.IMPORT,
	}

	if len(in) < 2 {
		return nil, fmt.Errorf("%s: %w: want (import package/name...)", in, ErrInvalidDecl)
	}

	for _, pkg := range in[1:] {
		name, ok := pkg.(language.Symbol)
		if !ok {
			return nil, fmt.Errorf("%s: %w", in, ErrInvalidDecl)
		}

		decl.Specs = append(decl.Specs, &goast.ImportSpec{
			Path: &goast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(name.String()),
			},
		})
	}

	return decl, nil
}

/*
(defn name (arg1 arg2 :- typ1 arg3 :- typ2 > result1 result2 ...)

	body_expr)

translates to

	func name(arg1 arg2 typ1, arg3 typ2) (result1, result2...) {
		body_expr
	}
*/
func (t *translator) translateFunc(in language.Expression) (*goast.FuncDecl, error) {
	spec, ok := in.(language.Sexp)
	if !ok {
		return nil, fmt.Errorf("%s: %w: expected (defn (arg1 arg2 :- typ1 arg3 :- typ2 => result) body...): unexpected input %T",
			in, ErrInvalidDecl, in)
	}

	if len(spec) != 4 {
		return nil, fmt.Errorf("%s: %w: expected (defn (arg1 arg2 :- typ1 arg3 :- typ2 => result) body...): not enough items %d", in, ErrInvalidDecl, len(spec))
	}

	h := spec[1]
	name, ok := h.(language.Symbol)
	if !ok {
		return nil, fmt.Errorf("%s: %w: expected (defn (arg1 arg2 :- typ1 arg3 :- typ2 => result) body...): unexpected name ident %T", in, ErrInvalidDecl, h)
	}

	head, ok := spec[2].(language.Sexp)
	if !ok {
		return nil, fmt.Errorf("%s: %w: expected (defn (arg1 arg2 :- typ1 arg3 :- typ2 => result) body...): unexpected function spec %T", in, ErrInvalidDecl, spec[2])
	}

	sign, errSign := t.translateFuncType(head)
	if errSign != nil {
		return nil, fmt.Errorf("%s: %w", in, errSign)
	}

	var body []goast.Stmt
	for i, expr := range spec[3:] {
		parsed, err := t.translateExpr(expr)
		if err != nil {
			return nil, fmt.Errorf("%w: body expr %d", err, i+1)
		}

		var stmt goast.Stmt = &goast.ExprStmt{
			X: parsed,
		}
		if i == len(spec[3:])-1 && len(sign.Results.List) > 0 {
			stmt = &goast.ReturnStmt{
				Results: []goast.Expr{parsed},
			}
		}

		body = append(body, stmt)
	}

	return &goast.FuncDecl{
		Name: goast.NewIdent(name.String()),
		Type: sign,
		Body: &goast.BlockStmt{
			List: body,
		},
	}, nil
}

// parses function signatures
//
//	(x y :- int -> int error) translates to func(x int, y int) (int, error)
//	() translates to func()
//	(x :- int) translates to func(x int)
//	(-> int) translates to func() int
//
// etc
func (t *translator) translateFuncType(spec language.Sexp) (*goast.FuncType, error) {
	returnSepIdx := slices.IndexFunc(spec, func(v language.Expression) bool {
		return isSymbol(v, "=>")
	})
	argsToParse := spec
	if returnSepIdx > 0 {
		argsToParse = spec[:returnSepIdx]
	}

	args := &goast.FieldList{}
	{
		field := &goast.Field{}
		consumeArgName := func(expr language.Expression) error {
			name, ok := expr.(language.Symbol)
			if !ok {
				return fmt.Errorf("expected symbol, got %T", expr)
			}
			field.Names = append(field.Names, goast.NewIdent(name.String()))
			return nil
		}

		consume := consumeArgName

		consumeType := func(expr language.Expression) error {
			t, err := t.translateExpr(expr)
			if err != nil {
				return err
			}
			field.Type = t
			args.List = append(args.List, field)
			field = &goast.Field{}
			consume = consumeArgName
			return nil
		}

		for _, arg := range argsToParse {
			if isKeyword(arg, "-") {
				consume = consumeType
				continue
			}
			if err := consume(arg); err != nil {
				return nil, err
			}
		}
	}

	result := &goast.FieldList{}
	if returnSepIdx > 0 {
		for _, arg := range spec[returnSepIdx+1:] {
			t, err := t.translateExpr(arg)
			if err != nil {
				return nil, err
			}
			result.List = append(result.List, &goast.Field{
				Type: t,
			})
		}
	}

	return &goast.FuncType{
		Params: &goast.FieldList{
			List: args.List,
		},
		Results: &goast.FieldList{
			List: result.List,
		},
	}, nil
}

type literal interface {
	language.Expression
	IsLiteral()
}

func is[E any](v any) bool {
	_, ok := v.(E)
	return ok
}

func isCall(expr language.Expression, of ...string) bool {
	call, ok := expr.(language.Sexp)
	if !ok {
		return false
	}

	if len(call) == 0 {
		return false
	}

	sym, ok := call[0].(language.Symbol)
	if !ok {
		return false
	}
	name := sym.String()

	return slices.Contains(of, name)
}

func isSymbol(expr language.Expression, of ...string) bool {
	sym, ok := expr.(language.Symbol)
	if !ok {
		return false
	}

	return slices.Contains(of, sym.String())
}

func isKeyword(expr language.Expression, of ...string) bool {
	kw, ok := expr.(language.Keyword)
	if !ok {
		return false
	}

	return slices.Contains(of, string(kw))
}

func (t *translator) translateExpr(in language.Expression) (goast.Expr, error) {
	if in == nil {
		return nil, ErrInvalidExpr
	}

	var expr goast.Expr
	var err error
	switch {
	case is[literal](in):
		expr, err = t.translateLiteral(in.(literal))
	case isCall(in, "+", "-", "*", "/"):
		expr, err = t.translateBinaryOp(in.(language.Sexp))
	case isCall(in, "go"):
		expr, err = t.translateGoroutineStart(in.(language.Sexp))
	case is[language.Symbol](in):
		expr = goast.NewIdent(in.(language.Symbol).String())
	default: // function call
		expr, err = t.translateCall(in)
	}

	if err != nil {
		return nil, err
	}

	return expr, nil
}

/*
	(go fn arg1 arg2...)

translates to

	func() {} struct{} {
		go fn(arg1, arg2...)

		return struct{}{}
	}()
*/
func (t *translator) translateGoroutineStart(in language.Sexp) (goast.Expr, error) {
	if len(in) < 2 {
		return nil, fmt.Errorf("%s: %w: expected (go fn arg1 arg2...)", in, ErrInvalidExpr)
	}

	fn, err := t.translateExpr(in[1])
	if err != nil {
		return nil, err
	}

	var args []goast.Expr
	for _, arg := range in[2:] {
		argExpr, err := t.translateExpr(arg)
		if err != nil {
			return nil, err
		}

		args = append(args, argExpr)
	}

	internal := &goast.CallExpr{
		Fun:  fn,
		Args: args,
	}

	goroutine := &goast.GoStmt{
		Call: internal,
	}

	return &goast.CallExpr{
		Fun: &goast.FuncLit{
			Type: &goast.FuncType{
				Params: &goast.FieldList{},
				Results: &goast.FieldList{
					List: []*goast.Field{
						{
							Type: &goast.StructType{
								Fields: &goast.FieldList{},
							},
						},
					},
				},
			},
			Body: &goast.BlockStmt{
				List: []goast.Stmt{goroutine,
					&goast.ReturnStmt{
						Results: []goast.Expr{
							&goast.CompositeLit{
								Type: &goast.StructType{Fields: &goast.FieldList{}},
							},
						},
					},
				},
			},
		},
	}, nil
}

/*
	(foo 1 2 3)

translates to

	foo(1, 2, 3)
*/
func (t *translator) translateCall(in language.Expression) (goast.Expr, error) {
	spec, ok := in.(language.Sexp)
	if !ok {
		return nil, fmt.Errorf("%s: %w: expected (foo arg1 arg2...)", in, ErrInvalidExpr)
	}

	if len(spec) < 1 {
		return nil, fmt.Errorf("%s: %w: expected (foo arg1 arg2...)", in, ErrInvalidExpr)
	}

	head := spec[0]
	name, ok := head.(language.Symbol)
	if !ok {
		return nil, fmt.Errorf("%s: %w: expected (foo arg1 arg2...)", in, ErrInvalidExpr)
	}

	var args []goast.Expr
	for _, arg := range spec[1:] {
		argExpr, err := t.translateExpr(arg)
		if err != nil {
			return nil, fmt.Errorf("%s: arg %d: %w", name, len(args), err)
		}
		args = append(args, argExpr)
	}

	return &goast.CallExpr{
		Fun:  goast.NewIdent(name.String()),
		Args: args,
	}, nil
}

/*
translates variable declaration

	(var
		x 1
		y 2
		x (+ 1 2))
*/
func (t *translator) translateVar(in language.Sexp) (*goast.GenDecl, error) {
	if len(in) < 2 {
		return nil, fmt.Errorf("%s: %w: want (var name value...)", in, ErrInvalidDecl)
	}

	decl := &goast.GenDecl{
		Tok: token.VAR,
	}

	if len(in)%2 != 1 {
		return nil, fmt.Errorf("%s: %w: want (var name value...)", in, ErrInvalidDecl)
	}

	for i := 1; i < len(in); i += 2 {
		name, ok := in[i].(language.Symbol)
		if !ok {
			return nil, fmt.Errorf("%s: %w", in, ErrInvalidDecl)
		}

		value := in[i+1]
		valueExpr, errValue := t.translateExpr(value)
		if errValue != nil {
			return nil, fmt.Errorf("%s: %w", in, ErrInvalidExpr)
		}

		decl.Specs = append(decl.Specs, &goast.ValueSpec{
			Names:  []*goast.Ident{goast.NewIdent(name.String())},
			Values: []goast.Expr{valueExpr},
		})
	}

	return decl, nil
}

func (t *translator) translateLiteral(in literal) (goast.Expr, error) {
	switch v := in.(type) {
	case *language.Literal[string]:
		return &goast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(v.Value),
		}, nil
	case *language.Literal[int]:
		return &goast.BasicLit{
			Kind:  token.INT,
			Value: v.String(),
		}, nil
	case *language.Literal[float64]:
		return &goast.BasicLit{
			Kind:  token.FLOAT,
			Value: v.String(),
		}, nil
	case *language.Literal[bool]:
		return goast.NewIdent(v.String()), nil
	}

	return nil, fmt.Errorf("%s: %w: invalid literal", in, ErrInvalidExpr)
}

func (t *translator) gensym() string {
	t.identCounter++
	return fmt.Sprintf("gensym_%d", t.identCounter)
}

func (t *translator) gensymIdent() *goast.Ident {
	return goast.NewIdent(t.gensym())
}

const binaryOps = "+-*/"

func (t *translator) translateOperator(in language.Sexp) (goast.Expr, error) {
	if len(in) != 3 {
		return nil, fmt.Errorf("%s: %w: expected (op arg1 arg2)", in, ErrInvalidExpr)
	}

	name := in[0].(language.Symbol).String()
	if strings.Contains(binaryOps, name) {
		return t.translateBinaryOp(in)
	}

	return nil, fmt.Errorf("%s: %w: unknown operator", in, ErrInvalidExpr)
}

var operators1 = map[string]token.Token{
	"+": token.ADD,
	"-": token.SUB,
	"*": token.MUL,
	"/": token.QUO,
}

/*
	(op arg1 arg2 arg3...)

translates to

	(arg1 op arg2 op arg3...)
*/
func (t *translator) translateBinaryOp(in language.Sexp) (goast.Expr, error) {
	if len(in) < 3 {
		return nil, fmt.Errorf("%s: %w: expected (op arg1 arg2...)", in, ErrInvalidExpr)
	}

	op := in[0].(language.Symbol).String()
	var expr goast.Expr
	for _, arg := range in[1:] {
		argExpr, err := t.translateExpr(arg)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", in, ErrInvalidExpr)
		}

		if expr == nil {
			expr = argExpr
			continue
		}

		expr = &goast.BinaryExpr{
			X:  expr,
			Op: operators1[op],
			Y:  argExpr,
		}
	}

	return &goast.ParenExpr{
		X: expr,
	}, nil
}
