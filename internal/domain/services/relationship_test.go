package services

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// relTestVectorDB is a test mock for VectorDB.
type relTestVectorDB struct {
	facts     map[string]entities.Fact
	saveErr   error
	deleteErr error
}

func newRelTestVectorDB() *relTestVectorDB {
	return &relTestVectorDB{facts: make(map[string]entities.Fact)}
}

func (m *relTestVectorDB) EnsureCollection(_ context.Context, _ uint64) error { return nil }
func (m *relTestVectorDB) DeleteCollection(_ context.Context) error           { return nil }

func (m *relTestVectorDB) Save(_ context.Context, fact *entities.Fact) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.facts[fact.ID] = *fact
	return nil
}

func (m *relTestVectorDB) SaveBatch(_ context.Context, facts []entities.Fact) error {
	for i := range facts {
		m.facts[facts[i].ID] = facts[i]
	}
	return nil
}

func (m *relTestVectorDB) FindByID(_ context.Context, id string) (entities.Fact, error) {
	if f, ok := m.facts[id]; ok {
		return f, nil
	}
	return entities.Fact{}, errors.New("not found")
}

func (m *relTestVectorDB) ExistsByIDs(_ context.Context, ids []string) (map[string]bool, error) {
	result := make(map[string]bool, len(ids))
	for _, id := range ids {
		_, exists := m.facts[id]
		result[id] = exists
	}
	return result, nil
}

func (m *relTestVectorDB) FindByIDs(_ context.Context, ids []string) ([]entities.Fact, error) {
	var result []entities.Fact
	for _, id := range ids {
		if f, ok := m.facts[id]; ok {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *relTestVectorDB) Search(_ context.Context, _ []float32, _ int) ([]entities.Fact, error) {
	return nil, nil
}

func (m *relTestVectorDB) SearchByType(_ context.Context, _ []float32, _ entities.FactType, _ int) ([]entities.Fact, error) {
	return nil, nil
}

func (m *relTestVectorDB) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.facts, id)
	return nil
}

func (m *relTestVectorDB) List(_ context.Context, _ int, _ uint64) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relTestVectorDB) ListByType(_ context.Context, _ entities.FactType, _ int) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relTestVectorDB) ListBySource(_ context.Context, _ string, _ int) ([]entities.Fact, error) {
	return nil, nil
}
func (m *relTestVectorDB) DeleteBySource(_ context.Context, _ string) error { return nil }
func (m *relTestVectorDB) DeleteAll(_ context.Context) error                { return nil }
func (m *relTestVectorDB) Count(_ context.Context) (uint64, error)          { return 0, nil }

// relTestRelationalDB is a test mock for RelationalDB with relationship support.
type relTestRelationalDB struct {
	entities      map[string]*entities.Entity
	relationships map[string]*entities.Relationship
	saveErr       error
	deleteErr     error
	findErr       error
}

func newRelTestRelationalDB() *relTestRelationalDB {
	return &relTestRelationalDB{
		entities:      make(map[string]*entities.Entity),
		relationships: make(map[string]*entities.Relationship),
	}
}

func (m *relTestRelationalDB) EnsureSchema(_ context.Context) error { return nil }
func (m *relTestRelationalDB) Close() error                         { return nil }

// Entity methods.

func (m *relTestRelationalDB) SaveEntity(_ context.Context, entity *entities.Entity) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.entities[entity.ID] = entity
	return nil
}

func (m *relTestRelationalDB) FindEntityByName(_ context.Context, worldID, name string) (*entities.Entity, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	return nil, nil
}

func (m *relTestRelationalDB) FindOrCreateEntity(_ context.Context, worldID, name string) (*entities.Entity, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	normalizedName := entities.NormalizeName(name)
	for _, e := range m.entities {
		if e.WorldID == worldID && e.NormalizedName == normalizedName {
			return e, nil
		}
	}
	// Create new entity
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

func (m *relTestRelationalDB) FindEntityByID(_ context.Context, entityID string) (*entities.Entity, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.entities[entityID], nil
}

func (m *relTestRelationalDB) ListEntities(_ context.Context, _ string, _, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *relTestRelationalDB) SearchEntities(_ context.Context, _, _ string, _ int) ([]*entities.Entity, error) {
	return nil, nil
}

func (m *relTestRelationalDB) DeleteEntity(_ context.Context, entityID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.entities, entityID)
	return nil
}

func (m *relTestRelationalDB) CountEntities(_ context.Context, _ string) (int, error) {
	return len(m.entities), nil
}

// Relationship methods.

func (m *relTestRelationalDB) SaveRelationship(_ context.Context, rel *entities.Relationship) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.relationships[rel.ID] = rel
	return nil
}

func (m *relTestRelationalDB) FindRelationshipsByEntity(_ context.Context, entityID string) ([]entities.Relationship, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
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

func (m *relTestRelationalDB) FindRelationshipsByType(_ context.Context, _ string) ([]entities.Relationship, error) {
	return nil, nil
}

func (m *relTestRelationalDB) DeleteRelationship(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.relationships, id)
	return nil
}

