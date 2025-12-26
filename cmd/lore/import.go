package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/services"
)

type importFlags struct {
	format     string
	dryRun     bool
	onConflict string
}

func newImportCmd() *cobra.Command {
	var flags importFlags

	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import facts from JSON or CSV",
		Long:  "Imports facts from a structured file. Generates embeddings automatically.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runImport(cmd, args[0], flags)
		},
	}

	cmd.Flags().StringVarP(&flags.format, "format", "f", "auto", "File format (json, csv, auto)")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Validate without saving")
	cmd.Flags().StringVar(&flags.onConflict, "on-conflict", "skip", "Conflict handling (skip, overwrite)")

	return cmd
}

func runImport(cmd *cobra.Command, filePath string, flags importFlags) error {
	// Validate on-conflict flag
	if flags.onConflict != "skip" && flags.onConflict != "overwrite" {
		return fmt.Errorf("invalid --on-conflict value %q (valid: skip, overwrite)", flags.onConflict)
	}

	ctx := cmd.Context()

	return withImportHandler(func(handler *handlers.ImportHandler) error {
		opts := handlers.ImportOptions{
			Format:     flags.format,
			DryRun:     flags.dryRun,
			OnConflict: flags.onConflict,
		}

		fmt.Printf("Importing %s...\n", filePath)

		result, err := handler.Handle(ctx, filePath, opts)
		if err != nil {
			return fmt.Errorf("importing file: %w", err)
		}

		// Display errors
		if len(result.Errors) > 0 {
			fmt.Printf("\nValidation errors (%d):\n", len(result.Errors))
			for _, e := range result.Errors {
				fmt.Printf("  %s\n", e.Error())
			}
		}

		// Display summary
		fmt.Println()
		if flags.dryRun {
			fmt.Printf("Dry run: %d facts would be imported", result.Imported)
		} else {
			fmt.Printf("Imported: %d facts", result.Imported)
		}

		if result.Skipped > 0 {
			fmt.Printf(", %d skipped (already exist)", result.Skipped)
		}

		if len(result.Errors) > 0 {
			fmt.Printf(", %d errors", len(result.Errors))
		}

		fmt.Println()

		return nil
	})
}

// withImportHandler creates an ImportHandler and calls the provided function.
func withImportHandler(fn func(*handlers.ImportHandler) error) error {
	return withInternalDeps(func(d *internalDeps) error {
		importService := services.NewImportService(d.embedder, d.repo)
		handler := handlers.NewImportHandler(importService)
		return fn(handler)
	})
}
