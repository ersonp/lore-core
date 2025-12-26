// Package mocks provides mock implementations for testing.
package mocks

import "context"

// CollectionManager is a mock implementation of ports.CollectionManager.
type CollectionManager struct {
	EnsureErr error
	DeleteErr error
}

// EnsureCollection returns the configured error.
func (m *CollectionManager) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	return m.EnsureErr
}

// DeleteCollection returns the configured error.
func (m *CollectionManager) DeleteCollection(ctx context.Context) error {
	return m.DeleteErr
}
