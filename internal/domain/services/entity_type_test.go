package services

import (
	"context"
	"sort"
	"testing"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRelationalDB for testing EntityTypeService.
type mockRelationalDB struct {
	types    map[string]*entities.EntityType
	entities map[string]*entities.Entity
}

func newMockRelationalDB() *mockRelationalDB {
	return &mockRelationalDB{
		types:    make(map[string]*entities.EntityType),
		entities: make(map[string]*entities.Entity),
	}
}

// EntityType methods - actual implementations for testing.

func (m *mockRelationalDB) SaveEntityType(_ context.Context, et *entities.EntityType) error {
	m.types[et.Name] = et
	return nil
}

func (m *mockRelationalDB) FindEntityType(_ context.Context, name string) (*entities.EntityType, error) {
	return m.types[name], nil
}

func (m *mockRelationalDB) ListEntityTypes(_ context.Context) ([]entities.EntityType, error) {
	result := make([]entities.EntityType, 0, len(m.types))
	for _, t := range m.types {
		result = append(result, *t)
	}
	// Sort by name for deterministic test results
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (m *mockRelationalDB) DeleteEntityType(_ context.Context, name string) error {
	delete(m.types, name)
	return nil
}

// No-op implementations for other RelationalDB methods.

func (m *mockRelationalDB) EnsureSchema(_ context.Context) error {
	return nil
}

func (m *mockRelationalDB) Close() error {
	return nil
}

// Entity methods.

func (m *mockRelationalDB) SaveEntity(_ context.Context, entity *entities.Entity) error {
	m.entities[entity.ID] = entity
	return nil
}

func (m *mockRelationalDB) FindEntityByName(_ context.Context, _, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	return nil, nil
}

func (m *mockRelationalDB) FindOrCreateEntity(_ context.Context, worldID, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	entity := &entities.Entity{
		ID:             "entity-" + name,
		WorldID:        worldID,
		Name:           name,
		NormalizedName: normalizedName,
	}
	m.entities[entity.ID] = entity
	return entity, nil
}

func (m *mockRelationalDB) FindEntityByID(_ context.Context, entityID string) (*entities.Entity, error) {
	return m.entities[entityID], nil
}

func (m *mockRelationalDB) ListEntities(_ context.Context, _ string, _, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *mockRelationalDB) SearchEntities(_ context.Context, _, _ string, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *mockRelationalDB) DeleteEntity(_ context.Context, entityID string) error {
	delete(m.entities, entityID)
	return nil
}

func (m *mockRelationalDB) CountEntities(_ context.Context, _ string) (int, error) {
	return len(m.entities), nil
}

// Relationship methods.

func (m *mockRelationalDB) SaveRelationship(_ context.Context, _ *entities.Relationship) error {
	return nil
}

func (m *mockRelationalDB) FindRelationshipsByEntity(_ context.Context, _ string) ([]entities.Relationship, error) {
	return nil, nil
}

func (m *mockRelationalDB) FindRelationshipsByType(_ context.Context, _ string) ([]entities.Relationship, error) {
	return nil, nil
}

func (m *mockRelationalDB) DeleteRelationship(_ context.Context, _ string) error {
	return nil
}

func (m *mockRelationalDB) DeleteRelationshipsByEntity(_ context.Context, _ string) error {
	return nil
}

func (m *mockRelationalDB) FindRelationshipBetween(_ context.Context, _, _ string) (*entities.Relationship, error) {
	return nil, nil
}

func (m *mockRelationalDB) FindRelatedEntities(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, nil
}

func (m *mockRelationalDB) CountRelationships(_ context.Context) (int, error) {
	return 0, nil
}

// Version methods.

func (m *mockRelationalDB) SaveVersion(_ context.Context, _ *entities.FactVersion) error {
	return nil
}

func (m *mockRelationalDB) FindVersionsByFact(_ context.Context, _ string) ([]entities.FactVersion, error) {
	return nil, nil
}

func (m *mockRelationalDB) FindLatestVersion(_ context.Context, _ string) (*entities.FactVersion, error) {
	return nil, nil
}

func (m *mockRelationalDB) CountVersions(_ context.Context, _ string) (int, error) {
	return 0, nil
}

// Audit log methods.

func (m *mockRelationalDB) LogAction(_ context.Context, _ string, _ string, _ map[string]any) error {
	return nil
}

func (m *mockRelationalDB) FindAuditLog(_ context.Context, _ string) ([]entities.AuditEntry, error) {
	return nil, nil
}

func (m *mockRelationalDB) FindAuditLogByAction(_ context.Context, _ string, _ int) ([]entities.AuditEntry, error) {
	return nil, nil
}

// Tests

func TestEntityTypeService_LoadDefaults(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)
}

func TestEntityTypeService_LoadDefaults_Idempotent(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	// Load twice
	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)
	err = svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6) // Still 6, not 12
}

