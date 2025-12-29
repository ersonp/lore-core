package handlers

import (
	"context"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
)

// EntityTypeHandler handles entity type operations.
type EntityTypeHandler struct {
	service *services.EntityTypeService
}

// NewEntityTypeHandler creates a new EntityTypeHandler.
func NewEntityTypeHandler(service *services.EntityTypeService) *EntityTypeHandler {
	return &EntityTypeHandler{
		service: service,
	}
}

// HandleList returns all entity types.
func (h *EntityTypeHandler) HandleList(ctx context.Context) ([]entities.EntityType, error) {
	return h.service.List(ctx)
}

// HandleAdd creates a new custom entity type.
func (h *EntityTypeHandler) HandleAdd(ctx context.Context, name, description string) error {
	return h.service.Add(ctx, name, description)
}

// HandleRemove deletes a custom entity type.
func (h *EntityTypeHandler) HandleRemove(ctx context.Context, name string) error {
	return h.service.Remove(ctx, name)
}

// HandleDescribe returns details about a specific entity type.
func (h *EntityTypeHandler) HandleDescribe(ctx context.Context, name string) (*entities.EntityType, error) {
	types, err := h.service.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range types {
		if types[i].Name == name {
			return &types[i], nil
		}
	}
	return nil, nil
}
