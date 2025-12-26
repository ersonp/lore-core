package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/domain/services"
)

func TestQueryHandler_Handle(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:        "1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Frodo",
			Predicate: "has_trait",
			Object:    "brave",
		},
	}

	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: facts}
	queryService := services.NewQueryService(emb, db)
	handler := NewQueryHandler(queryService)

	result, err := handler.Handle(t.Context(), "Who is brave?", 10)
	require.NoError(t, err)
	assert.Equal(t, "Who is brave?", result.Query)
	assert.Len(t, result.Facts, 1)
	assert.Equal(t, "Frodo", result.Facts[0].Subject)
}

func TestQueryHandler_Handle_NoResults(t *testing.T) {
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: []entities.Fact{}}
	queryService := services.NewQueryService(emb, db)
	handler := NewQueryHandler(queryService)

	result, err := handler.Handle(t.Context(), "Unknown query", 10)
	require.NoError(t, err)
	assert.Empty(t, result.Facts)
}

func TestQueryHandler_HandleByType(t *testing.T) {
	facts := []entities.Fact{
		{
			ID:        "1",
			Type:      entities.FactTypeCharacter,
			Subject:   "Frodo",
			Predicate: "has_trait",
			Object:    "brave",
		},
		{
			ID:        "2",
			Type:      entities.FactTypeLocation,
			Subject:   "Shire",
			Predicate: "is_a",
			Object:    "region",
		},
	}

	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{Facts: facts}
	queryService := services.NewQueryService(emb, db)
	handler := NewQueryHandler(queryService)

	result, err := handler.HandleByType(t.Context(), "characters", entities.FactTypeCharacter, 10)
	require.NoError(t, err)
	assert.Len(t, result.Facts, 1)
	assert.Equal(t, entities.FactTypeCharacter, result.Facts[0].Type)
}

func TestNewQueryHandler(t *testing.T) {
	emb := &mocks.Embedder{}
	db := &mocks.VectorDB{}
	queryService := services.NewQueryService(emb, db)

	handler := NewQueryHandler(queryService)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.queryService)
}
