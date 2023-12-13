package parser

import "github.com/ninedraft/sulisp/language/ast"

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
