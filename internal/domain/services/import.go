package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/infrastructure/parsers"
)

// ConflictStrategy defines how to handle existing facts during import.
type ConflictStrategy string

const (
	// ConflictSkip skips facts that already exist (by ID).
	ConflictSkip ConflictStrategy = "skip"
	// ConflictOverwrite overwrites existing facts with new data.
	ConflictOverwrite ConflictStrategy = "overwrite"
)

// ImportOptions controls import behavior.
type ImportOptions struct {
	DryRun     bool             // Validate without saving
	OnConflict ConflictStrategy // How to handle existing facts
}

// ImportError represents an error for a specific fact during import.
type ImportError struct {
	Line    int    // Line number (1-indexed, 0 if unknown)
	Field   string // Which field has the error
	Value   string // The invalid value
	Message string // Human-readable error message
}

func (e ImportError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return e.Message
}

// ImportResult contains the result of an import operation.
type ImportResult struct {
	Imported int
	Skipped  int
	Errors   []ImportError
}

// ImportService handles importing facts from external sources.
type ImportService struct {
	embedder ports.Embedder
	vectorDB ports.VectorDB
}

// NewImportService creates a new import service.
func NewImportService(embedder ports.Embedder, vectorDB ports.VectorDB) *ImportService {
	return &ImportService{
		embedder: embedder,
		vectorDB: vectorDB,
	}
}

// Import validates and imports raw facts into the database.
func (s *ImportService) Import(ctx context.Context, rawFacts []parsers.RawFact, opts ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	// Validate all facts first
	validFacts, validationErrors := s.validateFacts(rawFacts)
	result.Errors = validationErrors

	if len(validFacts) == 0 {
		return result, nil
	}

	// Convert to domain entities
	facts := s.convertToEntities(validFacts)

	// Generate embeddings
	if err := s.generateEmbeddings(ctx, facts); err != nil {
		return nil, fmt.Errorf("generating embeddings: %w", err)
	}

	// Handle dry run
	if opts.DryRun {
		result.Imported = len(facts)
		return result, nil
	}

	// Handle conflicts and save
	imported, skipped, err := s.saveWithConflictHandling(ctx, facts, opts.OnConflict)
	if err != nil {
		return nil, fmt.Errorf("saving facts: %w", err)
	}

	result.Imported = imported
	result.Skipped = skipped

	return result, nil
}

// validateFacts validates raw facts and returns valid ones with any errors.
func (s *ImportService) validateFacts(rawFacts []parsers.RawFact) ([]parsers.RawFact, []ImportError) {
	var valid []parsers.RawFact
	var errors []ImportError

	for i, raw := range rawFacts {
		lineNum := i + 2 // +2 for 1-indexed and header row

		// Check required fields
		if raw.Type == "" {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "type",
				Message: "missing required field: type",
			})
			continue
		}

		if raw.Subject == "" {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "subject",
				Message: "missing required field: subject",
			})
			continue
		}

		if raw.Predicate == "" {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "predicate",
				Message: "missing required field: predicate",
			})
			continue
		}

		if raw.Object == "" {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "object",
				Message: "missing required field: object",
			})
			continue
		}

		// Validate type
		factType := entities.FactType(raw.Type)
		if !factType.IsValid() {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "type",
				Value:   raw.Type,
				Message: fmt.Sprintf("invalid type %q (valid: character, location, event, relationship, rule, timeline)", raw.Type),
			})
			continue
		}

		// Validate confidence range
		if raw.Confidence < 0 || raw.Confidence > 1 {
			errors = append(errors, ImportError{
				Line:    lineNum,
				Field:   "confidence",
				Value:   fmt.Sprintf("%f", raw.Confidence),
				Message: "confidence must be between 0 and 1",
			})
			continue
		}

		valid = append(valid, raw)
	}

	return valid, errors
}

// convertToEntities converts raw facts to domain entities.
func (s *ImportService) convertToEntities(rawFacts []parsers.RawFact) []entities.Fact {
	facts := make([]entities.Fact, 0, len(rawFacts))
	now := time.Now()

	for _, raw := range rawFacts {
		id := raw.ID
		if id == "" {
			id = uuid.New().String()
		}

		confidence := raw.Confidence
		if confidence == 0 {
			confidence = 1.0 // Default confidence
		}

		fact := entities.Fact{
			ID:         id,
			Type:       entities.FactType(raw.Type),
			Subject:    raw.Subject,
			Predicate:  raw.Predicate,
			Object:     raw.Object,
			Context:    raw.Context,
			SourceFile: raw.SourceFile,
			Confidence: confidence,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		facts = append(facts, fact)
	}

	return facts
}

// generateEmbeddings generates embeddings for all facts.
func (s *ImportService) generateEmbeddings(ctx context.Context, facts []entities.Fact) error {
	texts := make([]string, len(facts))
	for i, f := range facts {
		texts[i] = factToText(f)
	}

	embeddings, err := s.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return err
	}

	for i := range facts {
		facts[i].Embedding = embeddings[i]
	}

	return nil
}

// saveWithConflictHandling saves facts with conflict handling.
func (s *ImportService) saveWithConflictHandling(ctx context.Context, facts []entities.Fact, onConflict ConflictStrategy) (imported, skipped int, err error) {
	if onConflict != ConflictSkip {
		// Default: overwrite (upsert)
		if err := s.vectorDB.SaveBatch(ctx, facts); err != nil {
			return 0, 0, err
		}
		return len(facts), 0, nil
	}

	// Skip mode: check each fact for existence
	toSave, skipped := s.filterExisting(ctx, facts)
	if len(toSave) == 0 {
		return 0, skipped, nil
	}

	if err := s.vectorDB.SaveBatch(ctx, toSave); err != nil {
		return 0, 0, err
	}
	return len(toSave), skipped, nil
}

// filterExisting filters out facts that already exist in the database.
func (s *ImportService) filterExisting(ctx context.Context, facts []entities.Fact) ([]entities.Fact, int) {
	var toSave []entities.Fact
	var skipped int

	for _, fact := range facts {
		_, err := s.vectorDB.FindByID(ctx, fact.ID)
		if err != nil {
			// Fact doesn't exist, save it
			toSave = append(toSave, fact)
		} else {
			// Fact exists, skip it
			skipped++
		}
	}

	return toSave, skipped
}
