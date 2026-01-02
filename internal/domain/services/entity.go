package services

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

// EntityService manages entity operations.
type EntityService struct {
	relationalDB ports.RelationalDB
}

// NewEntityService creates a new EntityService.
func NewEntityService(relationalDB ports.RelationalDB) *EntityService {
	return &EntityService{
		relationalDB: relationalDB,
	}
}

// FindOrCreate finds an entity by name or creates it if not found.
func (s *EntityService) FindOrCreate(ctx context.Context, worldID, name string) (*entities.Entity, error) {
	return s.relationalDB.FindOrCreateEntity(ctx, worldID, name)
}

// FindByName finds an entity by its name (case-insensitive).
func (s *EntityService) FindByName(ctx context.Context, worldID, name string) (*entities.Entity, error) {
	return s.relationalDB.FindEntityByName(ctx, worldID, name)
}

// FindByID finds an entity by its ID.
func (s *EntityService) FindByID(ctx context.Context, entityID string) (*entities.Entity, error) {
	return s.relationalDB.FindEntityByID(ctx, entityID)
}

// List returns all entities for a world with pagination.
func (s *EntityService) List(ctx context.Context, worldID string, limit, offset int) ([]*entities.Entity, error) {
	return s.relationalDB.ListEntities(ctx, worldID, limit, offset)
}

// Search searches entities by name pattern.
func (s *EntityService) Search(ctx context.Context, worldID, query string, limit int) ([]*entities.Entity, error) {
	return s.relationalDB.SearchEntities(ctx, worldID, query, limit)
}

// Delete removes an entity and its relationships.
func (s *EntityService) Delete(ctx context.Context, entityID string) error {
	// First delete all relationships involving this entity
	if err := s.relationalDB.DeleteRelationshipsByEntity(ctx, entityID); err != nil {
		return fmt.Errorf("deleting entity relationships: %w", err)
	}

	// Then delete the entity
	if err := s.relationalDB.DeleteEntity(ctx, entityID); err != nil {
		return fmt.Errorf("deleting entity: %w", err)
	}

	return nil
}

// Count returns the number of entities in a world.
func (s *EntityService) Count(ctx context.Context, worldID string) (int, error) {
	return s.relationalDB.CountEntities(ctx, worldID)
}
