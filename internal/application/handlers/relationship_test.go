package handlers

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// relHandlerVectorDB is a test mock for VectorDB.
type relHandlerVectorDB struct {
	facts map[string]entities.Fact
}

func newRelHandlerVectorDB() *relHandlerVectorDB {
	return &relHandlerVectorDB{facts: make(map[string]entities.Fact)}
}

func (m *relHandlerVectorDB) EnsureCollection(_ context.Context, _ uint64) error { return nil }
func (m *relHandlerVectorDB) DeleteCollection(_ context.Context) error           { return nil }

func (m *relHandlerVectorDB) Save(_ context.Context, fact *entities.Fact) error {
	m.facts[fact.ID] = *fact
	return nil
}

func (m *relHandlerVectorDB) SaveBatch(_ context.Context, facts []entities.Fact) error {
	for i := range facts {
		m.facts[facts[i].ID] = facts[i]
	}
	return nil
}

func (m *relHandlerVectorDB) FindByID(_ context.Context, id string) (entities.Fact, error) {
	if f, ok := m.facts[id]; ok {
		return f, nil
	}
	return entities.Fact{}, nil
}

func (m *relHandlerVectorDB) ExistsByIDs(_ context.Context, ids []string) (map[string]bool, error) {
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		_, exists := m.facts[id]
		result[id] = exists
	}
	return result, nil
}

func (m *relHandlerVectorDB) FindByIDs(_ context.Context, ids []string) ([]entities.Fact, error) {
	result := make([]entities.Fact, 0, len(ids))
	for _, id := range ids {
		if f, ok := m.facts[id]; ok {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *relHandlerVectorDB) Search(_ context.Context, _ []float32, _ int) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relHandlerVectorDB) SearchByType(_ context.Context, _ []float32, _ entities.FactType, _ int) ([]entities.Fact, error) {
	return nil, nil
}

func (m *relHandlerVectorDB) Delete(_ context.Context, id string) error {
	delete(m.facts, id)
	return nil
}

func (m *relHandlerVectorDB) List(_ context.Context, _ int, _ uint64) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relHandlerVectorDB) ListByType(_ context.Context, _ entities.FactType, _ int) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relHandlerVectorDB) ListBySource(_ context.Context, _ string, _ int) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relHandlerVectorDB) DeleteBySource(_ context.Context, _ string) error { return nil }
func (m *relHandlerVectorDB) DeleteAll(_ context.Context) error                { return nil }
func (m *relHandlerVectorDB) Count(_ context.Context) (uint64, error)          { return 0, nil }

// relHandlerRelationalDB is a test mock for RelationalDB.
type relHandlerRelationalDB struct {
	entities      map[string]*entities.Entity
	relationships map[string]*entities.Relationship
}

func newRelHandlerRelationalDB() *relHandlerRelationalDB {
	return &relHandlerRelationalDB{
		entities:      make(map[string]*entities.Entity),
		relationships: make(map[string]*entities.Relationship),
	}
}

func (m *relHandlerRelationalDB) EnsureSchema(_ context.Context) error { return nil }
func (m *relHandlerRelationalDB) Close() error                         { return nil }

// Entity methods.

func (m *relHandlerRelationalDB) SaveEntity(_ context.Context, entity *entities.Entity) error {
	m.entities[entity.ID] = entity
	return nil
}

func (m *relHandlerRelationalDB) FindEntityByName(_ context.Context, worldID, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	return nil, nil
}

func (m *relHandlerRelationalDB) FindOrCreateEntity(_ context.Context, worldID, name string) (*entities.Entity, error) {
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	entity := &entities.Entity{
		ID:             "entity-" + normalizedName,
		WorldID:        worldID,
		Name:           name,
		NormalizedName: normalizedName,
		CreatedAt:      time.Now(),
	}
	m.entities[entity.ID] = entity
	return entity, nil
}

func (m *relHandlerRelationalDB) FindEntityByID(_ context.Context, entityID string) (*entities.Entity, error) {
	return m.entities[entityID], nil
}

