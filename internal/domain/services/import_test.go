package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/infrastructure/parsers"
)

func TestImportService_Import_ValidFacts(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is a", Object: "wizard"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
	assert.Equal(t, 1, vectorDB.SaveBatchCallCount)
}

func TestImportService_Import_ValidationErrors(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
		{Type: "character", Subject: "", Predicate: "is", Object: "wizard"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Len(t, result.Errors, 2)
	assert.Equal(t, "type", result.Errors[0].Field)
	assert.Equal(t, "subject", result.Errors[1].Field)
}

func TestImportService_Import_InvalidFactType(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "invalid_type", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "type", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "invalid type")
}

func TestImportService_Import_InvalidConfidence(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	invalidConf := 1.5
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard", Confidence: &invalidConf},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "confidence", result.Errors[0].Field)
}

func TestImportService_Import_ConfidenceZero(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	zeroConf := 0.0
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard", Confidence: &zeroConf},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Empty(t, result.Errors)

	// Verify confidence=0 is preserved (not defaulted to 1.0)
	require.Len(t, vectorDB.SaveBatchLastFacts, 1)
	assert.Equal(t, 0.0, vectorDB.SaveBatchLastFacts[0].Confidence)
}

func TestImportService_Import_ConfidenceUnset(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"}, // No confidence
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)

	// Verify confidence defaults to 1.0 when not set
	require.Len(t, vectorDB.SaveBatchLastFacts, 1)
	assert.Equal(t, 1.0, vectorDB.SaveBatchLastFacts[0].Confidence)
}

func TestImportService_Import_DryRun(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{DryRun: true})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Zero(t, vectorDB.SaveBatchCallCount, "SaveBatch should not be called in dry run")
}

func TestImportService_Import_SkipExisting(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{
		Facts: []entities.Fact{{ID: "existing-id"}},
	}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{ID: "existing-id", Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
		{ID: "new-id", Type: "character", Subject: "Frodo", Predicate: "is", Object: "hobbit"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictSkip})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 1, result.Skipped)
}

func TestImportService_Import_OverwritePreservesCreatedAt(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	originalTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	vectorDB := &mocks.VectorDB{
		Facts: []entities.Fact{
			{ID: "existing-id", CreatedAt: originalTime},
		},
	}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{ID: "existing-id", Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
		{ID: "new-id", Type: "character", Subject: "Frodo", Predicate: "is", Object: "hobbit"},
	}

	result, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Imported)

	// Verify existing fact preserves original CreatedAt
	require.Len(t, vectorDB.SaveBatchLastFacts, 2)
	var existingFact, newFact entities.Fact
	for _, f := range vectorDB.SaveBatchLastFacts {
		if f.ID == "existing-id" {
			existingFact = f
		} else {
			newFact = f
		}
	}

	assert.Equal(t, originalTime, existingFact.CreatedAt, "existing fact should preserve original CreatedAt")
	assert.True(t, newFact.CreatedAt.After(originalTime), "new fact should have recent CreatedAt")
}

func TestImportService_Import_EmptyInput(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	result, err := service.Import(context.Background(), []parsers.RawFact{}, ImportOptions{})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportService_Import_EmbeddingError(t *testing.T) {
	embedder := &mocks.Embedder{Err: assert.AnError}
	vectorDB := &mocks.VectorDB{}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
	}

	_, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "generating embeddings")
}

func TestImportService_Import_SaveError(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{Err: assert.AnError}

	service := NewImportService(embedder, vectorDB)
	rawFacts := []parsers.RawFact{
		{Type: "character", Subject: "Gandalf", Predicate: "is", Object: "wizard"},
	}

	_, err := service.Import(context.Background(), rawFacts, ImportOptions{OnConflict: ConflictOverwrite})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "saving facts")
}

func TestImportError_Error(t *testing.T) {
	t.Run("with line number", func(t *testing.T) {
		err := ImportError{Line: 5, Message: "invalid type"}
		assert.Equal(t, "line 5: invalid type", err.Error())
	})

	t.Run("without line number", func(t *testing.T) {
		err := ImportError{Line: 0, Message: "general error"}
		assert.Equal(t, "general error", err.Error())
	})
}
