package integration

import (
	"context"
	"path/filepath"
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
	handler := handlers.NewRelationshipHandler(svc, vectorDB)

	return handler, vectorDB, repo
}

func TestRelationship_Integration_CreateAndQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create two facts
	vectorDB.facts["fact-alice"] = entities.Fact{
		ID:        "fact-alice",
		Type:      entities.FactTypeCharacter,
		Subject:   "Alice",
		Predicate: "is",
		Object:    "a warrior",
		CreatedAt: time.Now(),
	}
	vectorDB.facts["fact-bob"] = entities.Fact{
		ID:        "fact-bob",
		Type:      entities.FactTypeCharacter,
		Subject:   "Bob",
		Predicate: "is",
		Object:    "a mage",
		CreatedAt: time.Now(),
	}

	// Create relationship
	rel, err := handler.HandleCreate(ctx, "fact-alice", "ally", "fact-bob", true)
	require.NoError(t, err)
	require.NotNil(t, rel)

	assert.Equal(t, "fact-alice", rel.SourceFactID)
	assert.Equal(t, "fact-bob", rel.TargetFactID)
	assert.Equal(t, entities.RelationAlly, rel.Type)
	assert.True(t, rel.Bidirectional)

	// Query relationships from source
	result, err := handler.HandleList(ctx, "fact-alice", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, result.Relationships, 1)

	info := result.Relationships[0]
	assert.Equal(t, entities.RelationAlly, info.Relationship.Type)
	require.NotNil(t, info.SourceFact)
	require.NotNil(t, info.TargetFact)
	assert.Equal(t, "Alice", info.SourceFact.Subject)
	assert.Equal(t, "Bob", info.TargetFact.Subject)
}

func TestRelationship_Integration_Bidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	// Create bidirectional relationship
	_, err := handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	// Query from source - should find target
	resultA, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultA.Relationships, 1)

	// Query from target - should find source (bidirectional)
	resultB, err := handler.HandleList(ctx, "fact-b", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultB.Relationships, 1)

	// Both should reference the same relationship
	assert.Equal(t, resultA.Relationships[0].Relationship.ID, resultB.Relationships[0].Relationship.ID)
}

func TestRelationship_Integration_Unidirectional(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	// Create unidirectional relationship (A -> B only)
	_, err := handler.HandleCreate(ctx, "fact-a", "located_in", "fact-b", false)
	require.NoError(t, err)

	// Query from source - should find target
	resultA, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, resultA.Relationships, 1)

	// Query from target - should NOT find source (unidirectional)
	resultB, err := handler.HandleList(ctx, "fact-b", handlers.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, resultB.Relationships, 0)
}

func TestRelationship_Integration_TypeFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}
	vectorDB.facts["fact-c"] = entities.Fact{ID: "fact-c", Subject: "Entity C"}

	// Create relationships of different types
	_, err := handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	_, err = handler.HandleCreate(ctx, "fact-a", "enemy", "fact-c", true)
	require.NoError(t, err)

	// Query all relationships
	all, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, all.Relationships, 2)

	// Query only ally relationships
	allies, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{Type: "ally"})
	require.NoError(t, err)
	require.Len(t, allies.Relationships, 1)
	assert.Equal(t, entities.RelationAlly, allies.Relationships[0].Relationship.Type)

	// Query only enemy relationships
	enemies, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{Type: "enemy"})
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

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	// Create relationship
	rel, err := handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	// Verify it exists
	result, err := handler.HandleList(ctx, "fact-a", handlers.ListOptions{})
	require.NoError(t, err)
	require.Len(t, result.Relationships, 1)

	// Delete relationship
	err = handler.HandleDelete(ctx, rel.ID)
	require.NoError(t, err)

	// Verify it's gone
	result, err = handler.HandleList(ctx, "fact-a", handlers.ListOptions{})
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

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}
	vectorDB.facts["fact-c"] = entities.Fact{ID: "fact-c", Subject: "Entity C"}

	// Create relationship between A and B
	_, err := handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	// Find relationship between A and B - should exist
	rel, err := handler.HandleFindBetween(ctx, "fact-a", "fact-b")
	require.NoError(t, err)
	require.NotNil(t, rel)
	assert.Equal(t, entities.RelationAlly, rel.Type)

	// Find relationship between A and C - should not exist
	rel, err = handler.HandleFindBetween(ctx, "fact-a", "fact-c")
	require.NoError(t, err)
	assert.Nil(t, rel)
}

func TestRelationship_Integration_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Initially no relationships
	count, err := handler.HandleCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}
	vectorDB.facts["fact-c"] = entities.Fact{ID: "fact-c", Subject: "Entity C"}

	// Create relationships
	_, err = handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	_, err = handler.HandleCreate(ctx, "fact-a", "enemy", "fact-c", true)
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

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-source"] = entities.Fact{ID: "fact-source", Subject: "Source"}

	validTypes := []string{
		"parent", "child", "sibling", "spouse",
		"ally", "enemy", "located_in", "owns", "member_of", "created",
	}

	for i, relType := range validTypes {
		targetID := "fact-target-" + relType
		vectorDB.facts[targetID] = entities.Fact{ID: targetID, Subject: "Target " + relType}

		rel, err := handler.HandleCreate(ctx, "fact-source", relType, targetID, false)
		require.NoError(t, err, "type %s should be valid", relType)
		assert.Equal(t, entities.RelationType(relType), rel.Type)

		// Clean up for next iteration
		_ = handler.HandleDelete(ctx, rel.ID)
		_ = i // silence unused warning
	}
}

func TestRelationship_Integration_InvalidType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Create facts
	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	// Try invalid relationship type
	_, err := handler.HandleCreate(ctx, "fact-a", "invalid_type", "fact-b", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid relationship type")
}

func TestRelationship_Integration_SourceNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler, vectorDB, _ := setupRelationshipTest(t)
	ctx := context.Background()

	// Only create target fact
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	// Try to create relationship with non-existent source
	_, err := handler.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source fact not found")
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

	vectorDB.facts["fact-a"] = entities.Fact{ID: "fact-a", Subject: "Entity A"}
	vectorDB.facts["fact-b"] = entities.Fact{ID: "fact-b", Subject: "Entity B"}

	svc1 := services.NewRelationshipService(vectorDB, repo1, embedder)
	handler1 := handlers.NewRelationshipHandler(svc1, vectorDB)

	ctx := context.Background()
	rel, err := handler1.HandleCreate(ctx, "fact-a", "ally", "fact-b", true)
	require.NoError(t, err)

	repo1.Close()

	// Second session: verify persistence
	repo2, err := sqlite.NewRepository(config.SQLiteConfig{Path: dbPath})
	require.NoError(t, err)
	defer repo2.Close()

	// Relationship should still exist in SQLite
	rels, err := repo2.FindRelationshipsByFact(ctx, "fact-a")
	require.NoError(t, err)
	require.Len(t, rels, 1)
	assert.Equal(t, rel.ID, rels[0].ID)
	assert.Equal(t, entities.RelationAlly, rels[0].Type)
}
