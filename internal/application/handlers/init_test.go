package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func TestNewInitHandler(t *testing.T) {
	db := &mocks.VectorDB{}
	cm := &mocks.CollectionManager{}

	handler := NewInitHandler(db, cm)

	require.NotNil(t, handler)
	assert.Equal(t, db, handler.vectorDB)
	assert.Equal(t, cm, handler.collectionManager)
}

func TestInitHandler_Handle_Success(t *testing.T) {
	tmpDir := t.TempDir()

	db := &mocks.VectorDB{}
	cm := &mocks.CollectionManager{}

	handler := NewInitHandler(db, cm)

	result, err := handler.Handle(t.Context(), tmpDir)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.ConfigPath, "config.yaml")

	// Verify config was created
	assert.True(t, config.Exists(tmpDir))
}

func TestInitHandler_Handle_AlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize first
	err := config.WriteDefault(tmpDir)
	require.NoError(t, err)

	db := &mocks.VectorDB{}
	cm := &mocks.CollectionManager{}

	handler := NewInitHandler(db, cm)

	_, err = handler.Handle(t.Context(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestInitHandler_Handle_CollectionManagerError(t *testing.T) {
	tmpDir := t.TempDir()

	db := &mocks.VectorDB{}
	cm := &mocks.CollectionManager{
		EnsureErr: errors.New("connection failed"),
	}

	handler := NewInitHandler(db, cm)

	_, err := handler.Handle(t.Context(), tmpDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating collection")
	assert.Contains(t, err.Error(), "connection failed")
}

func TestInitHandler_Handle_NilCollectionManager(t *testing.T) {
	tmpDir := t.TempDir()

	db := &mocks.VectorDB{}

	// Pass nil CollectionManager - should still work
	handler := NewInitHandler(db, nil)

	result, err := handler.Handle(t.Context(), tmpDir)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.ConfigPath, "config.yaml")
}
