package mocks

import (
	"context"
	"sort"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// RelationalDB is a mock implementation of ports.RelationalDB.
type RelationalDB struct {
	Types    map[string]*entities.EntityType
	Entities map[string]*entities.Entity
	Err      error
}

// NewRelationalDB creates a new mock RelationalDB.
func NewRelationalDB() *RelationalDB {
	return &RelationalDB{
		Types:    make(map[string]*entities.EntityType),
		Entities: make(map[string]*entities.Entity),
	}
}

// EnsureSchema creates the database schema if it doesn't exist.
func (m *RelationalDB) EnsureSchema(_ context.Context) error {
	return m.Err
}

// Close closes the database connection.
func (m *RelationalDB) Close() error {
	return nil
}

// Entity methods.

// SaveEntity saves or updates an entity.
func (m *RelationalDB) SaveEntity(_ context.Context, entity *entities.Entity) error {
	if m.Err != nil {
		return m.Err
	}
	m.Entities[entity.ID] = entity
	return nil
}

// FindEntityByName finds an entity by its normalized name.
func (m *RelationalDB) FindEntityByName(_ context.Context, worldID, name string) (*entities.Entity, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.Entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	return nil, nil
}

// FindOrCreateEntity finds an entity by name or creates it if not found.
func (m *RelationalDB) FindOrCreateEntity(_ context.Context, worldID, name string) (*entities.Entity, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.Entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	// Create new entity
	entity := &entities.Entity{
		ID:             "entity-" + name,
		WorldID:        worldID,
		Name:           name,
		NormalizedName: normalizedName,
	}
	m.Entities[entity.ID] = entity
	return entity, nil
}

// FindEntityByID finds an entity by its ID.
func (m *RelationalDB) FindEntityByID(_ context.Context, entityID string) (*entities.Entity, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Entities[entityID], nil
}

// ListEntities lists all entities for a world with pagination.
func (m *RelationalDB) ListEntities(_ context.Context, worldID string, limit, offset int) ([]*entities.Entity, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	var result []*entities.Entity
	for _, e := range m.Entities {
		if e.WorldID == worldID {
			result = append(result, e)
		}
	}
	return result, nil
}

// SearchEntities searches entities by name pattern.
func (m *RelationalDB) SearchEntities(_ context.Context, worldID, query string, limit int) ([]*entities.Entity, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return nil, nil
}

// DeleteEntity deletes an entity by ID.
func (m *RelationalDB) DeleteEntity(_ context.Context, entityID string) error {
	if m.Err != nil {
		return m.Err
	}
	delete(m.Entities, entityID)
	return nil
}

// CountEntities returns the total number of entities for a world.
func (m *RelationalDB) CountEntities(_ context.Context, worldID string) (int, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	count := 0
	for _, e := range m.Entities {
		if e.WorldID == worldID {
			count++
		}
	}
	return count, nil
}

// Entity type methods.

// SaveEntityType saves or updates a custom entity type.
func (m *RelationalDB) SaveEntityType(_ context.Context, et *entities.EntityType) error {
	if m.Err != nil {
		return m.Err
	}
	m.Types[et.Name] = et
	return nil
}

// FindEntityType finds a custom entity type by name.
func (m *RelationalDB) FindEntityType(_ context.Context, name string) (*entities.EntityType, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Types[name], nil
}

// ListEntityTypes lists all custom entity types.
func (m *RelationalDB) ListEntityTypes(_ context.Context) ([]entities.EntityType, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	result := make([]entities.EntityType, 0, len(m.Types))
	for _, t := range m.Types {
		result = append(result, *t)
	}
	// Sort by name for deterministic test results
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// DeleteEntityType deletes a custom entity type by name.
func (m *RelationalDB) DeleteEntityType(_ context.Context, name string) error {
	if m.Err != nil {
		return m.Err
	}
	delete(m.Types, name)
	return nil
}

// Relationship methods - no-op implementations.

// SaveRelationship saves or updates a relationship.
func (m *RelationalDB) SaveRelationship(_ context.Context, _ *entities.Relationship) error {
	return m.Err
}

// FindRelationshipsByEntity finds all relationships involving an entity.
func (m *RelationalDB) FindRelationshipsByEntity(_ context.Context, _ string) ([]entities.Relationship, error) {
	return nil, m.Err
}

// FindRelationshipsByType finds all relationships of a given type.
func (m *RelationalDB) FindRelationshipsByType(_ context.Context, _ string) ([]entities.Relationship, error) {
	return nil, m.Err
}

// DeleteRelationship deletes a relationship by ID.
func (m *RelationalDB) DeleteRelationship(_ context.Context, _ string) error {
	return m.Err
}

// DeleteRelationshipsByEntity deletes all relationships involving an entity.
func (m *RelationalDB) DeleteRelationshipsByEntity(_ context.Context, _ string) error {
	return m.Err
}

// FindRelationshipBetween finds a direct relationship between two entities.
func (m *RelationalDB) FindRelationshipBetween(_ context.Context, _, _ string) (*entities.Relationship, error) {
	return nil, m.Err
}

// FindRelatedEntities finds all entity IDs connected to the given entity up to the specified depth.
func (m *RelationalDB) FindRelatedEntities(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, m.Err
}

// CountRelationships returns the total number of relationships in the database.
func (m *RelationalDB) CountRelationships(_ context.Context) (int, error) {
	return 0, m.Err
}

// Version methods - no-op implementations.

// SaveVersion saves a new fact version.
func (m *RelationalDB) SaveVersion(_ context.Context, _ *entities.FactVersion) error {
	return m.Err
}

// FindVersionsByFact finds all versions of a fact.
func (m *RelationalDB) FindVersionsByFact(_ context.Context, _ string) ([]entities.FactVersion, error) {
	return nil, m.Err
}

// FindLatestVersion finds the most recent version of a fact.
func (m *RelationalDB) FindLatestVersion(_ context.Context, _ string) (*entities.FactVersion, error) {
	return nil, m.Err
}

// CountVersions counts how many versions a fact has.
func (m *RelationalDB) CountVersions(_ context.Context, _ string) (int, error) {
	return 0, m.Err
}

// Audit log methods - no-op implementations.

// LogAction logs an action to the audit log.
func (m *RelationalDB) LogAction(_ context.Context, _ string, _ string, _ map[string]any) error {
	return m.Err
}

// FindAuditLog finds audit log entries for a specific fact.
func (m *RelationalDB) FindAuditLog(_ context.Context, _ string) ([]entities.AuditEntry, error) {
	return nil, m.Err
}

// FindAuditLogByAction finds audit log entries by action type.
func (m *RelationalDB) FindAuditLogByAction(_ context.Context, _ string, _ int) ([]entities.AuditEntry, error) {
	return nil, m.Err
}
