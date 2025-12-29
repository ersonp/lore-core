package main

// Default limits for CLI commands.
const (
	DefaultQueryLimit  = 10
	DefaultListLimit   = 50
	DefaultExportLimit = 1000
	MaxDeleteBatchSize = 1000
)

// Valid export formats.
var validFormats = []string{"json", "csv", "markdown"}
