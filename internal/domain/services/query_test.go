package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/mocks"
)

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

	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: facts}

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

	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: facts}

	svc := NewQueryService(emb, db)

	result, err := svc.SearchByType(context.Background(), "characters", entities.FactTypeCharacter, 10)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Frodo", result[0].Subject)
}

func TestQueryService_DefaultLimit(t *testing.T) {
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: []entities.Fact{}}

	svc := NewQueryService(emb, db)

	_, err := svc.Search(context.Background(), "test", 0)
	require.NoError(t, err)
}
