package handlers

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
)

// EntityHandler handles entity operations at the application layer.
type EntityHandler struct {
	entityService *services.EntityService
}

// NewEntityHandler creates a new EntityHandler.
func NewEntityHandler(entityService *services.EntityService) *EntityHandler {
	return &EntityHandler{
		entityService: entityService,
	}
}

// EntityListResult contains the result of listing entities.
type EntityListResult struct {
	Entities []*entities.Entity `json:"entities"`
	Total    int                `json:"total"`
}

// HandleList returns all entities for a world with pagination.
func (h *EntityHandler) HandleList(ctx context.Context, worldID string, limit, offset int) (*EntityListResult, error) {
	entitiesList, err := h.entityService.List(ctx, worldID, limit, offset)
	if err != nil {
		return nil, err
	}

	count, err := h.entityService.Count(ctx, worldID)
	if err != nil {
		return nil, err
	}

	return &EntityListResult{
		Entities: entitiesList,
		Total:    count,
	}, nil
}

// HandleSearch searches entities by name pattern.
func (h *EntityHandler) HandleSearch(ctx context.Context, worldID, query string, limit int) (*EntityListResult, error) {
	entitiesList, err := h.entityService.Search(ctx, worldID, query, limit)
	if err != nil {
		return nil, err
	}

	return &EntityListResult{
		Entities: entitiesList,
		Total:    len(entitiesList),
	}, nil
}

// HandleDelete removes an entity and its relationships.
func (h *EntityHandler) HandleDelete(ctx context.Context, entityID string) error {
	return h.entityService.Delete(ctx, entityID)
}

// HandleCount returns the number of entities in a world.
func (h *EntityHandler) HandleCount(ctx context.Context, worldID string) (int, error) {
	return h.entityService.Count(ctx, worldID)
}
