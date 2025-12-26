package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
)

func TestCollectionLifecycle(t *testing.T) {
	ctx := context.Background()

	// Collection should already exist from TestMain
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)

	// Ensure idempotent - calling EnsureCollection again should not fail
	err = testRepo.EnsureCollection(ctx, uint64(embedder.VectorSize))
	require.NoError(t, err)
}

func TestSaveAndFindByID(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	fact := entities.Fact{
		ID:         uuid.New().String(),
		Type:       entities.FactTypeCharacter,
		Subject:    "Frodo",
		Predicate:  "is a",
		Object:     "hobbit",
		Context:    "The Lord of the Rings",
		SourceFile: "test.txt",
		Confidence: 0.95,
		Embedding:  make([]float32, embedder.VectorSize),
	}

	// Save
	err := testRepo.Save(ctx, fact)
	require.NoError(t, err)

	// Retrieve by ID
	retrieved, err := testRepo.FindByID(ctx, fact.ID)
	require.NoError(t, err)
	assert.Equal(t, fact.ID, retrieved.ID)
	assert.Equal(t, fact.Subject, retrieved.Subject)
	assert.Equal(t, fact.Predicate, retrieved.Predicate)
	assert.Equal(t, fact.Object, retrieved.Object)
	assert.Equal(t, fact.Context, retrieved.Context)
	assert.Equal(t, fact.SourceFile, retrieved.SourceFile)
	assert.Equal(t, fact.Type, retrieved.Type)
	assert.InDelta(t, fact.Confidence, retrieved.Confidence, 0.001)
}

func TestSaveAndCount(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	// Start with empty
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)

	// Save one fact
	fact := entities.Fact{
		ID:        uuid.New().String(),
		Type:      entities.FactTypeLocation,
		Subject:   "Mordor",
		Predicate: "is",
		Object:    "a dark land",
		Embedding: make([]float32, embedder.VectorSize),
	}
	err = testRepo.Save(ctx, fact)
	require.NoError(t, err)

	// Count should be 1
	count, err = testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), count)
}

func TestSaveAndDelete(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	fact := entities.Fact{
		ID:        uuid.New().String(),
		Type:      entities.FactTypeCharacter,
		Subject:   "Gandalf",
		Predicate: "is a",
		Object:    "wizard",
		Embedding: make([]float32, embedder.VectorSize),
	}

	// Save
	err := testRepo.Save(ctx, fact)
	require.NoError(t, err)

	// Verify exists
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), count)

	// Delete
	err = testRepo.Delete(ctx, fact.ID)
	require.NoError(t, err)

	// Verify deleted
	count, err = testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestBatchSave(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	facts := []entities.Fact{
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Aragorn",
			Predicate: "is",
			Object:    "king of Gondor",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Legolas",
			Predicate: "is an",
			Object:    "elf",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeLocation,
			Subject:   "Rivendell",
			Predicate: "is",
			Object:    "elven city",
			Embedding: make([]float32, embedder.VectorSize),
		},
	}

	// Batch save
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// Verify count
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), count)
}

func TestDeleteAll(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	// Save multiple facts
	facts := []entities.Fact{
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Sam",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Merry",
			Embedding: make([]float32, embedder.VectorSize),
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// Verify saved
	count, err := testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count)

	// Delete all
	err = testRepo.DeleteAll(ctx)
	require.NoError(t, err)

	// Verify empty
	count, err = testRepo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestList(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { cleanupFacts(t) })

	// Save facts
	facts := []entities.Fact{
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeCharacter,
			Subject:   "Bilbo",
			Predicate: "found",
			Object:    "the ring",
			Embedding: make([]float32, embedder.VectorSize),
		},
		{
			ID:        uuid.New().String(),
			Type:      entities.FactTypeEvent,
			Subject:   "Battle of Five Armies",
			Predicate: "occurred at",
			Object:    "Erebor",
			Embedding: make([]float32, embedder.VectorSize),
		},
	}
	err := testRepo.SaveBatch(ctx, facts)
	require.NoError(t, err)

	// List all
	listed, err := testRepo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, listed, 2)
}
