package main

// Default limits for CLI commands.
const (
	DefaultQueryLimit  = 10
	DefaultListLimit   = 50
	DefaultExportLimit = 1000
	MaxDeleteBatchSize = 1000
)

// Valid fact types for filtering.
var validTypes = []string{"character", "location", "event", "relationship", "rule", "timeline"}

// Valid export formats.
var validFormats = []string{"json", "csv", "markdown"}
