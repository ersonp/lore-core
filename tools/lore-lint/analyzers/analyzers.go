// Package analyzers provides all custom static analyzers for lore-core.
package analyzers

import "golang.org/x/tools/go/analysis"

// All returns all analyzers to run.
func All() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		// Analyzers will be added here as they're implemented
	}
}
