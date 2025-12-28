package maplookup_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/ersonp/lore-core/tools/lore-lint/analyzers/maplookup"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, maplookup.Analyzer, "a")
}
