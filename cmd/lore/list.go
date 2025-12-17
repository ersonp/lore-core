package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

func newListCmd() *cobra.Command {
	var (
		limit    int
		factType string
		source   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all facts",
		Long:  "Lists all facts stored in the database with optional filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, limit, factType, source)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", DefaultListLimit, "Maximum number of facts to display")
	cmd.Flags().StringVarP(&factType, "type", "t", "", "Filter by fact type")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Filter by source file")

	return cmd
}

func runList(cmd *cobra.Command, limit int, factType string, source string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	_, _, repo, err := buildDependencies(cfg, globalWorld)
	if err != nil {
		return err
	}
	defer repo.Close()

	var facts []entities.Fact

	switch {
	case factType != "":
		if !isValidType(factType) {
			return fmt.Errorf("invalid type %q, valid types: %v", factType, validTypes)
		}
		facts, err = repo.ListByType(ctx, entities.FactType(factType), limit)
	case source != "":
		facts, err = repo.ListBySource(ctx, source, limit)
	default:
		facts, err = repo.List(ctx, limit, 0)
	}

	if err != nil {
		return fmt.Errorf("listing facts: %w", err)
	}

	if len(facts) == 0 {
		fmt.Println("No facts found.")
		return nil
	}

	count, err := repo.Count(ctx)
	if err != nil {
		fmt.Printf("Showing %d facts:\n\n", len(facts))
	} else {
		fmt.Printf("Showing %d of %d facts:\n\n", len(facts), count)
	}

	for _, fact := range facts {
		fmt.Printf("ID: %s\n", fact.ID)
		fmt.Printf("  [%s] %s %s %s\n", fact.Type, fact.Subject, fact.Predicate, fact.Object)
		if fact.Context != "" {
			fmt.Printf("  Context: %s\n", fact.Context)
		}
		if fact.SourceFile != "" {
			fmt.Printf("  Source: %s\n", fact.SourceFile)
		}
		fmt.Println()
	}

	return nil
}
