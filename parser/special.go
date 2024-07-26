package parser

import (
	"github.com/ninedraft/sulisp/language/ast"
	"golang.org/x/exp/maps"
)

var isSpecial = map[string]bool{
	"import-go": true,
	"if":        true, "cond": true,
	"+": true, "-": true, "*": true, "/": true,
	".":   true,
	"*fn": true,
}

var specialOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true,
}

func init() {
	maps.Copy(isSpecial, specialOperators)
}

func (parser *Parser) buildSpecial(sexp *ast.SExp) ast.Node {
	head, ok := sexp.Items[0].(*ast.Symbol)
	if !ok {
		parser.errorf("special form head must be a symbol")
		return nil
	}

	switch head.Value {
	case "import-go":
		return parser.buildImportGo(sexp)
	case "if", "cond":
		return parser.buildIf(sexp)
	case ".":
		return parser.buildDotSelector(sexp)
	case "*fn":
		return parser.buildFunctionDecl(sexp)
	}

	if specialOperators[head.Value] {
		return parser.buildSpecialOperator(sexp)
	}

	parser.errorf("unknown special form %s", head.Value)
	return nil
}

func (parser *Parser) buildIf(sexp *ast.SExp) *ast.If {
	var head *ast.Symbol // 'if or 'cond
	var cond ast.Node
	var then_ ast.Node
	var else_ ast.Node // optional

	errMatch := sexpMatch(sexp, p(&head),
		p(&cond),
		p(&then_),
		pOpt(&else_),
	)

	if errMatch != nil {
		parser.errorf("invalid if form: %w", errMatch)
		return nil
	}

	if head.Value == "cond" && else_ != nil {
		parser.errorf("cond form must not have else branch")
		return nil
	}

	return &ast.If{
		PosRange: parser.posRange(),
		Cond:     cond,
		Then:     then_,
		Else:     else_,
	}
}

var matchImportGoItem = pOr(
	pMatch[*ast.Literal[string]](),
	pMatch[*ast.Symbol](),
	matchImportGoAliasItem,
)

var matchImportGoAliasItem = pSexp(
	pMatch[*ast.Symbol](),
	pOr(
		pMatch[*ast.Literal[string]](),
		pMatch[*ast.Symbol](),
	),
)

func (parser *Parser) buildImportGo(sexp *ast.SExp) *ast.ImportGo {
	importgo := &ast.ImportGo{
		PosRange: parser.posRange(),
	}

	if len(sexp.Items) == 1 {
		return importgo
	}

	for i, item := range sexp.Items[1:] {
		err := matchImportGoItem(item)
		if err != nil {
			parser.errorf("import-go item %d: %w", i+1, err)
			return nil
		}
		importgo.Items = append(importgo.Items, item)
	}

	return importgo
}

func (parser *Parser) buildSpecialOperator(sexp *ast.SExp) *ast.SpecialOp {
	if len(sexp.Items) < 2 {
		parser.errorf("operator must have at least 1 operand")
		return nil
	}

	head := sexp.Items[0].(*ast.Symbol)
	if !specialOperators[head.Value] {
		parser.errorf("expected an operator, got %s", head.Value)
		return nil
	}

	// only commutative operators can have more than 2 operands
	if (head.Value != "+" && head.Value != "*") && len(sexp.Items) != 3 {
		parser.errorf("operator %s must have 2 operands, got %d", head.Value, len(sexp.Items)-1)
		return nil
	}

	return &ast.SpecialOp{
		PosRange: parser.posRange(),
		Op:       head.Value,
		Items:    sexp.Items[1:],
	}
}

var dot = &ast.Symbol{Value: "."}

func (parser *Parser) buildDotSelector(sexp *ast.SExp) *ast.DotSelector {
	var left, right ast.Node

	errMatch := sexpMatch(sexp, pEq(dot), p(&left), p(&right))

	if errMatch != nil {
		parser.errorf("invalid dot selector: %w", errMatch)
		return nil
	}

	return &ast.DotSelector{
		PosRange: parser.posRange(),
		Left:     left,
		Right:    right,
	}
}

var fnHead = &ast.Symbol{Value: "*fn"}

/*
Examples:

	(*fn foo (p1 :_ t1 p2 :_ t2) (r1 r2 r3) (body)) ; multiple results, with name
	(*fn (p1 :_ t1 p2 :_ t2) (body)) ; no results, anonymous
*/
func (parser *Parser) buildFunctionDecl(sexp *ast.SExp) *ast.FunctionDecl {
	if len(sexp.Items) < 3 {
		parser.errorf("function decl must have at least 3 arguments")
		return nil
	}

	if err := pEq(fnHead)(sexp.Items[0]); err != nil {
		parser.errorf("invalid function decl: %w", err)
		return nil
	}

	rest := sexp.Items[1:]
	name, rest := matchConsume[*ast.Symbol](rest)
	params, rest := matchConsume[*ast.SExp](rest)
	if params == nil {
		parser.errorf("function declaration must have params, got %s", rest[0].Name())
		return nil
	}

	if len(rest) == 0 {
		parser.errorf("function declaration must have body")
		return nil
	}

	fnSpec := &ast.SExp{Items: []ast.Node{params}}

	if len(rest) > 1 {
		fnSpec.Items = append(fnSpec.Items, rest[0])
		rest = rest[1:]
	}

	spec := parser.buildFnSpec(fnSpec)
	if spec == nil {
		parser.errorf("invalid function declaration spec")
		return nil
	}

	return &ast.FunctionDecl{
		PosRange: parser.posRange(),
		FnName:   name,
		Spec:     spec,
		Body:     rest[0],
	}
}

func matchConsume[N ast.Node](nodes []ast.Node) (N, []ast.Node) {
	var empty N
	if len(nodes) == 0 {
		return empty, nil
	}

	n, ok := nodes[0].(N)
	if !ok {
		return empty, nodes
	}

	return n, nodes[1:]
}

var paramTypeSep = &ast.Keyword{Value: "_"}

/*
Example:

	((p1 :_ t1 p2 :_ t2) (r1 r2 r3)) ; multiple results
	((p1 :_ t1 p2 :_ tc)) ; no results
*/
func (parser *Parser) buildFnSpec(sexp *ast.SExp) *ast.FunctionSpec {
	if len(sexp.Items) < 1 {
		parser.errorf("function spec must have at least 1 argument")
		return nil
	}

	var params, results *ast.SExp

	errMatch := sexpMatch(sexp, p(&params), pOpt(&results))
	if errMatch != nil {
		parser.errorf("invalid function spec: %w", errMatch)
		return nil
	}

	if len(params.Items)%3 != 0 {
		parser.errorf("function spec params must consist of $name :_ $type pairs, got %d items", len(params.Items))
		return nil
	}

	paramFields := make([]*ast.FieldSpec, 0, len(params.Items)/2)
	for i := 0; i < len(params.Items); i += 3 {
		var name, typ ast.Node
		field := &ast.SExp{Items: params.Items[i : i+3]}
		errMatch = sexpMatch(field, p(&name), pEq(paramTypeSep), p(&typ))
		if errMatch != nil {
			parser.errorf("invalid param spec %s: %w", field, errMatch)
			return nil
		}

		paramFields = append(paramFields, &ast.FieldSpec{
			Name: name,
			Type: typ,
		})
	}

	if results == nil {
		results = &ast.SExp{}
	}

	return &ast.FunctionSpec{
		PosRange: parser.posRange(),
		Params:   paramFields,
		Ret:      results.Items,
	}
}
