package integration

import (
	"context"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	"github.com/ersonp/lore-core/internal/infrastructure/relationaldb/sqlite"
)

// relTestVectorDB is a mock VectorDB for integration tests.
type relTestVectorDB struct {
	facts map[string]entities.Fact
}

func newRelTestVectorDB() *relTestVectorDB {
	return &relTestVectorDB{facts: make(map[string]entities.Fact)}
}

func (m *relTestVectorDB) EnsureCollection(_ context.Context, _ uint64) error { return nil }
func (m *relTestVectorDB) DeleteCollection(_ context.Context) error           { return nil }

func (m *relTestVectorDB) Save(_ context.Context, fact *entities.Fact) error {
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
	return entities.Fact{}, nil
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

// relTestEmbedder is a mock Embedder for integration tests.
type relTestEmbedder struct{}

func (m *relTestEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return make([]float32, 1536), nil
}

func (m *relTestEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 1536)
	}
	return result, nil
}

// setupRelationshipTest creates test dependencies.
func setupRelationshipTest(t *testing.T) (*handlers.RelationshipHandler, *relTestVectorDB, ports.RelationalDB) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	repo, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)

	err = repo.EnsureSchema(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() {
		repo.Close()
	})

	vectorDB := newRelTestVectorDB()
	embedder := &relTestEmbedder{}

	svc := services.NewRelationshipService(vectorDB, repo, embedder)
	handler := handlers.NewRelationshipHandler(svc, repo)

	return handler, vectorDB, repo
}

const testWorldID = "test-world"

func TestRelationship_Integration_CreateAndQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create relationship (entities auto-created)
	rel, err := handler.HandleCreate(ctx, testWorldID, "Alice", "ally", "Bob", true)
	require.NoError(t, err)
	require.NotNil(t, rel)

	assert.NotEmpty(t, rel.SourceEntityID)
	assert.NotEmpty(t, rel.TargetEntityID)
	assert.Equal(t, entities.RelationAlly, rel.Type)
	assert.True(t, rel.Bidirectional)

	// Query relationships from source
	result, err := handler.HandleList(ctx, testWorldID, "Alice", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, result.Relationships, 1)

	info := result.Relationships[0]
	assert.Equal(t, entities.RelationAlly, info.Relationship.Type)
	require.NotNil(t, info.SourceEntity)
	require.NotNil(t, info.TargetEntity)
	assert.Equal(t, "Alice", info.SourceEntity.Name)
	assert.Equal(t, "Bob", info.TargetEntity.Name)
}

func TestRelationship_Integration_Bidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create bidirectional relationship
	_, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "ally", "EntityB", true)
	require.NoError(t, err)

	// Query from source - should find target
	resultA, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultA.Relationships, 1)

	// Query from target - should find source (bidirectional)
	resultB, err := handler.HandleList(ctx, testWorldID, "EntityB", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultB.Relationships, 1)

	// Both should reference the same relationship
	assert.Equal(t, resultA.Relationships[0].Relationship.ID, resultB.Relationships[0].Relationship.ID)
}

func TestRelationship_Integration_Unidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create unidirectional relationship (A -> B only)
	_, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "located_in", "EntityB", false)
	require.NoError(t, err)

	// Query from source - should find target
	resultA, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultA.Relationships, 1)

	// Query from target - should NOT find source (unidirectional)
	resultB, err := handler.HandleList(ctx, testWorldID, "EntityB", handlers.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, resultB.Relationships, 0)
}

func TestRelationship_Integration_TypeFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create relationships of different types
	_, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "ally", "EntityB", true)
	require.NoError(t, err)

	_, err = handler.HandleCreate(ctx, testWorldID, "EntityA", "enemy", "EntityC", true)
	require.NoError(t, err)

	// Query all relationships
	all, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, all.Relationships, 2)

	// Query only ally relationships
	allies, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{Type: "ally"})
	require.NoError(t, err)
	require.Len(t, allies.Relationships, 1)
	assert.Equal(t, entities.RelationAlly, allies.Relationships[0].Relationship.Type)

	// Query only enemy relationships
	enemies, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{Type: "enemy"})
	require.NoError(t, err)
	require.Len(t, enemies.Relationships, 1)
	assert.Equal(t, entities.RelationEnemy, enemies.Relationships[0].Relationship.Type)
}

func TestRelationship_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create relationship
	rel, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "ally", "EntityB", true)
	require.NoError(t, err)

	// Verify it exists
	result, err := handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, result.Relationships, 1)

	// Delete relationship
	err = handler.HandleDelete(ctx, rel.ID)
	require.NoError(t, err)

	// Verify it's gone
	result, err = handler.HandleList(ctx, testWorldID, "EntityA", handlers.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, result.Relationships, 0)

	// Verify relationship fact is gone from VectorDB
	_, exists := vectorDB.facts[rel.ID]
	assert.False(t, exists, "relationship fact should be deleted from VectorDB")
}

