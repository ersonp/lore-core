package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

type ingestFlags struct {
	recursive bool
	pattern   string
	check     bool
	checkOnly bool
}

func newIngestCmd() *cobra.Command {
	var flags ingestFlags

	cmd := &cobra.Command{
		Use:   "ingest <path>",
		Short: "Extract facts from a file or directory",
		Long:  "Reads text files, extracts facts using LLM, generates embeddings, and stores them in Qdrant.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIngest(cmd, args[0], flags)
		},
	}

	cmd.Flags().BoolVarP(&flags.recursive, "recursive", "r", false, "Process subdirectories recursively")
	cmd.Flags().StringVarP(&flags.pattern, "pattern", "p", "*.txt", "File pattern to match (default: *.txt)")
	cmd.Flags().BoolVarP(&flags.check, "check", "c", false, "Check for consistency with existing facts")
	cmd.Flags().BoolVar(&flags.checkOnly, "check-only", false, "Check consistency without saving (dry run)")

	return cmd
}

func runIngest(cmd *cobra.Command, path string, flags ingestFlags) error {
	ctx := cmd.Context()

	return withDeps(func(d *Deps) error {
		opts := handlers.IngestOptions{
			CheckConsistency: flags.check || flags.checkOnly,
			CheckOnly:        flags.checkOnly,
		}

		if handlers.IsDirectory(path) {
			return runIngestDirectory(ctx, d.IngestHandler, path, flags.pattern, flags.recursive, opts)
		}

		return runIngestFile(ctx, d.IngestHandler, path, opts)
	})
}

func runIngestFile(ctx context.Context, handler *handlers.IngestHandler, filePath string, opts handlers.IngestOptions) error {
	fmt.Printf("Ingesting %s...\n", filePath)

	result, err := handler.HandleWithOptions(ctx, filePath, opts)
	if err != nil {
		return fmt.Errorf("ingesting file: %w", err)
	}

	fmt.Printf("Found %d facts\n", result.FactsCount)

	for i := range result.Facts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, result.Facts[i].Type, result.Facts[i].Subject, result.Facts[i].Predicate, result.Facts[i].Object)
	}

	// Display consistency issues if any
	if len(result.Issues) > 0 {
		fmt.Println()
		displayConsistencyIssues(result.Issues)
	}

	// Show save status
	if opts.CheckOnly {
		fmt.Printf("\nDry run - no facts saved (use --check to save with warnings)\n")
	} else {
		fmt.Printf("\nSaved %d facts to database\n", result.FactsCount)
	}

	return nil
}

func runIngestDirectory(ctx context.Context, handler *handlers.IngestHandler, dirPath string, pattern string, recursive bool, opts handlers.IngestOptions) error {
	fmt.Printf("Ingesting directory %s (pattern: %s, recursive: %v)...\n", dirPath, pattern, recursive)

	progressFn := func(file string) {
		fmt.Printf("  Processing: %s\n", file)
	}

	result, err := handler.HandleDirectoryWithOptions(ctx, dirPath, pattern, recursive, progressFn, opts)
	if err != nil {
		return fmt.Errorf("ingesting directory: %w", err)
	}

	// Collect all issues from all files
	var allIssues []ports.ConsistencyIssue
	for _, fileResult := range result.FileResults {
		allIssues = append(allIssues, fileResult.Issues...)
	}

	// Display consistency issues if any
	if len(allIssues) > 0 {
		fmt.Println()
		displayConsistencyIssues(allIssues)
	}

	// Show summary
	if opts.CheckOnly {
		fmt.Printf("\nDry run: %d files, %d facts found (not saved)\n", result.TotalFiles, result.TotalFacts)
	} else {
		fmt.Printf("\nCompleted: %d files, %d facts saved\n", result.TotalFiles, result.TotalFacts)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}

	return nil
}

func displayConsistencyIssues(issues []ports.ConsistencyIssue) {
	fmt.Printf("Consistency Issues Found: %d\n\n", len(issues))

	for i := range issues {
		severityLabel := formatSeverity(issues[i].Severity)
		fmt.Printf("%s: %s\n", severityLabel, issues[i].Description)
		fmt.Printf("  New:      %s %s %s (%s)\n",
			issues[i].NewFact.Subject, issues[i].NewFact.Predicate, issues[i].NewFact.Object, issues[i].NewFact.SourceFile)
		fmt.Printf("  Existing: %s %s %s (%s)\n\n",
			issues[i].ExistingFact.Subject, issues[i].ExistingFact.Predicate, issues[i].ExistingFact.Object, issues[i].ExistingFact.SourceFile)
	}
}

func formatSeverity(severity string) string {
	switch severity {
	case "critical":
		return "CRITICAL"
	case "major":
		return "MAJOR"
	case "minor":
		return "MINOR"
	default:
		return severity
	}
}
