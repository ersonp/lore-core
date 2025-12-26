package main

import (
	"fmt"
	"os"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	llm "github.com/ersonp/lore-core/internal/infrastructure/llm/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

// Deps holds all initialized dependencies for a command.
type Deps struct {
	Config            *config.Config
	Worlds            *config.WorldsConfig
	IngestHandler     *handlers.IngestHandler
	QueryHandler      *handlers.QueryHandler
	Repository        *qdrant.Repository
	ExtractionService *services.ExtractionService
}

// withDeps loads config and builds dependencies, then calls the provided function.
// It handles cleanup automatically.
func withDeps(fn func(*Deps) error) error {
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

	deps := &Deps{
		Config:            cfg,
		Worlds:            worlds,
		IngestHandler:     handlers.NewIngestHandler(extractionService),
		QueryHandler:      handlers.NewQueryHandler(queryService),
		Repository:        repo,
		ExtractionService: extractionService,
	}

	return fn(deps)
}

// withRepo is a simpler variant when only repository is needed.
func withRepo(fn func(*qdrant.Repository) error) error {
	return withDeps(func(d *Deps) error {
		return fn(d.Repository)
	})
}
