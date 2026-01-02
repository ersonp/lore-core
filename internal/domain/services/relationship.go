package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/google/uuid"
)

// RelatedEntity represents an entity connected through relationships.
type RelatedEntity struct {
	EntityID string `json:"entity_id"`
	Depth    int    `json:"depth"`
}

// RelationshipService manages relationships between entities.
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

// Create creates a new relationship between two entities.
// Entities are automatically created if they don't exist.
// Stores the relationship in both SQLite (for graph queries) and Qdrant (for semantic search).
func (s *RelationshipService) Create(
	ctx context.Context,
	worldID string,
	sourceEntityName string,
	relType entities.RelationType,
	targetEntityName string,
	bidirectional bool,
) (*entities.Relationship, error) {
	// Find or create source entity
	sourceEntity, err := s.relationalDB.FindOrCreateEntity(ctx, worldID, sourceEntityName)
	if err != nil {
		return nil, fmt.Errorf("finding/creating source entity: %w", err)
	}

	// Find or create target entity
	targetEntity, err := s.relationalDB.FindOrCreateEntity(ctx, worldID, targetEntityName)
	if err != nil {
		return nil, fmt.Errorf("finding/creating target entity: %w", err)
	}

	// Check for duplicate relationship
	existing, err := s.relationalDB.FindRelationshipBetween(ctx, sourceEntity.ID, targetEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("checking existing relationship: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("relationship already exists between these entities (id: %s)", existing.ID)
	}

	// Create relationship
	rel := &entities.Relationship{
		ID:             uuid.New().String(),
		SourceEntityID: sourceEntity.ID,
		TargetEntityID: targetEntity.ID,
		Type:           relType,
		Bidirectional:  bidirectional,
		CreatedAt:      time.Now(),
	}

	// Save to SQLite for graph queries
	if err := s.relationalDB.SaveRelationship(ctx, rel); err != nil {
		return nil, fmt.Errorf("saving relationship to relational db: %w", err)
	}

	// Create a fact for semantic search
	if err := s.createRelationshipFact(ctx, rel, sourceEntity.Name, targetEntity.Name); err != nil {
		// Rollback SQLite save
		if rollbackErr := s.relationalDB.DeleteRelationship(ctx, rel.ID); rollbackErr != nil {
			log.Printf("warning: failed to rollback relationship %s: %v", rel.ID, rollbackErr)
		}
		return nil, fmt.Errorf("creating relationship fact: %w", err)
	}

	return rel, nil
}

// createRelationshipFact creates a Fact representing the relationship for semantic search.
func (s *RelationshipService) createRelationshipFact(ctx context.Context, rel *entities.Relationship, sourceName, targetName string) error {
	// Build searchable text
	predicate := string(rel.Type)
	searchText := fmt.Sprintf("%s %s %s", sourceName, predicate, targetName)

	// Generate embedding
	embedding, err := s.embedder.Embed(ctx, searchText)
	if err != nil {
		return fmt.Errorf("generating embedding: %w", err)
	}

	// Create fact with relationship type
	fact := &entities.Fact{
		ID:         rel.ID,
		Type:       entities.FactTypeRelationship,
		Subject:    sourceName,
		Predicate:  predicate,
		Object:     targetName,
		Context:    fmt.Sprintf("Relationship between %s and %s", sourceName, targetName),
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

// List returns all relationships for an entity.
func (s *RelationshipService) List(ctx context.Context, entityID string) ([]entities.Relationship, error) {
	return s.relationalDB.FindRelationshipsByEntity(ctx, entityID)
}

// ListByName returns all relationships for an entity by name.
func (s *RelationshipService) ListByName(ctx context.Context, worldID, entityName string) ([]entities.Relationship, error) {
	entity, err := s.relationalDB.FindEntityByName(ctx, worldID, entityName)
	if err != nil {
		return nil, fmt.Errorf("finding entity: %w", err)
	}
	if entity == nil {
		return []entities.Relationship{}, nil
	}
	return s.relationalDB.FindRelationshipsByEntity(ctx, entity.ID)
}

// ListWithDepth returns related entities up to the specified depth.
func (s *RelationshipService) ListWithDepth(ctx context.Context, entityID string, depth int) ([]RelatedEntity, error) {
	if depth < 1 {
		return []RelatedEntity{}, nil
	}

	entityIDs, err := s.relationalDB.FindRelatedEntities(ctx, entityID, depth)
	if err != nil {
		return nil, fmt.Errorf("finding related entities: %w", err)
	}

	// Convert to RelatedEntity structs
	// Note: The current implementation doesn't track depth per entity,
	// so we set depth to 0 (unknown). A more sophisticated implementation
	// could track the actual depth during traversal.
	result := make([]RelatedEntity, len(entityIDs))
	for i, id := range entityIDs {
		result[i] = RelatedEntity{
			EntityID: id,
			Depth:    0,
		}
	}

	return result, nil
}

// FindBetween finds a direct relationship between two entities.
func (s *RelationshipService) FindBetween(ctx context.Context, sourceEntityID, targetEntityID string) (*entities.Relationship, error) {
	return s.relationalDB.FindRelationshipBetween(ctx, sourceEntityID, targetEntityID)
}

// Count returns the total number of relationships.
func (s *RelationshipService) Count(ctx context.Context) (int, error) {
	return s.relationalDB.CountRelationships(ctx)
}
