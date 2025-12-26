package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/domain/services"
)

func TestNewIngestHandler(t *testing.T) {
	llm := &mocks.LLMClient{}
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	require.NotNil(t, handler)
}

func TestIngestHandler_Handle(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Frodo is a hobbit from the Shire."), 0644)
	require.NoError(t, err)

	// Setup mocks
	llm := &mocks.LLMClient{
		Facts: []entities.Fact{
			{
				Type:      entities.FactTypeCharacter,
				Subject:   "Frodo",
				Predicate: "is a",
				Object:    "hobbit",
			},
		},
	}
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	result, err := handler.Handle(t.Context(), testFile)

	require.NoError(t, err)
	assert.Equal(t, testFile, result.FilePath)
	assert.Equal(t, 1, result.FactsCount)
	assert.Len(t, result.Facts, 1)
	assert.Equal(t, "Frodo", result.Facts[0].Subject)
}

func TestIngestHandler_HandleWithOptions_CheckOnly(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("Gandalf is a wizard."), 0644)
	require.NoError(t, err)

	llm := &mocks.LLMClient{
		Facts: []entities.Fact{
			{
				Type:      entities.FactTypeCharacter,
				Subject:   "Gandalf",
				Predicate: "is a",
				Object:    "wizard",
			},
		},
	}
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	opts := IngestOptions{CheckOnly: true}
	result, err := handler.HandleWithOptions(t.Context(), testFile, opts)

	require.NoError(t, err)
	assert.Equal(t, 1, result.FactsCount)
	// In check-only mode, facts are extracted but not saved
	assert.Equal(t, 0, db.SaveBatchCallCount, "SaveBatch should not be called in check-only mode")
}

func TestIngestHandler_Handle_FileNotFound(t *testing.T) {
	llm := &mocks.LLMClient{}
	emb := &mocks.Embedder{}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	_, err := handler.Handle(t.Context(), "/nonexistent/file.txt")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "accessing file")
}

func TestIngestHandler_Handle_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	llm := &mocks.LLMClient{}
	emb := &mocks.Embedder{}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	_, err := handler.Handle(t.Context(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory")
}

func TestIngestHandler_HandleDirectory(t *testing.T) {
	// Create temp directory with files
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("Content 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("Content 2"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "other.md"), []byte("Other content"), 0644)
	require.NoError(t, err)

	llm := &mocks.LLMClient{
		Facts: []entities.Fact{
			{
				Type:      entities.FactTypeCharacter,
				Subject:   "Test",
				Predicate: "is",
				Object:    "content",
			},
		},
	}
	emb := &mocks.Embedder{EmbeddingResult: []float32{0.1, 0.2, 0.3}}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	var progressFiles []string
	progressFn := func(file string) {
		progressFiles = append(progressFiles, file)
	}

	result, err := handler.HandleDirectory(t.Context(), tmpDir, "*.txt", false, progressFn)

	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalFiles)
	assert.Equal(t, 2, result.TotalFacts)
	assert.Len(t, progressFiles, 2)
}

func TestIngestHandler_HandleDirectory_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "file.md"), []byte("Content"), 0644)
	require.NoError(t, err)

	llm := &mocks.LLMClient{}
	emb := &mocks.Embedder{}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	_, err = handler.HandleDirectory(t.Context(), tmpDir, "*.txt", false, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no files matching")
}

func TestIngestHandler_HandleDirectory_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(testFile, []byte("Content"), 0644)
	require.NoError(t, err)

	llm := &mocks.LLMClient{}
	emb := &mocks.Embedder{}
	db := &mocks.VectorDB{}

	svc := services.NewExtractionService(llm, emb, db)
	handler := NewIngestHandler(svc)

	_, err = handler.HandleDirectory(t.Context(), testFile, "*.txt", false, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(testFile, []byte("Content"), 0644)
	require.NoError(t, err)

	assert.True(t, IsDirectory(tmpDir))
	assert.False(t, IsDirectory(testFile))
	assert.False(t, IsDirectory("/nonexistent/path"))
}

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"*.txt", true},
		{"file?.txt", true},
		{"[abc].txt", true},
		{"file.txt", false},
		{"/path/to/file.txt", false},
		{"**/*.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsGlobPattern(tt.path))
		})
	}
}
