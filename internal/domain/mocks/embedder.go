// Package mocks provides mock implementations for testing.
package mocks

import "context"

// Embedder is a mock implementation of ports.Embedder.
type Embedder struct {
	EmbeddingResult []float32
	Err             error
}

// Embed returns the configured embedding or error.
func (m *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.EmbeddingResult, nil
}

// EmbedBatch returns embeddings for multiple texts.
func (m *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = m.EmbeddingResult
	}
	return result, nil
}
