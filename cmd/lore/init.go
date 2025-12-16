package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new lore database",
		Long:  "Creates a .lore directory with default configuration and sets up the Qdrant collection.",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if config.Exists(cwd) {
		return fmt.Errorf("lore already initialized in %s", cwd)
	}

	if err := config.WriteDefault(cwd); err != nil {
		return fmt.Errorf("writing default config: %w", err)
	}

	fmt.Printf("Created %s\n", config.ConfigFilePath(cwd))

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	repo, err := qdrant.NewRepository(cfg.Qdrant)
	if err != nil {
		return fmt.Errorf("connecting to qdrant: %w", err)
	}
	defer repo.Close()

	initHandler := handlers.NewInitHandler(repo)
	_ = initHandler

	if err := repo.EnsureCollection(ctx, embedder.VectorSize); err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}

	fmt.Printf("Created Qdrant collection: %s\n", cfg.Qdrant.Collection)
	fmt.Println("Lore initialized successfully!")

	return nil
}
