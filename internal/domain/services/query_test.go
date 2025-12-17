package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

type mockEmbedder struct {
	embedding []float32
	err       error
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.embedding, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = m.embedding
	}
	return result, nil
}

type mockVectorDB struct {
	facts []entities.Fact
	err   error
}

func (m *mockVectorDB) Save(ctx context.Context, fact entities.Fact) error {
	return m.err
}

func (m *mockVectorDB) SaveBatch(ctx context.Context, facts []entities.Fact) error {
	return m.err
}

func (m *mockVectorDB) FindByID(ctx context.Context, id string) (entities.Fact, error) {
	if m.err != nil {
		return entities.Fact{}, m.err
	}
	for _, f := range m.facts {
		if f.ID == id {
			return f, nil
		}
	}
	return entities.Fact{}, nil
}

func (m *mockVectorDB) Search(ctx context.Context, embedding []float32, limit int) ([]entities.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	if limit > len(m.facts) {
		return m.facts, nil
	}
	return m.facts[:limit], nil
}

func (m *mockVectorDB) SearchByType(ctx context.Context, embedding []float32, factType entities.FactType, limit int) ([]entities.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []entities.Fact
	for _, f := range m.facts {
		if f.Type == factType {
			filtered = append(filtered, f)
		}
	}
	if limit > len(filtered) {
		return filtered, nil
	}
	return filtered[:limit], nil
}

func (m *mockVectorDB) Delete(ctx context.Context, id string) error {
	return m.err
}

func (m *mockVectorDB) List(ctx context.Context, limit int, offset uint64) ([]entities.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.facts, nil
}

func (m *mockVectorDB) ListByType(ctx context.Context, factType entities.FactType, limit int) ([]entities.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []entities.Fact
	for _, f := range m.facts {
		if f.Type == factType {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

func (m *mockVectorDB) ListBySource(ctx context.Context, sourceFile string, limit int) ([]entities.Fact, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []entities.Fact
	for _, f := range m.facts {
		if f.SourceFile == sourceFile {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

func (m *mockVectorDB) DeleteBySource(ctx context.Context, sourceFile string) error {
	return m.err
}

func (m *mockVectorDB) DeleteAll(ctx context.Context) error {
	return m.err
}

func (m *mockVectorDB) Count(ctx context.Context) (uint64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return uint64(len(m.facts)), nil
}

func TestQueryService_Search(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:        "1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Frodo",
			Predicate: "eye_color",
			Object:    "blue",
		},
		{
			ID:        "2",
			Type:      entities.FactTypeLocation,
			Subject:   "Shire",
			Predicate: "type",
			Object:    "region",
		},
	}

	emb := &mockEmbedder{embedding: []float32{0.1, 0.2, 0.3}}
	db := &mockVectorDB{facts: facts}

	svc := NewQueryService(emb, db)

	result, err := svc.Search(context.Background(), "What color are Frodo's eyes?", 10)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestQueryService_SearchByType(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:        "1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Frodo",
			Predicate: "eye_color",
			Object:    "blue",
		},
		{
			ID:        "2",
			Type:      entities.FactTypeLocation,
			Subject:   "Shire",
			Predicate: "type",
			Object:    "region",
		},
	}

	emb := &mockEmbedder{embedding: []float32{0.1, 0.2, 0.3}}
	db := &mockVectorDB{facts: facts}

	svc := NewQueryService(emb, db)

	result, err := svc.SearchByType(context.Background(), "characters", entities.FactTypeCharacter, 10)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Frodo", result[0].Subject)
}

func TestQueryService_DefaultLimit(t *testing.T) {
	emb := &mockEmbedder{embedding: []float32{0.1, 0.2, 0.3}}
	db := &mockVectorDB{facts: []entities.Fact{}}

	svc := NewQueryService(emb, db)

	_, err := svc.Search(context.Background(), "test", 0)
	require.NoError(t, err)
}
