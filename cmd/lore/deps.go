package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	llm "github.com/ersonp/lore-core/internal/infrastructure/llm/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/relationaldb/sqlite"
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
	relationalDB      *sqlite.Repository
	embedder          *embedder.Embedder
	extractionService *services.ExtractionService
	entityTypeService *services.EntityTypeService
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
		return errors.New("world is required (use --world flag)")
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

	// Initialize RelationalDB (SQLite)
	sqlitePath := config.SQLitePathForWorld(cwd, globalWorld)
	relationalDB, err := sqlite.NewRepository(config.SQLiteConfig{Path: sqlitePath})
	if err != nil {
		return fmt.Errorf("creating sqlite repository: %w", err)
	}
	defer relationalDB.Close()

	// Ensure schema exists
	ctx := context.Background()
	if err := relationalDB.EnsureSchema(ctx); err != nil {
		return fmt.Errorf("ensuring sqlite schema: %w", err)
	}

	// Auto-migrate: seed default types if table is empty
	if err := migrateDefaultEntityTypes(ctx, relationalDB); err != nil {
		return fmt.Errorf("migrating entity types: %w", err)
	}

	emb, err := embedder.NewEmbedder(cfg.Embedder)
	if err != nil {
		return fmt.Errorf("creating embedder: %w", err)
	}

	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating llm client: %w", err)
	}

	entityTypeService := services.NewEntityTypeService(relationalDB)
	extractionService := services.NewExtractionService(llmClient, emb, repo, entityTypeService)
	queryService := services.NewQueryService(emb, repo)

	deps := &internalDeps{
		Deps: Deps{
			Config:        cfg,
			Worlds:        worlds,
			IngestHandler: handlers.NewIngestHandler(extractionService),
			QueryHandler:  handlers.NewQueryHandler(queryService),
		},
		repo:              repo,
		relationalDB:      relationalDB,
		embedder:          emb,
		extractionService: extractionService,
		entityTypeService: entityTypeService,
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

// withRelationalDB provides direct relational database access.
//
//nolint:unused // Will be used by future commands (history, relationships)
func withRelationalDB(fn func(ports.RelationalDB) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		return fn(d.relationalDB)
	})
}

// withRelationshipHandler provides access to the RelationshipHandler for relationship commands.
func withRelationshipHandler(fn func(*handlers.RelationshipHandler) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		relationshipService := services.NewRelationshipService(d.repo, d.relationalDB, d.embedder)
		handler := handlers.NewRelationshipHandler(relationshipService, d.relationalDB)
		return fn(handler)
	})
}

// withEntityHandler provides access to the EntityHandler for entity commands.
func withEntityHandler(fn func(*handlers.EntityHandler) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		entityService := services.NewEntityService(d.relationalDB)
		handler := handlers.NewEntityHandler(entityService)
		return fn(handler)
	})
}

// migrateDefaultEntityTypes seeds default entity types if the table is empty.
// This provides transparent migration for worlds created before dynamic entity types.
func migrateDefaultEntityTypes(ctx context.Context, db ports.RelationalDB) error {
	existingTypes, err := db.ListEntityTypes(ctx)
	if err != nil {
		return fmt.Errorf("listing entity types: %w", err)
	}
	if len(existingTypes) > 0 {
		return nil
	}
	for _, et := range entities.DefaultEntityTypes {
		etCopy := et
		if err := db.SaveEntityType(ctx, &etCopy); err != nil {
			return fmt.Errorf("seeding entity type %s: %w", et.Name, err)
		}
	}
	return nil
}
