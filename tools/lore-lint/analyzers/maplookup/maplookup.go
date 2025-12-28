// Package maplookup detects repeated map lookups with the same key.
package maplookup

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer detects repeated map lookups with the same key.
var Analyzer = &analysis.Analyzer{
	Name:     "maplookup",
	Doc:      "detects repeated map lookups with the same key",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return
		}

		// Skip if statements with Init (already using := pattern)
		if ifStmt.Init != nil {
			return
		}

		condLookups := findMapLookups(ifStmt.Cond)
		if len(condLookups) == 0 {
			return
		}

		bodyLookups := findMapLookupsInBlock(ifStmt.Body)

		for _, condLookup := range condLookups {
			for _, bodyLookup := range bodyLookups {
				if sameMapLookup(condLookup, bodyLookup) {
					pass.Reportf(bodyLookup.Pos(),
						"repeated map lookup - store result in variable using := in if statement")
				}
			}
		}
	})

	return nil, nil
}

func findMapLookups(expr ast.Expr) []*ast.IndexExpr {
	var lookups []*ast.IndexExpr
	ast.Inspect(expr, func(n ast.Node) bool {
		if idx, ok := n.(*ast.IndexExpr); ok {
			lookups = append(lookups, idx)
		}
		return true
	})
	return lookups
}

func findMapLookupsInBlock(block *ast.BlockStmt) []*ast.IndexExpr {
	var lookups []*ast.IndexExpr
	for _, stmt := range block.List {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if idx, ok := n.(*ast.IndexExpr); ok {
				lookups = append(lookups, idx)
			}
			return true
		})
	}
	return lookups
}

func sameMapLookup(a, b *ast.IndexExpr) bool {
	return exprString(a.X) == exprString(b.X) &&
		exprString(a.Index) == exprString(b.Index)
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprString(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return exprString(e.X) + "[" + exprString(e.Index) + "]"
	default:
		return ""
	}
}
