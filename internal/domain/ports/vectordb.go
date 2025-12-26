package ports

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
)

// VectorDB defines the interface for vector database operations.
type VectorDB interface {
	// EnsureCollection creates the collection if it doesn't exist.
	EnsureCollection(ctx context.Context, vectorSize uint64) error

	// DeleteCollection removes the collection and all its data.
	DeleteCollection(ctx context.Context) error

	// Save stores a fact with its embedding.
	Save(ctx context.Context, fact entities.Fact) error

	// SaveBatch stores multiple facts.
	SaveBatch(ctx context.Context, facts []entities.Fact) error

	// FindByID retrieves a fact by its ID.
	FindByID(ctx context.Context, id string) (entities.Fact, error)

	// Search performs a semantic search and returns similar facts.
	Search(ctx context.Context, embedding []float32, limit int) ([]entities.Fact, error)

	// SearchByType performs a semantic search filtered by fact type.
	SearchByType(ctx context.Context, embedding []float32, factType entities.FactType, limit int) ([]entities.Fact, error)

	// Delete removes a fact by its ID.
	Delete(ctx context.Context, id string) error

	// List returns all facts with pagination.
	List(ctx context.Context, limit int, offset uint64) ([]entities.Fact, error)

	// ListByType returns facts filtered by type.
	ListByType(ctx context.Context, factType entities.FactType, limit int) ([]entities.Fact, error)

	// ListBySource returns facts filtered by source file.
	ListBySource(ctx context.Context, sourceFile string, limit int) ([]entities.Fact, error)

	// DeleteBySource removes all facts from a source file.
	DeleteBySource(ctx context.Context, sourceFile string) error

	// DeleteAll removes all facts.
	DeleteAll(ctx context.Context) error

	// Count returns the total number of facts.
	Count(ctx context.Context) (uint64, error)
}