func (m *relTestRelationalDB) DeleteRelationshipsByEntity(_ context.Context, _ string) error {
	return nil
}

func (m *relTestRelationalDB) FindRelationshipBetween(_ context.Context, sourceID, targetID string) (*entities.Relationship, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
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

func (m *relTestRelationalDB) FindRelatedEntities(_ context.Context, entityID string, depth int) ([]string, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if depth < 1 {
		return []string{}, nil
	}
	// Simple implementation: return direct connections only
	seen := make(map[string]bool)
	for _, rel := range m.relationships {
		if rel.SourceEntityID == entityID && rel.TargetEntityID != entityID {
			seen[rel.TargetEntityID] = true
		}
		if rel.Bidirectional && rel.TargetEntityID == entityID && rel.SourceEntityID != entityID {
			seen[rel.SourceEntityID] = true
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	return result, nil
}

func (m *relTestRelationalDB) CountRelationships(_ context.Context) (int, error) {
	return len(m.relationships), nil
}

// No-op implementations for other interface methods
func (m *relTestRelationalDB) SaveEntityType(_ context.Context, _ *entities.EntityType) error {
	return nil
}
func (m *relTestRelationalDB) FindEntityType(_ context.Context, _ string) (*entities.EntityType, error) {
	return nil, nil
}
func (m *relTestRelationalDB) ListEntityTypes(_ context.Context) ([]entities.EntityType, error) {
	return nil, nil
}
func (m *relTestRelationalDB) DeleteEntityType(_ context.Context, _ string) error { return nil }
func (m *relTestRelationalDB) SaveVersion(_ context.Context, _ *entities.FactVersion) error {
	return nil
}
func (m *relTestRelationalDB) FindVersionsByFact(_ context.Context, _ string) ([]entities.FactVersion, error) {
	return nil, nil
}
func (m *relTestRelationalDB) FindLatestVersion(_ context.Context, _ string) (*entities.FactVersion, error) {
	return nil, nil
}
func (m *relTestRelationalDB) CountVersions(_ context.Context, _ string) (int, error) { return 0, nil }
func (m *relTestRelationalDB) LogAction(_ context.Context, _ string, _ string, _ map[string]any) error {
	return nil
}
func (m *relTestRelationalDB) FindAuditLog(_ context.Context, _ string) ([]entities.AuditEntry, error) {
	return nil, nil
}
func (m *relTestRelationalDB) FindAuditLogByAction(_ context.Context, _ string, _ int) ([]entities.AuditEntry, error) {
	return nil, nil
}

// relTestEmbedder is a test mock for Embedder.
type relTestEmbedder struct {
	embedding []float32
	err       error
}

func (m *relTestEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.embedding, nil
}

func (m *relTestEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = m.embedding
	}
	return result, nil
}

// Test setup helper
func setupRelationshipTest() (*RelationshipService, *relTestVectorDB, *relTestRelationalDB, *relTestEmbedder) {
	vectorDB := newRelTestVectorDB()
	relationalDB := newRelTestRelationalDB()
	embedder := &relTestEmbedder{embedding: []float32{0.1, 0.2, 0.3}}

	svc := NewRelationshipService(vectorDB, relationalDB, embedder)
	return svc, vectorDB, relationalDB, embedder
}

const testWorldID = "test-world"

func TestRelationshipService_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		svc, vectorDB, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		rel, err := svc.Create(ctx, testWorldID, "Alice", entities.RelationAlly, "Bob", true)
		require.NoError(t, err)
		require.NotNil(t, rel)

		assert.NotEmpty(t, rel.ID)
		assert.NotEmpty(t, rel.SourceEntityID)
		assert.NotEmpty(t, rel.TargetEntityID)
		assert.Equal(t, entities.RelationAlly, rel.Type)
		assert.True(t, rel.Bidirectional)

		// Verify entities were created
		assert.Len(t, relationalDB.entities, 2)

		// Verify saved to relationalDB
		assert.Len(t, relationalDB.relationships, 1)

		// Verify relationship fact saved to vectorDB
		_, exists := vectorDB.facts[rel.ID]
		assert.True(t, exists)
	})

	t.Run("duplicate relationship", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		// Pre-add entities and relationship
		relationalDB.entities["entity-alice"] = &entities.Entity{
			ID:             "entity-alice",
			WorldID:        testWorldID,
			Name:           "Alice",
			NormalizedName: "alice",
		}
		relationalDB.entities["entity-bob"] = &entities.Entity{
			ID:             "entity-bob",
			WorldID:        testWorldID,
			Name:           "Bob",
			NormalizedName: "bob",
		}
		relationalDB.relationships["existing"] = &entities.Relationship{
			ID:             "existing",
			SourceEntityID: "entity-alice",
			TargetEntityID: "entity-bob",
			Type:           entities.RelationAlly,
		}

		_, err := svc.Create(ctx, testWorldID, "Alice", entities.RelationAlly, "Bob", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "relationship already exists")
	})

	t.Run("embedding error rolls back", func(t *testing.T) {
		svc, _, relationalDB, embedder := setupRelationshipTest()
		ctx := context.Background()

		embedder.err = errors.New("embedding failed")

		_, err := svc.Create(ctx, testWorldID, "Alice", entities.RelationAlly, "Bob", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding")

		// Relationship should be rolled back
		assert.Len(t, relationalDB.relationships, 0)
	})
}

