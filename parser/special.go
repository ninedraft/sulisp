package parser

import (
	"github.com/ninedraft/sulisp/language/ast"
	"golang.org/x/exp/maps"
)

var isSpecial = map[string]bool{
	"import-go": true,
	"if":        true, "cond": true,
	"+": true, "-": true, "*": true, "/": true,
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
	}

	if specialOperators[head.Value] {
		return parser.buildSpecialOperator(sexp)
	}

	parser.errorf("unknown special form %s", head.Value)
	return nil
}

func (parser *Parser) buildIf(sexp *ast.SExp) *ast.If {
	if len(sexp.Items) < 2 || len(sexp.Items) > 4 {
		parser.errorf("if form must have 2 or 3 items")
		return nil
	}

	head := sexp.Items[0].(*ast.Symbol)
	if head.Value == "cond" && len(sexp.Items) > 3 {
		parser.errorf("cond form must have at most 2 items")
		return nil
	}

	ifForm := &ast.If{
		PosRange: parser.posRange(),
		Cond:     sexp.Items[1],
		Then:     sexp.Items[2],
	}

	if len(sexp.Items) > 3 {
		ifForm.Else = sexp.Items[3]
	}

	return ifForm
}

func (parser *Parser) buildImportGo(sexp *ast.SExp) *ast.ImportGo {
	if len(sexp.Items) == 1 {
		parser.errorf("empty import-go")
		return nil
	}

	importgo := &ast.ImportGo{
		PosRange: parser.posRange(),
	}

	validateAliasItem := func(item *ast.SExp) bool {
		// (alias string) or (alias symbol)

		alias := pMatch[*ast.Symbol]()

		ref := pOr(
			pMatch[*ast.Literal[string]](),
			pMatch[*ast.Symbol](),
		)

		return sexpMatch(item, alias, ref)
	}

	for i, item := range sexp.Items[1:] {
		switch item := item.(type) {
		case *ast.Literal[string], *ast.Symbol:
			importgo.Items = append(importgo.Items, item)
		case *ast.SExp:
			if !validateAliasItem(item) {
				parser.errorf("unexpected import-go item %d: %s", i+1, item.Tok())
				return nil
			}
			importgo.Items = append(importgo.Items, item)
		default:
			parser.errorf("unexpected import-go item: %s", item.Tok())
			return nil
		}
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
