// Package maplookup detects repeated map lookups with the same key.
package maplookup

import (
	"go/ast"
	"go/types"

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

		condLookups := findMapLookups(pass, ifStmt.Cond)
		if len(condLookups) == 0 {
			return
		}

		bodyLookups := findMapLookupsInBlock(pass, ifStmt.Body)

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

// isMapType checks if the expression is a map type.
func isMapType(pass *analysis.Pass, expr ast.Expr) bool {
	tv := pass.TypesInfo.TypeOf(expr)
	if tv == nil {
		return false
	}
	_, ok := tv.Underlying().(*types.Map)
	return ok
}

func findMapLookups(pass *analysis.Pass, expr ast.Expr) []*ast.IndexExpr {
	var lookups []*ast.IndexExpr
	ast.Inspect(expr, func(n ast.Node) bool {
		idx, ok := n.(*ast.IndexExpr)
		if !ok {
			return true
		}
		// Only include actual map lookups, not slice indexing
		if isMapType(pass, idx.X) {
			lookups = append(lookups, idx)
		}
		return true
	})
	return lookups
}

func findMapLookupsInBlock(pass *analysis.Pass, block *ast.BlockStmt) []*ast.IndexExpr {
	var lookups []*ast.IndexExpr
	for _, stmt := range block.List {
		// Skip assignments where map lookup is on the left side (setting, not getting)
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			// Check right-hand side only
			for _, rhs := range assign.Rhs {
				ast.Inspect(rhs, func(n ast.Node) bool {
					idx, ok := n.(*ast.IndexExpr)
					if !ok {
						return true
					}
					if isMapType(pass, idx.X) {
						lookups = append(lookups, idx)
					}
					return true
				})
			}
			continue
		}

		// For other statements, inspect normally
		ast.Inspect(stmt, func(n ast.Node) bool {
			idx, ok := n.(*ast.IndexExpr)
			if !ok {
				return true
			}
			if isMapType(pass, idx.X) {
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
