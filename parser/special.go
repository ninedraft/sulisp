package parser

import (
	"fmt"

	"github.com/ninedraft/sulisp/language/ast"
)

var isSpecial = map[string]bool{
	"import-go": true,
	"if":        true, "cond": true,
	".":      true,
	"fn":     true,
	"assign": true,
	"handle": true,
	"while":  true,
	"begin":  true,
	"let":    true,
	"match":  true,
}

type operatorDef struct {
	minArgs int
	maxArgs int
}

var specialOperatorDefs = map[string]operatorDef{
	"+":   {minArgs: 1, maxArgs: -1},
	"*":   {minArgs: 1, maxArgs: -1},
	"-":   {minArgs: 2, maxArgs: 2},
	"/":   {minArgs: 2, maxArgs: 2},
	"<":   {minArgs: 2, maxArgs: 2},
	"<=":  {minArgs: 2, maxArgs: 2},
	">":   {minArgs: 2, maxArgs: 2},
	">=":  {minArgs: 2, maxArgs: 2},
	"not": {minArgs: 1, maxArgs: 1},
	"and": {minArgs: 2, maxArgs: -1},
	"or":  {minArgs: 2, maxArgs: -1},
}

func init() {
	for name := range specialOperatorDefs {
		isSpecial[name] = true
	}
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
	case "handle":
		return parser.buildHandle(sexp)
	case "while":
		return parser.buildWhile(sexp)
	case "begin":
		return parser.buildSequence(sexp)
	case "let":
		return parser.buildLet(sexp)
	case "match":
		return parser.buildMatch(sexp)
	}

	if sys, ok := specialOperatorDefs[head.Value]; ok {
		return parser.buildSpecialOperator(sexp, sys)
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

func (parser *Parser) buildHandle(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) < 4 {
		parser.errorf("handle requires effect name, operations and body")
		return nil
	}

	effect, ok := sexp.Items[1].(*ast.Symbol)
	if !ok {
		parser.errorf("handle effect must be a symbol")
		return nil
	}

	ops, ok := sexp.Items[2].(*ast.SExp)
	if !ok {
		parser.errorf("handle operations must be a list")
		return nil
	}

	operations := make([]*ast.HandleOp, 0, len(ops.Items))
	for _, item := range ops.Items {
		op, ok := item.(*ast.SExp)
		if !ok || len(op.Items) != 2 {
			parser.errorf("each handle operation must be (name function)")
			return nil
		}

		name, ok := op.Items[0].(*ast.Symbol)
		if !ok {
			parser.errorf("handle operation name must be symbol")
			return nil
		}

		operations = append(operations, &ast.HandleOp{
			PosRange: parser.posRange(),
			OpName:   name.Value,
			Body:     op.Items[1],
		})
	}

	body := sexp.Items[3:]
	if len(body) == 0 {
		parser.errorf("handle body missing")
		return nil
	}

	return &ast.Handle{
		PosRange:   parser.posRange(),
		Effect:     effect.Value,
		Operations: operations,
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

func (parser *Parser) buildLet(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) < 3 {
		parser.errorf("let requires bindings and body")
		return nil
	}

	bindingsExpr, ok := sexp.Items[1].(*ast.SExp)
	if !ok {
		parser.errorf("let bindings must be a list")
		return nil
	}

	if len(bindingsExpr.Items) == 0 {
		parser.errorf("let requires at least one binding")
		return nil
	}

	bindings := make([]*ast.Binding, 0, len(bindingsExpr.Items))
	for _, item := range bindingsExpr.Items {
		entry, ok := item.(*ast.SExp)
		if !ok || len(entry.Items) != 2 {
			parser.errorf("let binding must be (name value)")
			return nil
		}

		name, ok := entry.Items[0].(*ast.Symbol)
		if !ok {
			parser.errorf("let binding name must be symbol")
			return nil
		}

		value := entry.Items[1]
		if value == nil {
			parser.errorf("let binding value missing")
			return nil
		}

		bindings = append(bindings, &ast.Binding{
			PosRange:   parser.posRange(),
			Identifier: name.Value,
			Value:      value,
		})
	}

	body := sexp.Items[2:]
	if len(body) == 0 {
		parser.errorf("let body missing")
		return nil
	}

	return &ast.Let{
		PosRange: parser.posRange(),
		Bindings: bindings,
		Body:     body,
	}
}

func (parser *Parser) buildMatch(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) < 3 {
		parser.errorf("match requires expression and at least one case")
		return nil
	}

	expr := sexp.Items[1]
	if expr == nil {
		parser.errorf("match subject missing")
		return nil
	}

	cases := make([]*ast.MatchCase, 0, len(sexp.Items)-2)
	for i, item := range sexp.Items[2:] {
		caseIdx := i + 1
		caseExpr, ok := item.(*ast.SExp)
		if !ok {
			parser.errorf("match case %d must be a list", caseIdx)
			return nil
		}

		if len(caseExpr.Items) < 2 {
			parser.errorf("match case %d requires pattern and body", caseIdx)
			return nil
		}

		patternNode := caseExpr.Items[0]
		pattern, err := parser.buildPattern(patternNode)
		if err != nil {
			parser.errorf("match case %d: %w", caseIdx, err)
			return nil
		}

		body := caseExpr.Items[1:]
		if len(body) == 0 {
			parser.errorf("match case %d body missing", caseIdx)
			return nil
		}

		cases = append(cases, &ast.MatchCase{
			PosRange: parser.posRange(),
			Pattern:  pattern,
			Body:     body,
		})
	}

	return &ast.Match{
		PosRange: parser.posRange(),
		Expr:     expr,
		Cases:    cases,
	}
}

