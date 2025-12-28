// Package analyzers provides all custom static analyzers for lore-core.
package analyzers

import (
	"golang.org/x/tools/go/analysis"

	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/loopcall"
)

// All returns all analyzers to run.
func All() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		loopcall.Analyzer,
	}
}
