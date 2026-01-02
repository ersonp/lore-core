package mocks

import (
	"context"
	"sort"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// RelationalDB is a mock implementation of ports.RelationalDB.
type RelationalDB struct {
	Types map[string]*entities.EntityType
	Err   error
}

// NewRelationalDB creates a new mock RelationalDB.
func NewRelationalDB() *RelationalDB {
	return &RelationalDB{
		Types: make(map[string]*entities.EntityType),
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

// FindRelationshipsByFact finds all relationships involving a fact.
func (m *RelationalDB) FindRelationshipsByFact(_ context.Context, _ string) ([]entities.Relationship, error) {
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

// DeleteRelationshipsByFact deletes all relationships involving a fact.
func (m *RelationalDB) DeleteRelationshipsByFact(_ context.Context, _ string) error {
	return m.Err
}

// FindRelationshipBetween finds a direct relationship between two facts.
func (m *RelationalDB) FindRelationshipBetween(_ context.Context, _, _ string) (*entities.Relationship, error) {
	return nil, m.Err
}

// FindRelatedFacts finds all fact IDs connected to the given fact up to the specified depth.
func (m *RelationalDB) FindRelatedFacts(_ context.Context, _ string, _ int) ([]string, error) {
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
