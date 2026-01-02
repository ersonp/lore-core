package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/google/uuid"
)

// RelatedFact represents a fact connected through relationships.
type RelatedFact struct {
	FactID string `json:"fact_id"`
	Depth  int    `json:"depth"`
}

// RelationshipService manages relationships between facts.
type RelationshipService struct {
	vectorDB     ports.VectorDB
	relationalDB ports.RelationalDB
	embedder     ports.Embedder
}

// NewRelationshipService creates a new RelationshipService.
func NewRelationshipService(
	vectorDB ports.VectorDB,
	relationalDB ports.RelationalDB,
	embedder ports.Embedder,
) *RelationshipService {
	return &RelationshipService{
		vectorDB:     vectorDB,
		relationalDB: relationalDB,
		embedder:     embedder,
	}
}

// Create creates a new relationship between two facts.
// It validates both facts exist, checks for duplicates, and stores the relationship
// in both SQLite (for graph queries) and Qdrant (for semantic search).
func (s *RelationshipService) Create(
	ctx context.Context,
	sourceFactID string,
	relType entities.RelationType,
	targetFactID string,
	bidirectional bool,
) (*entities.Relationship, error) {
	// Validate both facts exist
	exists, err := s.vectorDB.ExistsByIDs(ctx, []string{sourceFactID, targetFactID})
	if err != nil {
		return nil, fmt.Errorf("checking facts exist: %w", err)
	}
	if !exists[sourceFactID] {
		return nil, fmt.Errorf("source fact not found: %s", sourceFactID)
	}
	if !exists[targetFactID] {
		return nil, fmt.Errorf("target fact not found: %s", targetFactID)
	}

	// Check for duplicate relationship
	existing, err := s.relationalDB.FindRelationshipBetween(ctx, sourceFactID, targetFactID)
	if err != nil {
		return nil, fmt.Errorf("checking existing relationship: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("relationship already exists between these facts (id: %s)", existing.ID)
	}

	// Create relationship
	rel := &entities.Relationship{
		ID:            uuid.New().String(),
		SourceFactID:  sourceFactID,
		TargetFactID:  targetFactID,
		Type:          relType,
		Bidirectional: bidirectional,
		CreatedAt:     time.Now(),
	}

	// Save to SQLite for graph queries
	if err := s.relationalDB.SaveRelationship(ctx, rel); err != nil {
		return nil, fmt.Errorf("saving relationship to relational db: %w", err)
	}

	// Create a fact for semantic search
	if err := s.createRelationshipFact(ctx, rel); err != nil {
		// Rollback SQLite save
		_ = s.relationalDB.DeleteRelationship(ctx, rel.ID)
		return nil, fmt.Errorf("creating relationship fact: %w", err)
	}

	return rel, nil
}

// createRelationshipFact creates a Fact representing the relationship for semantic search.
func (s *RelationshipService) createRelationshipFact(ctx context.Context, rel *entities.Relationship) error {
	// Fetch source and target facts to build meaningful text
	facts, err := s.vectorDB.FindByIDs(ctx, []string{rel.SourceFactID, rel.TargetFactID})
	if err != nil {
		return fmt.Errorf("fetching facts: %w", err)
	}

	var sourceSubject, targetSubject string
	for i := range facts {
		if facts[i].ID == rel.SourceFactID {
			sourceSubject = facts[i].Subject
		}
		if facts[i].ID == rel.TargetFactID {
			targetSubject = facts[i].Subject
		}
	}

	// Build searchable text
	predicate := string(rel.Type)
	searchText := fmt.Sprintf("%s %s %s", sourceSubject, predicate, targetSubject)

	// Generate embedding
	embedding, err := s.embedder.Embed(ctx, searchText)
	if err != nil {
		return fmt.Errorf("generating embedding: %w", err)
	}

	// Create fact with relationship type
	fact := &entities.Fact{
		ID:         rel.ID,
		Type:       entities.FactTypeRelationship,
		Subject:    sourceSubject,
		Predicate:  predicate,
		Object:     targetSubject,
		Context:    fmt.Sprintf("Relationship between %s and %s", rel.SourceFactID, rel.TargetFactID),
		SourceFile: "relationship",
		Confidence: 1.0,
		Embedding:  embedding,
		CreatedAt:  rel.CreatedAt,
		UpdatedAt:  rel.CreatedAt,
	}

	return s.vectorDB.Save(ctx, fact)
}

// Delete removes a relationship from both SQLite and Qdrant.
func (s *RelationshipService) Delete(ctx context.Context, id string) error {
	// Delete from Qdrant (relationship fact)
	if err := s.vectorDB.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting relationship fact: %w", err)
	}

	// Delete from SQLite
	if err := s.relationalDB.DeleteRelationship(ctx, id); err != nil {
		return fmt.Errorf("deleting relationship: %w", err)
	}

	return nil
}

// List returns all relationships for a fact.
func (s *RelationshipService) List(ctx context.Context, factID string) ([]entities.Relationship, error) {
	return s.relationalDB.FindRelationshipsByFact(ctx, factID)
}

// ListWithDepth returns related facts up to the specified depth.
func (s *RelationshipService) ListWithDepth(ctx context.Context, factID string, depth int) ([]RelatedFact, error) {
	if depth < 1 {
		return []RelatedFact{}, nil
	}

	factIDs, err := s.relationalDB.FindRelatedFacts(ctx, factID, depth)
	if err != nil {
		return nil, fmt.Errorf("finding related facts: %w", err)
	}

	// Convert to RelatedFact structs
	// Note: The current implementation doesn't track depth per fact,
	// so we set depth to 0 (unknown). A more sophisticated implementation
	// could track the actual depth during traversal.
	result := make([]RelatedFact, len(factIDs))
	for i, id := range factIDs {
		result[i] = RelatedFact{
			FactID: id,
			Depth:  0,
		}
	}

	return result, nil
}

// FindBetween finds a direct relationship between two facts.
func (s *RelationshipService) FindBetween(ctx context.Context, sourceID, targetID string) (*entities.Relationship, error) {
	return s.relationalDB.FindRelationshipBetween(ctx, sourceID, targetID)
}

// Count returns the total number of relationships.
func (s *RelationshipService) Count(ctx context.Context) (int, error) {
	return s.relationalDB.CountRelationships(ctx)
}
