// Package regexloop detects regex compilation inside loops.
package regexloop

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects regexp.Compile/MustCompile calls inside loops.
var Analyzer = &analysis.Analyzer{
	Name:     "regexloop",
	Doc:      "detects regexp.Compile/MustCompile calls inside loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var regexpFuncs = map[string]bool{
	"Compile":          true,
	"MustCompile":      true,
	"CompilePOSIX":     true,
	"MustCompilePOSIX": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.RangeStmt)(nil),
		(*ast.ForStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		var body *ast.BlockStmt
		switch stmt := n.(type) {
		case *ast.RangeStmt:
			body = stmt.Body
		case *ast.ForStmt:
			body = stmt.Body
		}
		if body == nil {
			return
		}

		ast.Inspect(body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "regexp" && regexpFuncs[sel.Sel.Name] {
				pass.Reportf(call.Pos(),
					"regexp.%s called inside loop - compile once outside loop",
					sel.Sel.Name)
			}

			return true
		})
	})

	return nil, nil
}