func TestEntityTypeService_Add(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.Add(context.Background(), "weapon", "Weapons and artifacts")
	require.NoError(t, err)

	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 1)
	assert.Equal(t, "weapon", types[0].Name)
}

func TestEntityTypeService_Add_InvalidName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "weapon", false},
		{"valid with underscore", "magic_item", false},
		{"valid with number", "type2", false},
		{"uppercase normalized to lowercase", "Weapon", false}, // normalized to "weapon"
		{"invalid starts with number", "2type", true},
		{"invalid special chars", "weapon!", true},
		{"invalid spaces", "magic item", true},
		{"invalid hyphen", "magic-item", true},
		{"invalid empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newMockRelationalDB()
			svc := NewEntityTypeService(db)

			err := svc.Add(context.Background(), tt.input, "description")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEntityTypeService_Add_NormalizesName(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	// Input with leading/trailing spaces and uppercase
	err := svc.Add(context.Background(), "  Weapon  ", "Weapons")
	require.NoError(t, err)

	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 1)
	assert.Equal(t, "weapon", types[0].Name)
}

func TestEntityTypeService_Add_Duplicate(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.Add(context.Background(), "weapon", "First")
	require.NoError(t, err)

	err = svc.Add(context.Background(), "weapon", "Second")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEntityTypeService_Remove(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.Add(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	err = svc.Remove(context.Background(), "weapon")
	require.NoError(t, err)

	types, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 0)
}

func TestEntityTypeService_Remove_DefaultType(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	err = svc.Remove(context.Background(), "character")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove default")
}

func TestEntityTypeService_Remove_NotFound(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.Remove(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEntityTypeService_IsValid(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	assert.True(t, svc.IsValid(context.Background(), "character"))
	assert.True(t, svc.IsValid(context.Background(), "location"))
	assert.False(t, svc.IsValid(context.Background(), "weapon"))
	assert.False(t, svc.IsValid(context.Background(), "nonexistent"))
}

func TestEntityTypeService_IsValid_CacheInvalidation(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	// Initially weapon doesn't exist
	assert.False(t, svc.IsValid(context.Background(), "weapon"))

	// Add weapon
	err := svc.Add(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	// Now weapon should be valid (cache should be invalidated)
	assert.True(t, svc.IsValid(context.Background(), "weapon"))
}

func TestEntityTypeService_GetValidTypes(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)

	types, err := svc.GetValidTypes(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)
	assert.Contains(t, types, "character")
	assert.Contains(t, types, "location")
	assert.Contains(t, types, "event")
	assert.Contains(t, types, "relationship")
	assert.Contains(t, types, "rule")
	assert.Contains(t, types, "timeline")
}

func TestEntityTypeService_GetValidTypes_Empty(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	types, err := svc.GetValidTypes(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 0)
}

func TestEntityTypeService_BuildPromptTypeList(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	err := svc.Add(context.Background(), "character", "Characters")
	require.NoError(t, err)
	err = svc.Add(context.Background(), "location", "Locations")
	require.NoError(t, err)

	list, err := svc.BuildPromptTypeList(context.Background())
	require.NoError(t, err)
	assert.Contains(t, list, "character")
	assert.Contains(t, list, "location")
	assert.Contains(t, list, ", ")
}

func TestEntityTypeService_BuildPromptTypeList_Empty(t *testing.T) {
	db := newMockRelationalDB()
	svc := NewEntityTypeService(db)

	list, err := svc.BuildPromptTypeList(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "", list)
}
