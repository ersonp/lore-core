// Package nestedloop detects O(n²) nested loops over the same collection.
package nestedloop

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects O(n²) nested loops over the same collection.
var Analyzer = &analysis.Analyzer{
	Name:     "nestedloop",
	Doc:      "detects O(n²) nested loops over the same collection",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		outer, ok := n.(*ast.RangeStmt)
		if !ok {
			return
		}

		outerIdent := getIdent(outer.X)
		if outerIdent == "" {
			return
		}

		ast.Inspect(outer.Body, func(n ast.Node) bool {
			inner, ok := n.(*ast.RangeStmt)
			if !ok {
				return true
			}

			innerIdent := getIdent(inner.X)
			if innerIdent == outerIdent {
				pass.Reportf(inner.Pos(),
					"O(n²) pattern: nested loop over same collection %q - consider using a map",
					outerIdent)
			}

			return true
		})
	})

	return nil, nil
}

// getIdent extracts the identifier name from an expression.
func getIdent(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name + "." + e.Sel.Name
		}
	}
	return ""
}
