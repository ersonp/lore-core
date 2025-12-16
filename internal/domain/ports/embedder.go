package ports

import "context"

// Embedder defines the interface for generating vector embeddings.
type Embedder interface {
	// Embed generates a vector embedding for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates vector embeddings for multiple texts.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}