func TestRelationship_Integration_FindBetween(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create relationship between A and B
	rel, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "ally", "EntityB", true)
	require.NoError(t, err)

	// Find relationship between A and B - should exist
	found, err := handler.HandleFindBetween(ctx, rel.SourceEntityID, rel.TargetEntityID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, entities.RelationAlly, found.Type)

	// Create another entity - no relationship with A
	_, err = handler.HandleCreate(ctx, testWorldID, "EntityA", "enemy", "EntityC", true)
	require.NoError(t, err)

	// Find relationship between B and C - should not exist
	found, err = handler.HandleFindBetween(ctx, rel.TargetEntityID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRelationship_Integration_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Initially no relationships
	count, err := handler.HandleCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create relationships
	_, err = handler.HandleCreate(ctx, testWorldID, "EntityA", "ally", "EntityB", true)
	require.NoError(t, err)

	_, err = handler.HandleCreate(ctx, testWorldID, "EntityA", "enemy", "EntityC", true)
	require.NoError(t, err)

	// Should have 2 relationships
	count, err = handler.HandleCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestRelationship_Integration_AllRelationTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	validTypes := []string{
		"parent", "child", "sibling", "spouse",
		"ally", "enemy", "located_in", "owns", "member_of", "created",
	}

	for _, relType := range validTypes {
		targetName := "Target" + relType

		rel, err := handler.HandleCreate(ctx, testWorldID, "Source", relType, targetName, false)
		require.NoError(t, err, "type %s should be valid", relType)
		assert.Equal(t, entities.RelationType(relType), rel.Type)

		// Clean up for next iteration
		_ = handler.HandleDelete(ctx, rel.ID)
	}
}

func TestRelationship_Integration_InvalidType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Try invalid relationship type
	_, err := handler.HandleCreate(ctx, testWorldID, "EntityA", "invalid_type", "EntityB", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid relationship type")
}

func TestRelationship_Integration_Persistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First session: create relationship
	repo1, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)

	err = repo1.EnsureSchema(context.Background())
	require.NoError(t, err)

	vectorDB := newRelTestVectorDB()
	embedder := &relTestEmbedder{}

	svc1 := services.NewRelationshipService(vectorDB, repo1, embedder)
	handler1 := handlers.NewRelationshipHandler(svc1, repo1)

	ctx := context.Background()
	rel, err := handler1.HandleCreate(ctx, testWorldID, "Alice", "ally", "Bob", true)
	require.NoError(t, err)

	repo1.Close()

	// Second session: verify persistence
	repo2, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo2.Close()

	// Relationship should still exist in SQLite
	rels, err := repo2.FindRelationshipsByEntity(ctx, rel.SourceEntityID)
	require.NoError(t, err)
	require.Len(t, rels, 1)
	assert.Equal(t, rel.ID, rels[0].ID)
	assert.Equal(t, entities.RelationAlly, rels[0].Type)
}

func TestRelationship_Integration_EntityAutoCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, relationalDB := setupRelationshipTest(t)
	ctx := context.Background()

	// Entities don't exist yet
	entity, err := relationalDB.FindEntityByName(ctx, testWorldID, "NewEntity1")
	require.NoError(t, err)
	assert.Nil(t, entity)

	// Create relationship - should auto-create entities
	_, err = handler.HandleCreate(ctx, testWorldID, "NewEntity1", "ally", "NewEntity2", true)
	require.NoError(t, err)

	// Now entities should exist
	entity1, err := relationalDB.FindEntityByName(ctx, testWorldID, "NewEntity1")
	require.NoError(t, err)
	require.NotNil(t, entity1)
	assert.Equal(t, "NewEntity1", entity1.Name)

	entity2, err := relationalDB.FindEntityByName(ctx, testWorldID, "NewEntity2")
	require.NoError(t, err)
	require.NotNil(t, entity2)
	assert.Equal(t, "NewEntity2", entity2.Name)
}

func TestRelationship_Integration_CaseInsensitiveEntityMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create relationship with "Alice"
	_, err := handler.HandleCreate(ctx, testWorldID, "Alice", "ally", "Bob", true)
	require.NoError(t, err)

	// Query with "alice" (lowercase) - should find same entity
	result, err := handler.HandleList(ctx, testWorldID, "alice", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, result.Relationships, 1)

	// Create another relationship with "ALICE" - should use same entity
	_, err = handler.HandleCreate(ctx, testWorldID, "ALICE", "enemy", "Eve", true)
	require.NoError(t, err)

	// Query should now show 2 relationships
	result, err = handler.HandleList(ctx, testWorldID, "Alice", handlers.ListOptions{})
	require.NoError(t, err)

	// Sort for deterministic test
	sort.Slice(result.Relationships, func(i, j int) bool {
		return result.Relationships[i].Relationship.ID < result.Relationships[j].Relationship.ID
	})
	assert.Len(t, result.Relationships, 2)
}

func TestRelationship_Integration_EntityTimestamps(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, _, relationalDB := setupRelationshipTest(t)
	ctx := context.Background()

	beforeCreate := time.Now()

	// Create relationship - should auto-create entities with timestamps
	_, err := handler.HandleCreate(ctx, testWorldID, "EntityWithTimestamp", "ally", "OtherEntity", true)
	require.NoError(t, err)

	afterCreate := time.Now()

	// Check entity has valid timestamp
	entity, err := relationalDB.FindEntityByName(ctx, testWorldID, "EntityWithTimestamp")
	require.NoError(t, err)
	require.NotNil(t, entity)

	assert.False(t, entity.CreatedAt.IsZero(), "CreatedAt should not be zero")
	assert.True(t, entity.CreatedAt.After(beforeCreate) || entity.CreatedAt.Equal(beforeCreate))
	assert.True(t, entity.CreatedAt.Before(afterCreate) || entity.CreatedAt.Equal(afterCreate))
}
