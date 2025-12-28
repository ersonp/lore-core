// Package loopcall detects API/database calls inside loops.
package loopcall

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects API/database calls inside loops that should be batched.
var Analyzer = &analysis.Analyzer{
	Name:     "loopcall",
	Doc:      "detects API/database calls inside loops that should be batched",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// externalMethods are method names that indicate external calls.
var externalMethods = map[string]bool{
	// Embedder interface
	"Embed": true,
	// VectorDB interface (non-batch)
	"Save":   true,
	"Search": true,
	"Delete": true,
	// LLMClient interface
	"Extract":          true,
	"ExtractFacts":     true,
	"CheckConsistency": true,
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

			methodName := sel.Sel.Name
			if externalMethods[methodName] {
				if hasNolintDirective(pass, call.Pos()) {
					return true
				}
				pass.Reportf(call.Pos(),
					"potential N+1: %s called inside loop - consider batching",
					methodName)
			}

			return true
		})
	})

	return nil, nil
}

// hasNolintDirective checks if there's a nolint:loopcall comment for the given position.
func hasNolintDirective(pass *analysis.Pass, pos token.Pos) bool {
	file := pass.Fset.File(pos)
	if file == nil {
		return false
	}

	line := file.Line(pos)

	for _, f := range pass.Files {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				commentLine := file.Line(c.Pos())
				// Check comment on same line or line before
				if commentLine == line || commentLine == line-1 {
					if strings.Contains(c.Text, "nolint:loopcall") {
						return true
					}
				}
			}
		}
	}

	return false
}
