// Package main provides the entry point for the lore CLI application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	llm "github.com/ersonp/lore-core/internal/infrastructure/llm/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

var version = "0.1.0-dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	rootCmd := &cobra.Command{
		Use:     "lore",
		Short:   "A factual consistency database for fictional worlds",
		Version: version,
	}

	rootCmd.AddCommand(
		newInitCmd(),
		newIngestCmd(),
		newQueryCmd(),
		newListCmd(),
		newDeleteCmd(),
	)

	return rootCmd.ExecuteContext(ctx)
}

// buildDependencies creates all dependencies from config.
func buildDependencies(cfg *config.Config) (*handlers.IngestHandler, *handlers.QueryHandler, *qdrant.Repository, error) {
	repo, err := qdrant.NewRepository(cfg.Qdrant)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating qdrant repository: %w", err)
	}

	emb, err := embedder.NewEmbedder(cfg.Embedder)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating embedder: %w", err)
	}

	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating llm client: %w", err)
	}

	extractionService := services.NewExtractionService(llmClient, emb, repo)
	queryService := services.NewQueryService(emb, repo)

	ingestHandler := handlers.NewIngestHandler(extractionService)
	queryHandler := handlers.NewQueryHandler(queryService)

	return ingestHandler, queryHandler, repo, nil
}
