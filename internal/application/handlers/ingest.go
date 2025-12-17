package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// IngestBatchResult contains the result of batch ingestion.
type IngestBatchResult struct {
	TotalFiles  int
	TotalFacts  int
	FileResults []*IngestResult
	Errors      []error
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

// HandleDirectory ingests all matching files in a directory.
func (h *IngestHandler) HandleDirectory(ctx context.Context, dirPath string, pattern string, recursive bool, progressFn func(file string)) (*IngestBatchResult, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("accessing path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	files, err := h.findFiles(absPath, pattern, recursive)
	if err != nil {
		return nil, fmt.Errorf("finding files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files matching pattern %q found in %s", pattern, absPath)
	}

	result := &IngestBatchResult{
		FileResults: make([]*IngestResult, 0, len(files)),
	}

	for _, file := range files {
		if progressFn != nil {
			progressFn(file)
		}

		fileResult, err := h.Handle(ctx, file)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", file, err))
			continue
		}

		result.FileResults = append(result.FileResults, fileResult)
		result.TotalFiles++
		result.TotalFacts += fileResult.FactsCount
	}

	return result, nil
}

// findFiles finds all files matching the pattern in the directory.
func (h *IngestHandler) findFiles(dirPath string, pattern string, recursive bool) ([]string, error) {
	var files []string

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return err
		}

		if matched {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(dirPath, walkFn); err != nil {
		return nil, err
	}

	return files, nil
}

// IsDirectory checks if the given path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsGlobPattern checks if the path contains glob characters.
func IsGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[")
}
