// Package parsers provides parsers for importing facts from various formats.
package parsers

import (
	"io"
	"path/filepath"
	"strings"
)

// RawFact represents a fact parsed from an external source before validation.
type RawFact struct {
	ID         string   `json:"id,omitempty"`
	Type       string   `json:"type"`
	Subject    string   `json:"subject"`
	Predicate  string   `json:"predicate"`
	Object     string   `json:"object"`
	Context    string   `json:"context,omitempty"`
	SourceFile string   `json:"source_file,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"` // Pointer to distinguish 0 from unset
	LineNum    int      `json:"-"`                    // Line number in source file (set by parser)
}

// Parser defines the interface for parsing facts from various formats.
type Parser interface {
	Parse(r io.Reader) ([]RawFact, error)
}

// ForFormat returns the appropriate parser for the given format.
// Supported formats: "json", "csv".
func ForFormat(format string) Parser {
	switch strings.ToLower(format) {
	case "json":
		return &JSONParser{}
	case "csv":
		return &CSVParser{}
	default:
		return nil
	}
}

// ForFile returns the appropriate parser based on file extension.
func ForFile(filename string) Parser {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return &JSONParser{}
	case ".csv":
		return &CSVParser{}
	default:
		return nil
	}
}
