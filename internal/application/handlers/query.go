package handlers

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/services"
)

// QueryHandler handles fact queries.
type QueryHandler struct {
	queryService *services.QueryService
}

// NewQueryHandler creates a new query handler.
func NewQueryHandler(queryService *services.QueryService) *QueryHandler {
	return &QueryHandler{
		queryService: queryService,
	}
}

// QueryResult contains the result of a query.
type QueryResult struct {
	Query string
	Facts []entities.Fact
}

// Handle searches for facts matching the query.
func (h *QueryHandler) Handle(ctx context.Context, query string, limit int) (*QueryResult, error) {
	facts, err := h.queryService.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("searching facts: %w", err)
	}

	return &QueryResult{
		Query: query,
		Facts: facts,
	}, nil
}

// HandleByType searches for facts filtered by type.
func (h *QueryHandler) HandleByType(ctx context.Context, query string, factType entities.FactType, limit int) (*QueryResult, error) {
	facts, err := h.queryService.SearchByType(ctx, query, factType, limit)
	if err != nil {
		return nil, fmt.Errorf("searching facts by type: %w", err)
	}

	return &QueryResult{
		Query: query,
		Facts: facts,
	}, nil
}
