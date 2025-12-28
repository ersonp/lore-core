// Package analyzers provides all custom static analyzers for lore-core.
package analyzers

import (
	"golang.org/x/tools/go/analysis"

	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/loopcall"
	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/maplookup"
	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/nestedloop"
	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/regexloop"
	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/stringconcat"
)

// All returns all analyzers to run.
func All() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		loopcall.Analyzer,
		maplookup.Analyzer,
		nestedloop.Analyzer,
		regexloop.Analyzer,
		stringconcat.Analyzer,
	}
}