func TestRelationshipService_Create_EntityReuse(t *testing.T) {
	svc, _, relationalDB, _ := setupRelationshipTest()
	ctx := context.Background()

	// Pre-add Alice entity
	relationalDB.entities["entity-alice"] = &entities.Entity{
		ID:             "entity-alice",
		WorldID:        testWorldID,
		Name:           "Alice",
		NormalizedName: "alice",
	}

	rel, err := svc.Create(ctx, testWorldID, "Alice", entities.RelationAlly, "Bob", true)
	require.NoError(t, err)

	// Should reuse Alice entity
	assert.Equal(t, "entity-alice", rel.SourceEntityID)

	// Bob should be created
	assert.Len(t, relationalDB.entities, 2)
}

func TestRelationshipService_Delete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		svc, vectorDB, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		// Pre-add relationship
		relationalDB.relationships["rel-1"] = &entities.Relationship{ID: "rel-1"}
		vectorDB.facts["rel-1"] = entities.Fact{ID: "rel-1"}

		err := svc.Delete(ctx, "rel-1")
		require.NoError(t, err)

		assert.Len(t, relationalDB.relationships, 0)
		_, exists := vectorDB.facts["rel-1"]
		assert.False(t, exists)
	})

	t.Run("vectorDB delete error", func(t *testing.T) {
		svc, vectorDB, _, _ := setupRelationshipTest()
		ctx := context.Background()

		vectorDB.deleteErr = errors.New("delete failed")

		err := svc.Delete(ctx, "rel-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleting relationship fact")
	})
}

func TestRelationshipService_List(t *testing.T) {
	t.Run("returns relationships for entity", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-1",
			TargetEntityID: "entity-2",
			Type:           entities.RelationAlly,
		}
		relationalDB.relationships["rel-2"] = &entities.Relationship{
			ID:             "rel-2",
			SourceEntityID: "entity-3",
			TargetEntityID: "entity-1",
			Type:           entities.RelationSibling,
			Bidirectional:  true,
		}

		rels, err := svc.List(ctx, "entity-1")
		require.NoError(t, err)
		assert.Len(t, rels, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		rels, err := svc.List(ctx, "entity-1")
		require.NoError(t, err)
		assert.Empty(t, rels)
	})
}

func TestRelationshipService_ListByName(t *testing.T) {
	t.Run("returns relationships for entity by name", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		// Add entity
		relationalDB.entities["entity-alice"] = &entities.Entity{
			ID:             "entity-alice",
			WorldID:        testWorldID,
			Name:           "Alice",
			NormalizedName: "alice",
		}

		// Add relationship
		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-alice",
			TargetEntityID: "entity-bob",
			Type:           entities.RelationAlly,
		}

		rels, err := svc.ListByName(ctx, testWorldID, "Alice")
		require.NoError(t, err)
		assert.Len(t, rels, 1)
	})

	t.Run("returns empty for nonexistent entity", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		rels, err := svc.ListByName(ctx, testWorldID, "Nonexistent")
		require.NoError(t, err)
		assert.Empty(t, rels)
	})
}

func TestRelationshipService_ListWithDepth(t *testing.T) {
	t.Run("depth 0 returns empty", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		entities, err := svc.ListWithDepth(ctx, "entity-1", 0)
		require.NoError(t, err)
		assert.Empty(t, entities)
	})

	t.Run("depth 1 returns direct connections", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-1",
			TargetEntityID: "entity-2",
		}

		related, err := svc.ListWithDepth(ctx, "entity-1", 1)
		require.NoError(t, err)
		assert.Len(t, related, 1)
		assert.Equal(t, "entity-2", related[0].EntityID)
	})
}

func TestRelationshipService_FindBetween(t *testing.T) {
	t.Run("finds direct relationship", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-1",
			TargetEntityID: "entity-2",
			Type:           entities.RelationAlly,
		}

		rel, err := svc.FindBetween(ctx, "entity-1", "entity-2")
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "rel-1", rel.ID)
	})

	t.Run("finds bidirectional relationship", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:             "rel-1",
			SourceEntityID: "entity-1",
			TargetEntityID: "entity-2",
			Type:           entities.RelationSibling,
			Bidirectional:  true,
		}

		// Find from reverse direction
		rel, err := svc.FindBetween(ctx, "entity-2", "entity-1")
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "rel-1", rel.ID)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		rel, err := svc.FindBetween(ctx, "entity-1", "entity-2")
		require.NoError(t, err)
		assert.Nil(t, rel)
	})
}

func TestRelationshipService_Count(t *testing.T) {
	t.Run("returns count", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{ID: "rel-1"}
		relationalDB.relationships["rel-2"] = &entities.Relationship{ID: "rel-2"}

		count, err := svc.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
