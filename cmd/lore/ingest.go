package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func newIngestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ingest <file>",
		Short: "Extract facts from a file",
		Long:  "Reads a text file, extracts facts using LLM, generates embeddings, and stores them in Qdrant.",
		Args:  cobra.ExactArgs(1),
		RunE:  runIngest,
	}
}

func runIngest(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	filePath := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ingestHandler, _, repo, err := buildDependencies(cfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	fmt.Printf("Ingesting %s...\n", filePath)

	result, err := ingestHandler.Handle(ctx, filePath)
	if err != nil {
		return fmt.Errorf("ingesting file: %w", err)
	}

	fmt.Printf("Extracted %d facts from %s\n", result.FactsCount, result.FilePath)

	for i, fact := range result.Facts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
	}

	return nil
}
