// Package ports defines interfaces for external service communication.
package ports

import "context"

// CollectionManager handles vector collection lifecycle operations.
// This is separate from VectorDB because not all implementations may support
// collection management, and it keeps the VectorDB interface focused on
// data operations (CRUD).
type CollectionManager interface {
	// EnsureCollection creates the collection if it doesn't exist.
	EnsureCollection(ctx context.Context, vectorSize uint64) error

	// DeleteCollection removes the collection and all its data.
	DeleteCollection(ctx context.Context) error
}
