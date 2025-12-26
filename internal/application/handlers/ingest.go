package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
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

// IngestOptions controls ingestion behavior.
type IngestOptions struct {
	CheckConsistency bool // Check for contradictions with existing facts
	CheckOnly        bool // Only check, don't save facts
}

// IngestResult contains the result of ingestion.
type IngestResult struct {
	FilePath   string
	FactsCount int
	Facts      []entities.Fact
	Issues     []ports.ConsistencyIssue
}

// IngestBatchResult contains the result of batch ingestion.
type IngestBatchResult struct {
	TotalFiles  int
	TotalFacts  int
	TotalIssues int
	FileResults []*IngestResult
	Errors      []error
}

// Handle ingests a file and extracts facts.
func (h *IngestHandler) Handle(ctx context.Context, filePath string) (*IngestResult, error) {
	return h.HandleWithOptions(ctx, filePath, IngestOptions{})
}

// HandleWithOptions ingests a file with consistency checking options.
// Uses streaming to avoid loading entire file into memory.
func (h *IngestHandler) HandleWithOptions(ctx context.Context, filePath string, opts IngestOptions) (*IngestResult, error) {
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

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	extractOpts := services.ExtractionOptions{
		CheckConsistency: opts.CheckConsistency,
		CheckOnly:        opts.CheckOnly,
	}

	result, err := h.extractionService.ExtractFromReader(ctx, file, absPath, extractOpts)
	if err != nil {
		return nil, fmt.Errorf("extracting facts: %w", err)
	}

	return &IngestResult{
		FilePath:   absPath,
		FactsCount: len(result.Facts),
		Facts:      result.Facts,
		Issues:     result.Issues,
	}, nil
}

// HandleDirectory ingests all matching files in a directory.
func (h *IngestHandler) HandleDirectory(ctx context.Context, dirPath string, pattern string, recursive bool, progressFn func(file string)) (*IngestBatchResult, error) {
	return h.HandleDirectoryWithOptions(ctx, dirPath, pattern, recursive, progressFn, IngestOptions{})
}

// HandleDirectoryWithOptions ingests all matching files with consistency checking options.
func (h *IngestHandler) HandleDirectoryWithOptions(ctx context.Context, dirPath string, pattern string, recursive bool, progressFn func(file string), opts IngestOptions) (*IngestBatchResult, error) {
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

		fileResult, err := h.HandleWithOptions(ctx, file, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", file, err))
			continue
		}

		result.FileResults = append(result.FileResults, fileResult)
		result.TotalFiles++
		result.TotalFacts += fileResult.FactsCount
		result.TotalIssues += len(fileResult.Issues)
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
