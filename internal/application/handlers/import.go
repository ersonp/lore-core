package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/parsers"
)

// ImportHandler handles importing facts from files.
type ImportHandler struct {
	service *services.ImportService
}

// NewImportHandler creates a new import handler.
func NewImportHandler(service *services.ImportService) *ImportHandler {
	return &ImportHandler{
		service: service,
	}
}

// ImportOptions controls import behavior.
type ImportOptions struct {
	Format     string                    // "json", "csv", or "auto"
	DryRun     bool                      // Validate without saving
	OnConflict services.ConflictStrategy // How to handle existing facts
}

// ImportResult contains the result of an import operation.
type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []services.ImportError
}

// Handle imports facts from a file.
func (h *ImportHandler) Handle(ctx context.Context, filePath string, opts ImportOptions) (*ImportResult, error) {
	// Get parser
	var parser parsers.Parser
	if opts.Format == "" || opts.Format == "auto" {
		parser = parsers.ForFile(filePath)
	} else {
		parser = parsers.ForFormat(opts.Format)
	}

	if parser == nil {
		return nil, fmt.Errorf("unsupported format for file: %s", filePath)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Parse facts
	rawFacts, err := parser.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	if len(rawFacts) == 0 {
		return &ImportResult{}, nil
	}

	// Import facts
	serviceOpts := services.ImportOptions{
		DryRun:     opts.DryRun,
		OnConflict: opts.OnConflict,
	}

	serviceResult, err := h.service.Import(ctx, rawFacts, serviceOpts)
	if err != nil {
		return nil, err
	}

	return &ImportResult{
		Imported: serviceResult.Imported,
		Skipped:  serviceResult.Skipped,
		Errors:   serviceResult.Errors,
	}, nil
}
