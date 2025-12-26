package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
)

func TestListByType(t *testing.T) {
	ctx := t.Context()
	t.Cleanup(func() { cleanupFacts(t) })

	// Save facts of different types
	facts := []entities.Fact{
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Sauron",
			Predicate: "is",
			Object:    "the dark lord",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Saruman",
			Predicate: "is",
			Object:    "a corrupted wizard",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeLocation,
			Subject:   "Isengard",
			Predicate: "is",
			Object:    "tower fortress",
			Embedding: make([]float32, embedder.VectorSize),
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// List by character type
	characters, err := testRepo.ListByType(ctx, entities.FactTypeCharacter, 10)
	require.NoError(t, err)
	assert.Len(t, characters, 2)
	for _, c := range characters {
		assert.Equal(t, entities.FactTypeCharacter, c.Type)
	}

	// List by location type
	locations, err := testRepo.ListByType(ctx, entities.FactTypeLocation, 10)
	require.NoError(t, err)
	assert.Len(t, locations, 1)
	assert.Equal(t, "Isengard", locations[0].Subject)
}

func TestListBySource(t *testing.T) {
	ctx := t.Context()
	t.Cleanup(func() { cleanupFacts(t) })

	// Save facts from different sources
	facts := []entities.Fact{
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeCharacter,
			Subject:    "Gollum",
			Predicate:  "was",
			Object:     "a hobbit",
			SourceFile: "chapter1.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeEvent,
			Subject:    "Ring",
			Predicate:  "corrupted",
			Object:     "Gollum",
			SourceFile: "chapter1.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeCharacter,
			Subject:    "Boromir",
			Predicate:  "is",
			Object:     "son of Denethor",
			SourceFile: "chapter2.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// List by source chapter1
	chapter1Facts, err := testRepo.ListBySource(ctx, "chapter1.txt", 10)
	require.NoError(t, err)
	assert.Len(t, chapter1Facts, 2)
	for _, f := range chapter1Facts {
		assert.Equal(t, "chapter1.txt", f.SourceFile)
	}

	// List by source chapter2
	chapter2Facts, err := testRepo.ListBySource(ctx, "chapter2.txt", 10)
	require.NoError(t, err)
	assert.Len(t, chapter2Facts, 1)
	assert.Equal(t, "Boromir", chapter2Facts[0].Subject)
}

func TestDeleteBySource(t *testing.T) {
	ctx := t.Context()
	t.Cleanup(func() { cleanupFacts(t) })

	// Save facts from different sources
	facts := []entities.Fact{
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeCharacter,
			Subject:    "Theoden",
			SourceFile: "rohan.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeCharacter,
			Subject:    "Eomer",
			SourceFile: "rohan.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
		{
			ID:         uuid.New().String(),
			Type:       entities.FactTypeCharacter,
			Subject:    "Faramir",
			SourceFile: "gondor.txt",
			Embedding:  make([]float32, embedder.VectorSize),
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// Delete by source rohan.txt
	err = testRepo.DeleteBySource(ctx, "rohan.txt")
	require.NoError(t, err)

	// Verify only gondor.txt remains
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), count)

	remaining, err := testRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, remaining, 1)
	assert.Equal(t, "Faramir", remaining[0].Subject)
}

func TestSearch(t *testing.T) {
	ctx := t.Context()
	t.Cleanup(func() { cleanupFacts(t) })

	// Create a simple embedding (non-zero for search to work)
	embedding := make([]float32, embedder.VectorSize)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Save a fact with embedding
	fact := entities.Fact{
		ID:        uuid.New().String(),
		Type:      entities.FactTypeCharacter,
		Subject:   "Elrond",
		Predicate: "is",
		Object:    "lord of Rivendell",
		Embedding: embedding,
	}
	err := testRepo.Save(ctx, fact)
	require.NoError(t, err)

	// Search with similar embedding
	results, err := testRepo.Search(ctx, embedding, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Elrond", results[0].Subject)
}

func TestSearchByType(t *testing.T) {
	ctx := t.Context()
	t.Cleanup(func() { cleanupFacts(t) })

	// Create embeddings
	embedding := make([]float32, embedder.VectorSize)
	for i := range embedding {
		embedding[i] = 0.1
	}

	// Save facts of different types
	facts := []entities.Fact{
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Galadriel",
			Embedding: embedding,
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeLocation,
			Subject:   "Lothlorien",
			Embedding: embedding,
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// Search by character type
	results, err := testRepo.SearchByType(ctx, embedding, entities.FactTypeCharacter, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Galadriel", results[0].Subject)

	// Search by location type
	results, err = testRepo.SearchByType(ctx, embedding, entities.FactTypeLocation, 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Lothlorien", results[0].Subject)
}
