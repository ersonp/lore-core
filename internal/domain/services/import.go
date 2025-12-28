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
	valid := make([]parsers.RawFact, 0, len(rawFacts))
	var errors []ImportError

	for i := range rawFacts {
		raw := &rawFacts[i]
		lineNum := raw.LineNum
		if lineNum == 0 {
			lineNum = i + 1
		}

		if err := validateRawFact(raw, lineNum); err != nil {
			errors = append(errors, *err)
			continue
		}

		valid = append(valid, *raw)
	}

	return valid, errors
}

// validateRawFact validates a single raw fact and returns an error if invalid.
func validateRawFact(raw *parsers.RawFact, lineNum int) *ImportError {
	if raw.Type == "" {
		return &ImportError{Line: lineNum, Field: "type", Message: "missing required field: type"}
	}
	if raw.Subject == "" {
		return &ImportError{Line: lineNum, Field: "subject", Message: "missing required field: subject"}
	}
	if raw.Predicate == "" {
		return &ImportError{Line: lineNum, Field: "predicate", Message: "missing required field: predicate"}
	}
	if raw.Object == "" {
		return &ImportError{Line: lineNum, Field: "object", Message: "missing required field: object"}
	}

	factType := entities.FactType(raw.Type)
	if !factType.IsValid() {
		return &ImportError{
			Line:    lineNum,
			Field:   "type",
			Value:   raw.Type,
			Message: fmt.Sprintf("invalid type %q (valid: character, location, event, relationship, rule, timeline)", raw.Type),
		}
	}

	if raw.Confidence != nil && (*raw.Confidence < 0 || *raw.Confidence > 1) {
		return &ImportError{
			Line:    lineNum,
			Field:   "confidence",
			Value:   fmt.Sprintf("%f", *raw.Confidence),
			Message: "confidence must be between 0 and 1",
		}
	}

	return nil
}

// convertToEntities converts raw facts to domain entities.
func (s *ImportService) convertToEntities(rawFacts []parsers.RawFact) []entities.Fact {
	facts := make([]entities.Fact, 0, len(rawFacts))
	now := time.Now()

	for i := range rawFacts {
		raw := &rawFacts[i]
		id := raw.ID
		if id == "" {
			id = uuid.New().String()
		}

		// Use provided confidence, or default to 1.0 if not set
		confidence := 1.0
		if raw.Confidence != nil {
			confidence = *raw.Confidence
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
	for i := range facts {
		texts[i] = factToText(&facts[i])
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
		// Overwrite mode: preserve CreatedAt for existing facts
		if err := s.preserveCreatedAt(ctx, facts); err != nil {
			return 0, 0, err
		}
		if err := s.vectorDB.SaveBatch(ctx, facts); err != nil {
			return 0, 0, err
		}
		return len(facts), 0, nil
	}

	// Skip mode: check each fact for existence
	toSave, skipped, err := s.filterExisting(ctx, facts)
	if err != nil {
		return 0, 0, err
	}
	if len(toSave) == 0 {
		return 0, skipped, nil
	}

	if err := s.vectorDB.SaveBatch(ctx, toSave); err != nil {
		return 0, 0, err
	}
	return len(toSave), skipped, nil
}

// preserveCreatedAt looks up existing facts and preserves their CreatedAt timestamps.
func (s *ImportService) preserveCreatedAt(ctx context.Context, facts []entities.Fact) error {
	if len(facts) == 0 {
		return nil
	}

	// Collect IDs to look up
	ids := make([]string, len(facts))
	for i := range facts {
		ids[i] = facts[i].ID
	}

	// Get existing facts' timestamps
	createdAtMap, err := s.getExistingCreatedAt(ctx, ids)
	if err != nil {
		return err
	}

	// Preserve CreatedAt for existing facts
	for i := range facts {
		if createdAt, exists := createdAtMap[facts[i].ID]; exists {
			facts[i].CreatedAt = createdAt
		}
	}

	return nil
}

// getExistingCreatedAt retrieves CreatedAt timestamps for existing fact IDs.
func (s *ImportService) getExistingCreatedAt(ctx context.Context, ids []string) (map[string]time.Time, error) {
	existingFacts, err := s.vectorDB.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("looking up existing facts: %w", err)
	}

	createdAtMap := make(map[string]time.Time, len(existingFacts))
	for i := range existingFacts {
		createdAtMap[existingFacts[i].ID] = existingFacts[i].CreatedAt
	}

	return createdAtMap, nil
}

// filterExisting filters out facts that already exist in the database.
func (s *ImportService) filterExisting(ctx context.Context, facts []entities.Fact) ([]entities.Fact, int, error) {
	if len(facts) == 0 {
		return nil, 0, nil
	}

	// Collect all IDs for batch lookup
	ids := make([]string, len(facts))
	for i := range facts {
		ids[i] = facts[i].ID
	}

	// Single batch query instead of N queries
	exists, err := s.vectorDB.ExistsByIDs(ctx, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("checking existing facts: %w", err)
	}

	// Filter out existing facts
	toSave := make([]entities.Fact, 0, len(facts))
	var skipped int
	for i := range facts {
		if exists[facts[i].ID] {
			skipped++
		} else {
			toSave = append(toSave, facts[i])
		}
	}

	return toSave, skipped, nil
}
