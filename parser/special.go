package parser

import (
	"github.com/ninedraft/sulisp/language/ast"
	"golang.org/x/exp/maps"
)

var isSpecial = map[string]bool{
	"import-go": true,
	"if":        true, "cond": true,
	"+": true, "-": true, "*": true, "/": true,
	".":      true,
	"fn":     true,
	"assign": true,
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
	case "fn":
		return parser.buildFunction(sexp)
	case "assign":
		return parser.buildAssign(sexp)
	}

	if specialOperators[head.Value] {
		return parser.buildSpecialOperator(sexp)
	}

	parser.errorf("unknown special form %s", head.Value)
	return nil
}

func (parser *Parser) buildFunction(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) < 4 {
		parser.errorf("fn form requires identifier, params, and body")
		return nil
	}

	name, ok := sexp.Items[1].(*ast.Symbol)
	if !ok {
		parser.errorf("fn name must be a symbol")
		return nil
	}

	params, ok := sexp.Items[2].(*ast.SExp)
	if !ok {
		parser.errorf("fn params must be a list")
		return nil
	}

	paramSymbols := make([]*ast.Symbol, 0, len(params.Items))
	for _, item := range params.Items {
		sym, ok := item.(*ast.Symbol)
		if !ok {
			parser.errorf("fn params must be symbols")
			return nil
		}
		paramSymbols = append(paramSymbols, sym)
	}

	body := sexp.Items[3]
	if body == nil {
		parser.errorf("fn body missing")
		return nil
	}

	return &ast.Function{
		PosRange:   parser.posRange(),
		Identifier: name.Value,
		Parameters: paramSymbols,
		Body:       body,
	}
}

func (parser *Parser) buildAssign(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) != 3 {
		parser.errorf("assign requires target and value")
		return nil
	}

	target, ok := sexp.Items[1].(*ast.Symbol)
	if !ok {
		parser.errorf("assign target must be symbol")
		return nil
	}

	value := sexp.Items[2]
	if value == nil {
		parser.errorf("assign value missing")
		return nil
	}

	return &ast.Assign{
		PosRange: parser.posRange(),
		Target:   target,
		Value:    value,
	}
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

func (parser *Parser) buildDotSelector(sexp *ast.SExp) *ast.DotSelector {
	var left, right ast.Node
	dot := &ast.Symbol{Value: "."}

	errMatch := sexpMatch(sexp, pEq(&dot), p(&left), p(&right))

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
