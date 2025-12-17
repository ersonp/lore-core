package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var (
	ingestRecursive bool
	ingestPattern   string
)

func newIngestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingest <path>",
		Short: "Extract facts from a file or directory",
		Long:  "Reads text files, extracts facts using LLM, generates embeddings, and stores them in Qdrant.",
		Args:  cobra.ExactArgs(1),
		RunE:  runIngest,
	}

	cmd.Flags().BoolVarP(&ingestRecursive, "recursive", "r", false, "Process subdirectories recursively")
	cmd.Flags().StringVarP(&ingestPattern, "pattern", "p", "*.txt", "File pattern to match (default: *.txt)")

	return cmd
}

func runIngest(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	path := args[0]

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

	if handlers.IsDirectory(path) {
		return runIngestDirectory(ctx, ingestHandler, path)
	}

	return runIngestFile(ctx, ingestHandler, path)
}

func runIngestFile(ctx context.Context, handler *handlers.IngestHandler, filePath string) error {
	fmt.Printf("Ingesting %s...\n", filePath)

	result, err := handler.Handle(ctx, filePath)
	if err != nil {
		return fmt.Errorf("ingesting file: %w", err)
	}

	fmt.Printf("Extracted %d facts from %s\n", result.FactsCount, result.FilePath)

	for i, fact := range result.Facts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
	}

	return nil
}

func runIngestDirectory(ctx context.Context, handler *handlers.IngestHandler, dirPath string) error {
	fmt.Printf("Ingesting directory %s (pattern: %s, recursive: %v)...\n", dirPath, ingestPattern, ingestRecursive)

	progressFn := func(file string) {
		fmt.Printf("  Processing: %s\n", file)
	}

	result, err := handler.HandleDirectory(ctx, dirPath, ingestPattern, ingestRecursive, progressFn)
	if err != nil {
		return fmt.Errorf("ingesting directory: %w", err)
	}

	fmt.Printf("\nCompleted: %d files, %d facts extracted\n", result.TotalFiles, result.TotalFacts)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}

	return nil
}
