package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var (
	ingestRecursive bool
	ingestPattern   string
	ingestCheck     bool
	ingestCheckOnly bool
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
	cmd.Flags().BoolVarP(&ingestCheck, "check", "c", false, "Check for consistency with existing facts")
	cmd.Flags().BoolVar(&ingestCheckOnly, "check-only", false, "Check consistency without saving (dry run)")

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

	opts := handlers.IngestOptions{
		CheckConsistency: ingestCheck || ingestCheckOnly,
		CheckOnly:        ingestCheckOnly,
	}

	if handlers.IsDirectory(path) {
		return runIngestDirectory(ctx, ingestHandler, path, opts)
	}

	return runIngestFile(ctx, ingestHandler, path, opts)
}

func runIngestFile(ctx context.Context, handler *handlers.IngestHandler, filePath string, opts handlers.IngestOptions) error {
	fmt.Printf("Ingesting %s...\n", filePath)

	result, err := handler.HandleWithOptions(ctx, filePath, opts)
	if err != nil {
		return fmt.Errorf("ingesting file: %w", err)
	}

	fmt.Printf("Found %d facts\n", result.FactsCount)

	for i, fact := range result.Facts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
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

func runIngestDirectory(ctx context.Context, handler *handlers.IngestHandler, dirPath string, opts handlers.IngestOptions) error {
	fmt.Printf("Ingesting directory %s (pattern: %s, recursive: %v)...\n", dirPath, ingestPattern, ingestRecursive)

	progressFn := func(file string) {
		fmt.Printf("  Processing: %s\n", file)
	}

	result, err := handler.HandleDirectoryWithOptions(ctx, dirPath, ingestPattern, ingestRecursive, progressFn, opts)
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

	for _, issue := range issues {
		severityLabel := formatSeverity(issue.Severity)
		fmt.Printf("%s: %s\n", severityLabel, issue.Description)
		fmt.Printf("  New:      %s %s %s (%s)\n",
			issue.NewFact.Subject, issue.NewFact.Predicate, issue.NewFact.Object, issue.NewFact.SourceFile)
		fmt.Printf("  Existing: %s %s %s (%s)\n\n",
			issue.ExistingFact.Subject, issue.ExistingFact.Predicate, issue.ExistingFact.Object, issue.ExistingFact.SourceFile)
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
