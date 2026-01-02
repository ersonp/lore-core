// Package main provides the entry point for the lore CLI application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	version     = "0.1.0-dev"
	globalWorld string
)

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
		Short:   "A factual knowledge base powered by vector search and LLM analysis",
		Version: version,
	}

	rootCmd.PersistentFlags().StringVarP(&globalWorld, "world", "w", "", "World to operate on (required)")

	rootCmd.AddCommand(
		newIngestCmd(),
		newQueryCmd(),
		newListCmd(),
		newDeleteCmd(),
		newExportCmd(),
		newImportCmd(),
		newWatchCmd(),
		newWorldsCmd(),
		newTypesCmd(),
		newRelateCmd(),
		newRelationsCmd(),
		newEntitiesCmd(),
	)

	return rootCmd.ExecuteContext(ctx)
}
