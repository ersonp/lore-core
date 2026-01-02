package services

import (
	"context"
	"errors"
	"testing"

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
	relationships map[string]*entities.Relationship
	saveErr       error
	deleteErr     error
	findErr       error
}

func newRelTestRelationalDB() *relTestRelationalDB {
	return &relTestRelationalDB{relationships: make(map[string]*entities.Relationship)}
}

func (m *relTestRelationalDB) EnsureSchema(_ context.Context) error { return nil }
func (m *relTestRelationalDB) Close() error                         { return nil }

func (m *relTestRelationalDB) SaveRelationship(_ context.Context, rel *entities.Relationship) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.relationships[rel.ID] = rel
	return nil
}

func (m *relTestRelationalDB) FindRelationshipsByFact(_ context.Context, factID string) ([]entities.Relationship, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []entities.Relationship
	for _, rel := range m.relationships {
		if rel.SourceFactID == factID || (rel.TargetFactID == factID && rel.Bidirectional) {
			result = append(result, *rel)
		}
	}
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

func (m *relTestRelationalDB) DeleteRelationshipsByFact(_ context.Context, _ string) error {
	return nil
}

func (m *relTestRelationalDB) FindRelationshipBetween(_ context.Context, sourceID, targetID string) (*entities.Relationship, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, rel := range m.relationships {
		if rel.SourceFactID == sourceID && rel.TargetFactID == targetID {
			return rel, nil
		}
		if rel.Bidirectional && rel.SourceFactID == targetID && rel.TargetFactID == sourceID {
			return rel, nil
		}
	}
	return nil, nil
}

func (m *relTestRelationalDB) FindRelatedFacts(_ context.Context, factID string, depth int) ([]string, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if depth < 1 {
		return []string{}, nil
	}
	// Simple implementation: return direct connections only
	seen := make(map[string]bool)
	for _, rel := range m.relationships {
		if rel.SourceFactID == factID && rel.TargetFactID != factID {
			seen[rel.TargetFactID] = true
		}
		if rel.Bidirectional && rel.TargetFactID == factID && rel.SourceFactID != factID {
			seen[rel.SourceFactID] = true
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

func TestRelationshipService_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		svc, vectorDB, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		// Add facts to vectorDB
		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		rel, err := svc.Create(ctx, "fact-1", entities.RelationAlly, "fact-2", true)
		require.NoError(t, err)
		require.NotNil(t, rel)

		assert.NotEmpty(t, rel.ID)
		assert.Equal(t, "fact-1", rel.SourceFactID)
		assert.Equal(t, "fact-2", rel.TargetFactID)
		assert.Equal(t, entities.RelationAlly, rel.Type)
		assert.True(t, rel.Bidirectional)

		// Verify saved to relationalDB
		assert.Len(t, relationalDB.relationships, 1)

		// Verify relationship fact saved to vectorDB
		_, exists := vectorDB.facts[rel.ID]
		assert.True(t, exists)
	})

	t.Run("source fact not found", func(t *testing.T) {
		svc, vectorDB, _, _ := setupRelationshipTest()
		ctx := context.Background()

		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		_, err := svc.Create(ctx, "fact-1", entities.RelationAlly, "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source fact not found")
	})

	t.Run("target fact not found", func(t *testing.T) {
		svc, vectorDB, _, _ := setupRelationshipTest()
		ctx := context.Background()

		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}

		_, err := svc.Create(ctx, "fact-1", entities.RelationAlly, "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target fact not found")
	})

	t.Run("duplicate relationship", func(t *testing.T) {
		svc, vectorDB, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		// Pre-add relationship
		relationalDB.relationships["existing"] = &entities.Relationship{
			ID:           "existing",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
		}

		_, err := svc.Create(ctx, "fact-1", entities.RelationAlly, "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "relationship already exists")
	})

	t.Run("embedding error rolls back", func(t *testing.T) {
		svc, vectorDB, relationalDB, embedder := setupRelationshipTest()
		ctx := context.Background()

		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}
		embedder.err = errors.New("embedding failed")

		_, err := svc.Create(ctx, "fact-1", entities.RelationAlly, "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding")

		// Relationship should be rolled back
		assert.Len(t, relationalDB.relationships, 0)
	})
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
	t.Run("returns relationships for fact", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
		}
		relationalDB.relationships["rel-2"] = &entities.Relationship{
			ID:            "rel-2",
			SourceFactID:  "fact-3",
			TargetFactID:  "fact-1",
			Type:          entities.RelationSibling,
			Bidirectional: true,
		}

		rels, err := svc.List(ctx, "fact-1")
		require.NoError(t, err)
		assert.Len(t, rels, 2)
	})

	t.Run("empty list", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		rels, err := svc.List(ctx, "fact-1")
		require.NoError(t, err)
		assert.Empty(t, rels)
	})
}

func TestRelationshipService_ListWithDepth(t *testing.T) {
	t.Run("depth 0 returns empty", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		facts, err := svc.ListWithDepth(ctx, "fact-1", 0)
		require.NoError(t, err)
		assert.Empty(t, facts)
	})

	t.Run("depth 1 returns direct connections", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
		}

		facts, err := svc.ListWithDepth(ctx, "fact-1", 1)
		require.NoError(t, err)
		assert.Len(t, facts, 1)
		assert.Equal(t, "fact-2", facts[0].FactID)
	})
}

func TestRelationshipService_FindBetween(t *testing.T) {
	t.Run("finds direct relationship", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
		}

		rel, err := svc.FindBetween(ctx, "fact-1", "fact-2")
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "rel-1", rel.ID)
	})

	t.Run("finds bidirectional relationship", func(t *testing.T) {
		svc, _, relationalDB, _ := setupRelationshipTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:            "rel-1",
			SourceFactID:  "fact-1",
			TargetFactID:  "fact-2",
			Type:          entities.RelationSibling,
			Bidirectional: true,
		}

		// Find from reverse direction
		rel, err := svc.FindBetween(ctx, "fact-2", "fact-1")
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "rel-1", rel.ID)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		svc, _, _, _ := setupRelationshipTest()
		ctx := context.Background()

		rel, err := svc.FindBetween(ctx, "fact-1", "fact-2")
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
