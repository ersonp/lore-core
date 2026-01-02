package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// relHandlerVectorDB is a test mock for VectorDB.
type relHandlerVectorDB struct {
	facts     map[string]entities.Fact
	saveErr   error
	deleteErr error
}

func newRelHandlerVectorDB() *relHandlerVectorDB {
	return &relHandlerVectorDB{facts: make(map[string]entities.Fact)}
}

func (m *relHandlerVectorDB) EnsureCollection(_ context.Context, _ uint64) error { return nil }
func (m *relHandlerVectorDB) DeleteCollection(_ context.Context) error           { return nil }

func (m *relHandlerVectorDB) Save(_ context.Context, fact *entities.Fact) error {
	if m.saveErr != nil {
		return m.saveErr
	}
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
	return entities.Fact{}, errors.New("not found")
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
	var result []entities.Fact
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
	if m.deleteErr != nil {
		return m.deleteErr
	}
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
	relationships map[string]*entities.Relationship
}

func newRelHandlerRelationalDB() *relHandlerRelationalDB {
	return &relHandlerRelationalDB{relationships: make(map[string]*entities.Relationship)}
}

func (m *relHandlerRelationalDB) EnsureSchema(_ context.Context) error { return nil }
func (m *relHandlerRelationalDB) Close() error                         { return nil }

func (m *relHandlerRelationalDB) SaveRelationship(_ context.Context, rel *entities.Relationship) error {
	m.relationships[rel.ID] = rel
	return nil
}

func (m *relHandlerRelationalDB) FindRelationshipsByFact(_ context.Context, factID string) ([]entities.Relationship, error) {
	var result []entities.Relationship
	for _, rel := range m.relationships {
		if rel.SourceFactID == factID || (rel.TargetFactID == factID && rel.Bidirectional) {
			result = append(result, *rel)
		}
	}
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

func (m *relHandlerRelationalDB) DeleteRelationshipsByFact(_ context.Context, _ string) error {
	return nil
}

func (m *relHandlerRelationalDB) FindRelationshipBetween(_ context.Context, sourceID, targetID string) (*entities.Relationship, error) {
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

func (m *relHandlerRelationalDB) FindRelatedFacts(_ context.Context, factID string, _ int) ([]string, error) {
	seen := make(map[string]bool)
	for _, rel := range m.relationships {
		if rel.SourceFactID == factID {
			seen[rel.TargetFactID] = true
		}
		if rel.Bidirectional && rel.TargetFactID == factID {
			seen[rel.SourceFactID] = true
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

// No-op implementations
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
	handler := NewRelationshipHandler(svc, vectorDB)

	return handler, vectorDB, relationalDB
}

func TestRelationshipHandler_HandleCreate(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		handler, vectorDB, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Add facts
		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		rel, err := handler.HandleCreate(ctx, "fact-1", "ally", "fact-2", true)
		require.NoError(t, err)
		require.NotNil(t, rel)

		assert.Equal(t, "fact-1", rel.SourceFactID)
		assert.Equal(t, "fact-2", rel.TargetFactID)
		assert.Equal(t, entities.RelationAlly, rel.Type)
		assert.True(t, rel.Bidirectional)
	})

	t.Run("invalid relationship type", func(t *testing.T) {
		handler, vectorDB, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		_, err := handler.HandleCreate(ctx, "fact-1", "invalid_type", "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid relationship type")
	})

	t.Run("source fact not found", func(t *testing.T) {
		handler, vectorDB, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		_, err := handler.HandleCreate(ctx, "fact-1", "ally", "fact-2", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "source fact not found")
	})

	t.Run("all valid relationship types", func(t *testing.T) {
		validTypes := []string{
			"parent", "child", "sibling", "spouse",
			"ally", "enemy", "located_in", "owns", "member_of", "created",
		}

		for _, relType := range validTypes {
			handler, vectorDB, _ := setupRelationshipHandlerTest()
			ctx := context.Background()

			vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
			vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

			rel, err := handler.HandleCreate(ctx, "fact-1", relType, "fact-2", false)
			require.NoError(t, err, "type %s should be valid", relType)
			assert.Equal(t, entities.RelationType(relType), rel.Type)
		}
	})
}

func TestRelationshipHandler_HandleDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		handler, vectorDB, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Pre-add relationship
		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
		}
		vectorDB.facts["rel-1"] = entities.Fact{ID: "rel-1"}

		err := handler.HandleDelete(ctx, "rel-1")
		require.NoError(t, err)

		assert.Len(t, relationalDB.relationships, 0)
	})
}

func TestRelationshipHandler_HandleList(t *testing.T) {
	t.Run("returns relationships with fact details", func(t *testing.T) {
		handler, vectorDB, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		// Add facts
		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}

		// Add relationship
		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
			CreatedAt:    time.Now(),
		}

		result, err := handler.HandleList(ctx, "fact-1", ListOptions{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Relationships, 1)

		info := result.Relationships[0]
		assert.Equal(t, "rel-1", info.Relationship.ID)
		require.NotNil(t, info.SourceFact)
		require.NotNil(t, info.TargetFact)
		assert.Equal(t, "Alice", info.SourceFact.Subject)
		assert.Equal(t, "Bob", info.TargetFact.Subject)
	})

	t.Run("filters by type", func(t *testing.T) {
		handler, vectorDB, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		vectorDB.facts["fact-1"] = entities.Fact{ID: "fact-1", Subject: "Alice"}
		vectorDB.facts["fact-2"] = entities.Fact{ID: "fact-2", Subject: "Bob"}
		vectorDB.facts["fact-3"] = entities.Fact{ID: "fact-3", Subject: "City"}

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
		}
		relationalDB.relationships["rel-2"] = &entities.Relationship{
			ID:           "rel-2",
			SourceFactID: "fact-1",
			TargetFactID: "fact-3",
			Type:         entities.RelationLocatedIn,
		}

		result, err := handler.HandleList(ctx, "fact-1", ListOptions{Type: "ally"})
		require.NoError(t, err)
		assert.Len(t, result.Relationships, 1)
		assert.Equal(t, entities.RelationAlly, result.Relationships[0].Relationship.Type)
	})

	t.Run("empty results", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		result, err := handler.HandleList(ctx, "nonexistent", ListOptions{})
		require.NoError(t, err)
		assert.Empty(t, result.Relationships)
	})
}

func TestRelationshipHandler_HandleFindBetween(t *testing.T) {
	t.Run("finds relationship", func(t *testing.T) {
		handler, _, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{
			ID:           "rel-1",
			SourceFactID: "fact-1",
			TargetFactID: "fact-2",
			Type:         entities.RelationAlly,
		}

		rel, err := handler.HandleFindBetween(ctx, "fact-1", "fact-2")
		require.NoError(t, err)
		require.NotNil(t, rel)
		assert.Equal(t, "rel-1", rel.ID)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		handler, _, _ := setupRelationshipHandlerTest()
		ctx := context.Background()

		rel, err := handler.HandleFindBetween(ctx, "fact-1", "fact-2")
		require.NoError(t, err)
		assert.Nil(t, rel)
	})
}

func TestRelationshipHandler_HandleCount(t *testing.T) {
	t.Run("returns count", func(t *testing.T) {
		handler, _, relationalDB := setupRelationshipHandlerTest()
		ctx := context.Background()

		relationalDB.relationships["rel-1"] = &entities.Relationship{ID: "rel-1"}
		relationalDB.relationships["rel-2"] = &entities.Relationship{ID: "rel-2"}

		count, err := handler.HandleCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