func (m *relHandlerRelationalDB) FindEntitiesByIDs(_ context.Context, ids []string) ([]*entities.Entity, error) {
	result := make([]*entities.Entity, 0, len(ids))
	for _, id := range ids {
		if e, ok := m.entities[id]; ok {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *relHandlerRelationalDB) ListEntities(_ context.Context, _ string, _, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *relHandlerRelationalDB) SearchEntities(_ context.Context, _, _ string, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *relHandlerRelationalDB) DeleteEntity(_ context.Context, entityID string) error {
	delete(m.entities, entityID)
	return nil
}

func (m *relHandlerRelationalDB) CountEntities(_ context.Context, _ string) (int, error) {
	return len(m.entities), nil
}

// Relationship methods.

func (m *relHandlerRelationalDB) SaveRelationship(_ context.Context, rel *entities.Relationship) error {
	m.relationships[rel.ID] = rel
	return nil
}

func (m *relHandlerRelationalDB) FindRelationshipsByEntity(_ context.Context, entityID string) ([]entities.Relationship, error) {
	var result []entities.Relationship
	for _, rel := range m.relationships {
		if rel.SourceEntityID == entityID || (rel.TargetEntityID == entityID && rel.Bidirectional) {
			result = append(result, *rel)
		}
	}
	// Sort for deterministic test results
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (m *relHandlerRelationalDB) FindRelationshipsByType(_ context.Context, relType string) ([]entities.Relationship, error) {
	var result []entities.Relationship
	for _, rel := range m.relationships {
		if string(rel.Type) == relType {
			result = append(result, *rel)
		}
	}
	return result, nil
}

func (m *relHandlerRelationalDB) DeleteRelationship(_ context.Context, id string) error {
	delete(m.relationships, id)
	return nil
}

func (m *relHandlerRelationalDB) DeleteRelationshipsByEntity(_ context.Context, _ string) error {
	return nil
}

func (m *relHandlerRelationalDB) FindRelationshipBetween(_ context.Context, sourceID, targetID string) (*entities.Relationship, error) {
	for _, rel := range m.relationships {
		if rel.SourceEntityID == sourceID && rel.TargetEntityID == targetID {
			return rel, nil
		}
		if rel.Bidirectional && rel.SourceEntityID == targetID && rel.TargetEntityID == sourceID {
			return rel, nil
		}
	}
	return nil, nil
}

func (m *relHandlerRelationalDB) FindRelatedEntities(_ context.Context, entityID string, _ int) ([]string, error) {
	seen := make(map[string]bool)
	for _, rel := range m.relationships {
		if rel.SourceEntityID == entityID {
			seen[rel.TargetEntityID] = true
		}
		if rel.Bidirectional && rel.TargetEntityID == entityID {
			seen[rel.SourceEntityID] = true
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result, nil
}

func (m *relHandlerRelationalDB) CountRelationships(_ context.Context) (int, error) {
	return len(m.relationships), nil
}

// No-op implementations for other methods.
func (m *relHandlerRelationalDB) SaveEntityType(_ context.Context, _ *entities.EntityType) error {
	return nil
}
func (m *relHandlerRelationalDB) FindEntityType(_ context.Context, _ string) (*entities.EntityType, error) {
	return nil, nil
}
func (m *relHandlerRelationalDB) ListEntityTypes(_ context.Context) ([]entities.EntityType, error) {
	return nil, nil
}
func (m *relHandlerRelationalDB) DeleteEntityType(_ context.Context, _ string) error { return nil }
func (m *relHandlerRelationalDB) SaveVersion(_ context.Context, _ *entities.FactVersion) error {
	return nil
}
func (m *relHandlerRelationalDB) FindVersionsByFact(_ context.Context, _ string) ([]entities.FactVersion, error) {
	return nil, nil
}
func (m *relHandlerRelationalDB) FindLatestVersion(_ context.Context, _ string) (*entities.FactVersion, error) {
	return nil, nil
}
func (m *relHandlerRelationalDB) CountVersions(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *relHandlerRelationalDB) LogAction(_ context.Context, _ string, _ string, _ map[string]any) error {
	return nil
}
func (m *relHandlerRelationalDB) FindAuditLog(_ context.Context, _ string) ([]entities.AuditEntry, error) {
	return nil, nil
}
func (m *relHandlerRelationalDB) FindAuditLogByAction(_ context.Context, _ string, _ int) ([]entities.AuditEntry, error) {
	return nil, nil
}

// relHandlerEmbedder is a test mock for Embedder.
type relHandlerEmbedder struct{}

func (m *relHandlerEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *relHandlerEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = []float32{0.1, 0.2, 0.3}
	}
	return result, nil
}

// Test setup
func setupRelationshipHandlerTest() (*RelationshipHandler, *relHandlerVectorDB, *relHandlerRelationalDB) {
	vectorDB := newRelHandlerVectorDB()
	relationalDB := newRelHandlerRelationalDB()
	embedder := &relHandlerEmbedder{}

	svc := services.NewRelationshipService(vectorDB, relationalDB, embedder)
	handler := NewRelationshipHandler(svc, relationalDB)

	return handler, vectorDB, relationalDB
}

func TestRelationshipHandler_HandleCreate(t *testing.T) {
	const worldID = "test-world"

	t.Run("successful creation", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		rel, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)
		require.NotNil(t, rel)

		assert.NotEmpty(t, rel.SourceEntityID)
		assert.NotEmpty(t, rel.TargetEntityID)
		assert.Equal(t, entities.RelationAlly, rel.Type)
		assert.True(t, rel.Bidirectional)
	})

	t.Run("invalid relationship type", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		_, err := handler.HandleCreate(ctx, worldID, "Alice", "invalid_type", "Bob", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid relationship type")
	})

	t.Run("all valid relationship types", func(t *testing.T) {
		validTypes := []string{
			"parent", "child", "sibling", "spouse",
			"ally", "enemy", "located_in", "owns", "member_of", "created",
		}

		for _, relType := range validTypes {
			handler, _, _ := setupRelationshipHandlerTest()
			ctx := context.Background()

			rel, err := handler.HandleCreate(ctx, worldID, "Alice", relType, "Bob", false)
			require.NoError(t, err, "type %s should be valid", relType)
			assert.Equal(t, entities.RelationType(relType), rel.Type)
		}
	})
}

func TestRelationshipHandler_HandleDelete(t *testing.T) {
	const worldID = "test-world"

	t.Run("successful deletion", func(t *testing.T) {
		handler, _, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Create relationship first
		rel, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)

		// Verify it exists
		assert.Len(t, relationalDB.relationships, 1)

		// Delete it
		err = handler.HandleDelete(ctx, rel.ID)
		require.NoError(t, err)

		assert.Len(t, relationalDB.relationships, 0)
	})
}

func TestRelationshipHandler_HandleList(t *testing.T) {
	const worldID = "test-world"

	t.Run("returns relationships with entity details", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Create relationship (entities auto-created)
		_, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)

		result, err := handler.HandleList(ctx, worldID, "Alice", ListOptions{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Relationships, 1)

		info := result.Relationships[0]
		require.NotNil(t, info.SourceEntity)
		require.NotNil(t, info.TargetEntity)
		assert.Equal(t, "Alice", info.SourceEntity.Name)
		assert.Equal(t, "Bob", info.TargetEntity.Name)
	})

	t.Run("filters by type", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Create relationships of different types
		_, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)

		_, err = handler.HandleCreate(ctx, worldID, "Alice", "located_in", "City", false)
		require.NoError(t, err)

		// Filter by ally
		result, err := handler.HandleList(ctx, worldID, "Alice", ListOptions{Type: "ally"})
		require.NoError(t, err)
		assert.Len(t, result.Relationships, 1)
		assert.Equal(t, entities.RelationAlly, result.Relationships[0].Relationship.Type)
	})

	t.Run("empty results", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		result, err := handler.HandleList(ctx, worldID, "NonexistentEntity", ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Relationships)
	})
}

func TestRelationshipHandler_HandleFindBetween(t *testing.T) {
	const worldID = "test-world"

	t.Run("finds relationship", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		rel, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)

		found, err := handler.HandleFindBetween(ctx, rel.SourceEntityID, rel.TargetEntityID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, rel.ID, found.ID)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		rel, err := handler.HandleFindBetween(ctx, "nonexistent-1", "nonexistent-2")
		require.NoError(t, err)
		assert.Nil(t, rel)
	})
}

func TestRelationshipHandler_HandleCount(t *testing.T) {
	const worldID = "test-world"

	t.Run("returns count", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Create relationships
		_, err := handler.HandleCreate(ctx, worldID, "Alice", "ally", "Bob", true)
		require.NoError(t, err)

		_, err = handler.HandleCreate(ctx, worldID, "Alice", "enemy", "Eve", true)
		require.NoError(t, err)

		count, err := handler.HandleCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
