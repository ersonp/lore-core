// Package handlers contains application use case handlers.
package handlers

import (
	"context"
	"fmt"

	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
)

// InitHandler handles database initialization.
type InitHandler struct {
	vectorDB          ports.VectorDB
	collectionManager ports.CollectionManager
}

// NewInitHandler creates a new init handler.
func NewInitHandler(vectorDB ports.VectorDB, collectionManager ports.CollectionManager) *InitHandler {
	return &InitHandler{
		vectorDB:          vectorDB,
		collectionManager: collectionManager,
	}
}

// InitResult contains the result of initialization.
type InitResult struct {
	ConfigPath     string
	CollectionName string
}

// Handle initializes the lore database.
func (h *InitHandler) Handle(ctx context.Context, basePath string) (*InitResult, error) {
	if config.Exists(basePath) {
		return nil, fmt.Errorf("lore already initialized in %s", basePath)
	}

	if err := config.WriteDefault(basePath); err != nil {
		return nil, fmt.Errorf("writing default config: %w", err)
	}

	cfg, err := config.Load(basePath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if h.collectionManager != nil {
		if err := h.collectionManager.EnsureCollection(ctx, embedder.VectorSize); err != nil {
			return nil, fmt.Errorf("creating collection: %w", err)
		}
	}

	return &InitResult{
		ConfigPath:     config.ConfigFilePath(basePath),
		CollectionName: cfg.Qdrant.Collection,
	}, nil
}
