package handlers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/domain/services"
)

func TestImportHandler_Handle_JSONFile(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "facts.json")
	content := `[{"type": "character", "subject": "Gandalf", "predicate": "is a", "object": "wizard"}]`
	require.NoError(t, os.WriteFile(jsonFile, []byte(content), 0644))

	result, err := handler.Handle(context.Background(), jsonFile, ImportOptions{
		OnConflict: services.ConflictOverwrite,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestImportHandler_Handle_CSVFile(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "facts.csv")
	content := "type,subject,predicate,object\ncharacter,Frodo,is a,hobbit\n"
	require.NoError(t, os.WriteFile(csvFile, []byte(content), 0644))

	result, err := handler.Handle(context.Background(), csvFile, ImportOptions{
		OnConflict: services.ConflictOverwrite,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
}

func TestImportHandler_Handle_AutoFormat(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "data.json")
	content := `[{"type": "location", "subject": "Mordor", "predicate": "is", "object": "dark"}]`
	require.NoError(t, os.WriteFile(jsonFile, []byte(content), 0644))

	// Test with Format="auto"
	result, err := handler.Handle(context.Background(), jsonFile, ImportOptions{
		Format:     "auto",
		OnConflict: services.ConflictOverwrite,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
}

func TestImportHandler_Handle_ExplicitFormat(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp file with .txt extension but JSON content
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "data.txt")
	content := `[{"type": "event", "subject": "Battle", "predicate": "occurred at", "object": "Helm's Deep"}]`
	require.NoError(t, os.WriteFile(txtFile, []byte(content), 0644))

	// Explicitly specify JSON format
	result, err := handler.Handle(context.Background(), txtFile, ImportOptions{
		Format:     "json",
		OnConflict: services.ConflictOverwrite,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
}

func TestImportHandler_Handle_UnsupportedFormat(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp file with unsupported extension
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "data.xml")
	require.NoError(t, os.WriteFile(txtFile, []byte("<data/>"), 0644))

	_, err := handler.Handle(context.Background(), txtFile, ImportOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestImportHandler_Handle_FileNotFound(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	_, err := handler.Handle(context.Background(), "/nonexistent/file.json", ImportOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "opening file")
}

func TestImportHandler_Handle_DryRun(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "facts.json")
	content := `[{"type": "character", "subject": "Gandalf", "predicate": "is a", "object": "wizard"}]`
	require.NoError(t, os.WriteFile(jsonFile, []byte(content), 0644))

	result, err := handler.Handle(context.Background(), jsonFile, ImportOptions{
		DryRun: true,
	})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Zero(t, vectorDB.SaveBatchCallCount, "SaveBatch should not be called in dry run")
}

func TestImportHandler_Handle_EmptyFile(t *testing.T) {
	embedder := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	vectorDB := &mocks.VectorDB{}
	service := services.NewImportService(embedder, vectorDB)
	handler := NewImportHandler(service)

	// Create temp empty JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte("[]"), 0644))

	result, err := handler.Handle(context.Background(), jsonFile, ImportOptions{})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 0, result.Skipped)
	assert.Empty(t, result.Errors)
}
