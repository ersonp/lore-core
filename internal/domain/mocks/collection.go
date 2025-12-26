// Package mocks provides mock implementations for testing.
package mocks

import "context"

// CollectionManager is a mock implementation of ports.CollectionManager.
type CollectionManager struct {
	EnsureErr error
	DeleteErr error

	// Call tracking
	EnsureCollectionCallCount int
	DeleteCollectionCallCount int
}

// EnsureCollection returns the configured error.
func (m *CollectionManager) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	m.EnsureCollectionCallCount++
	return m.EnsureErr
}

// DeleteCollection returns the configured error.
func (m *CollectionManager) DeleteCollection(ctx context.Context) error {
	m.DeleteCollectionCallCount++
	return m.DeleteErr
}
