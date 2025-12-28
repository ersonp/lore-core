// lore-lint is a custom static analyzer for lore-core performance patterns.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/ersonp/lore-core/tools/lore-lint/analyzers"
)

func main() {
	multichecker.Main(analyzers.All()...)
}
