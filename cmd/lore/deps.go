package main

import (
	"fmt"
	"os"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	llm "github.com/ersonp/lore-core/internal/infrastructure/llm/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

// Deps holds high-level dependencies for commands.
// Only handlers are exposed - services and repositories are internal.
type Deps struct {
	Config        *config.Config
	Worlds        *config.WorldsConfig
	IngestHandler *handlers.IngestHandler
	QueryHandler  *handlers.QueryHandler
}

// internalDeps holds all dependencies including low-level components.
// Used internally by helper functions.
type internalDeps struct {
	Deps
	repo              *qdrant.Repository
	extractionService *services.ExtractionService
}

// withDeps loads config and builds dependencies, then calls the provided function.
// It handles cleanup automatically.
func withDeps(fn func(*Deps) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		return fn(&d.Deps)
	})
}

// withInternalDeps provides access to all dependencies including low-level components.
// Used by commands that need direct repository or service access.
func withInternalDeps(fn func(*internalDeps) error) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	worlds, err := config.LoadWorlds(cwd)
	if err != nil {
		return fmt.Errorf("loading worlds: %w", err)
	}

	if globalWorld == "" {
		return fmt.Errorf("world is required (use --world flag)")
	}

	collection, err := worlds.GetCollection(globalWorld)
	if err != nil {
		return err
	}

	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return fmt.Errorf("creating qdrant repository: %w", err)
	}
	defer repo.Close()

	emb, err := embedder.NewEmbedder(cfg.Embedder)
	if err != nil {
		return fmt.Errorf("creating embedder: %w", err)
	}

	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating llm client: %w", err)
	}

	extractionService := services.NewExtractionService(llmClient, emb, repo)
	queryService := services.NewQueryService(emb, repo)

	deps := &internalDeps{
		Deps: Deps{
			Config:        cfg,
			Worlds:        worlds,
			IngestHandler: handlers.NewIngestHandler(extractionService),
			QueryHandler:  handlers.NewQueryHandler(queryService),
		},
		repo:              repo,
		extractionService: extractionService,
	}

	return fn(deps)
}

// withRepo provides direct repository access for commands that need it.
func withRepo(fn func(ports.VectorDB) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		return fn(d.repo)
	})
}

// withExtractionService provides direct service access for commands like watch.
func withExtractionService(fn func(*services.ExtractionService, ports.VectorDB) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		return fn(d.extractionService, d.repo)
	})
}