func (parser *Parser) buildPattern(node ast.Node) (ast.Pattern, error) {
	if node == nil {
		return nil, fmt.Errorf("pattern missing")
	}

	switch v := node.(type) {
	case *ast.Literal[int64]:
		return &ast.PatternLiteral{PosRange: v.Pos(), Value: v}, nil
	case *ast.Literal[float64]:
		return &ast.PatternLiteral{PosRange: v.Pos(), Value: v}, nil
	case *ast.Literal[string]:
		return &ast.PatternLiteral{PosRange: v.Pos(), Value: v}, nil
	case *ast.Literal[bool]:
		return &ast.PatternLiteral{PosRange: v.Pos(), Value: v}, nil
	case *ast.Keyword:
		return &ast.PatternLiteral{PosRange: v.Pos(), Value: v}, nil
	case *ast.Symbol:
		if v.Value == "_" {
			return &ast.PatternWildcard{PosRange: v.Pos()}, nil
		}
		return &ast.PatternVariable{PosRange: v.Pos(), Identifier: v.Value}, nil
	case *ast.SExp:
		items := make([]ast.Pattern, 0, len(v.Items))
		for _, child := range v.Items {
			pat, err := parser.buildPattern(child)
			if err != nil {
				return nil, err
			}
			items = append(items, pat)
		}
		return &ast.PatternSExp{PosRange: v.Pos(), Items: items}, nil
	}

	return nil, fmt.Errorf("unsupported pattern: %T", node)
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

func (parser *Parser) buildSequence(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) < 2 {
		parser.errorf("begin requires at least one expression")
		return nil
	}

	items := make([]ast.Node, 0, len(sexp.Items)-1)
	for _, item := range sexp.Items[1:] {
		if item == nil {
			parser.errorf("begin cannot contain nil expression")
			return nil
		}
		items = append(items, item)
	}

	return &ast.Sequence{
		PosRange: parser.posRange(),
		Items:    items,
	}
}

func (parser *Parser) buildWhile(sexp *ast.SExp) ast.Node {
	if len(sexp.Items) != 3 {
		parser.errorf("while requires condition and body")
		return nil
	}

	cond := sexp.Items[1]
	body := sexp.Items[2]

	if cond == nil {
		parser.errorf("while condition missing")
		return nil
	}
	if body == nil {
		parser.errorf("while body missing")
		return nil
	}

	return &ast.While{
		PosRange: parser.posRange(),
		Cond:     cond,
		Body:     body,
	}
}

func (parser *Parser) buildSpecialOperator(sexp *ast.SExp, def operatorDef) *ast.SpecialOp {
	if len(sexp.Items) < 2 {
		parser.errorf("operator must have at least 1 operand")
		return nil
	}

	head := sexp.Items[0].(*ast.Symbol)

	count := len(sexp.Items) - 1
	if count < def.minArgs {
		parser.errorf("operator %s requires at least %d operands, got %d", head.Value, def.minArgs, count)
		return nil
	}

	if def.maxArgs >= 0 && count > def.maxArgs {
		parser.errorf("operator %s requires at most %d operands, got %d", head.Value, def.maxArgs, count)
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
