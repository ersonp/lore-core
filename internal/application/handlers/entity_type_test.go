package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ersonp/lore-core/internal/domain/mocks"
	"github.com/ersonp/lore-core/internal/domain/services"
)

func newTestEntityTypeHandler() *EntityTypeHandler {
	db := mocks.NewRelationalDB()
	svc := services.NewEntityTypeService(db)
	return NewEntityTypeHandler(svc)
}

func newTestEntityTypeHandlerWithDefaults(t *testing.T) *EntityTypeHandler {
	t.Helper()
	db := mocks.NewRelationalDB()
	svc := services.NewEntityTypeService(db)
	err := svc.LoadDefaults(context.Background())
	require.NoError(t, err)
	return NewEntityTypeHandler(svc)
}

func TestEntityTypeHandler_HandleList_Empty(t *testing.T) {
	handler := newTestEntityTypeHandler()

	types, err := handler.HandleList(context.Background())
	require.NoError(t, err)
	assert.Empty(t, types)
}

func TestEntityTypeHandler_HandleList_WithDefaults(t *testing.T) {
	handler := newTestEntityTypeHandlerWithDefaults(t)

	types, err := handler.HandleList(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 6)

	// Verify default types are present
	typeNames := make(map[string]bool)
	for _, et := range types {
		typeNames[et.Name] = true
	}
	assert.True(t, typeNames["character"])
	assert.True(t, typeNames["location"])
	assert.True(t, typeNames["event"])
	assert.True(t, typeNames["relationship"])
	assert.True(t, typeNames["rule"])
	assert.True(t, typeNames["timeline"])
}

func TestEntityTypeHandler_HandleAdd(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleAdd(context.Background(), "weapon", "Weapons and artifacts")
	require.NoError(t, err)

	types, err := handler.HandleList(context.Background())
	require.NoError(t, err)
	assert.Len(t, types, 1)
	assert.Equal(t, "weapon", types[0].Name)
	assert.Equal(t, "Weapons and artifacts", types[0].Description)
}

func TestEntityTypeHandler_HandleAdd_Duplicate(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleAdd(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	err = handler.HandleAdd(context.Background(), "weapon", "Updated description")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestEntityTypeHandler_HandleAdd_InvalidName(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleAdd(context.Background(), "Invalid-Name", "Description")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be lowercase")
}

func TestEntityTypeHandler_HandleRemove(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleAdd(context.Background(), "weapon", "Weapons")
	require.NoError(t, err)

	err = handler.HandleRemove(context.Background(), "weapon")
	require.NoError(t, err)

	types, err := handler.HandleList(context.Background())
	require.NoError(t, err)
	assert.Empty(t, types)
}

func TestEntityTypeHandler_HandleRemove_DefaultType(t *testing.T) {
	handler := newTestEntityTypeHandlerWithDefaults(t)

	err := handler.HandleRemove(context.Background(), "character")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove default")
}

func TestEntityTypeHandler_HandleRemove_NotFound(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleRemove(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEntityTypeHandler_HandleDescribe(t *testing.T) {
	handler := newTestEntityTypeHandler()

	err := handler.HandleAdd(context.Background(), "weapon", "Weapons and artifacts")
	require.NoError(t, err)

	et, err := handler.HandleDescribe(context.Background(), "weapon")
	require.NoError(t, err)
	require.NotNil(t, et)
	assert.Equal(t, "weapon", et.Name)
	assert.Equal(t, "Weapons and artifacts", et.Description)
}

func TestEntityTypeHandler_HandleDescribe_DefaultType(t *testing.T) {
	handler := newTestEntityTypeHandlerWithDefaults(t)

	et, err := handler.HandleDescribe(context.Background(), "character")
	require.NoError(t, err)
	require.NotNil(t, et)
	assert.Equal(t, "character", et.Name)
}

func TestEntityTypeHandler_HandleDescribe_NotFound(t *testing.T) {
	handler := newTestEntityTypeHandler()

	et, err := handler.HandleDescribe(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, et)
}
