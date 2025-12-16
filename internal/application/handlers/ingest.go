package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
)

// IngestHandler handles file ingestion.
type IngestHandler struct {
	extractionService *services.ExtractionService
}

// NewIngestHandler creates a new ingest handler.
func NewIngestHandler(extractionService *services.ExtractionService) *IngestHandler {
	return &IngestHandler{
		extractionService: extractionService,
	}
}

// IngestResult contains the result of ingestion.
type IngestResult struct {
	FilePath   string
	FactsCount int
	Facts      []entities.Fact
}

// Handle ingests a file and extracts facts.
func (h *IngestHandler) Handle(ctx context.Context, filePath string) (*IngestResult, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("accessing file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", absPath)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	facts, err := h.extractionService.ExtractAndStore(ctx, string(content), absPath)
	if err != nil {
		return nil, fmt.Errorf("extracting facts: %w", err)
	}

	return &IngestResult{
		FilePath:   absPath,
		FactsCount: len(facts),
		Facts:      facts,
	}, nil
}
