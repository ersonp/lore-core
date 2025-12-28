// Package loopcall detects API/database calls inside loops.
package loopcall

import (
	"go/ast"

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
				pass.Reportf(call.Pos(),
					"potential N+1: %s called inside loop - consider batching",
					methodName)
			}

			return true
		})
	})

	return nil, nil
}
