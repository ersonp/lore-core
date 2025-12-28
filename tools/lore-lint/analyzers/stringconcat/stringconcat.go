// Package stringconcat detects O(n²) string concatenation in loops.
package stringconcat

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects O(n²) string concatenation patterns in loops.
var Analyzer = &analysis.Analyzer{
	Name:     "stringconcat",
	Doc:      "detects O(n²) string concatenation patterns in loops",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
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
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}

			if assign.Tok != token.ADD_ASSIGN {
				return true
			}

			if len(assign.Lhs) != 1 {
				return true
			}

			if isStringType(pass, assign.Lhs[0]) {
				pass.Reportf(assign.Pos(),
					"O(n²) string concatenation in loop - use strings.Builder")
			}

			return true
		})
	})

	return nil, nil
}

// isStringType checks if the expression has string type.
func isStringType(pass *analysis.Pass, expr ast.Expr) bool {
	tv := pass.TypesInfo.TypeOf(expr)
	if tv == nil {
		return false
	}

	basic, ok := tv.Underlying().(*types.Basic)
	if !ok {
		return false
	}

	return basic.Kind() == types.String
}
