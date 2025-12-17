package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var (
	listLimit  int
	listType   string
	listSource string
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all facts",
		Long:  "Lists all facts stored in the database with optional filtering.",
		RunE:  runList,
	}

	cmd.Flags().IntVarP(&listLimit, "limit", "l", 50, "Maximum number of facts to display")
	cmd.Flags().StringVarP(&listType, "type", "t", "", "Filter by fact type")
	cmd.Flags().StringVarP(&listSource, "source", "s", "", "Filter by source file")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	_, _, repo, err := buildDependencies(cfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	var facts []entities.Fact

	switch {
	case listType != "":
		if !isValidType(listType) {
			return fmt.Errorf("invalid type %q, valid types: %v", listType, validTypes)
		}
		facts, err = repo.ListByType(ctx, entities.FactType(listType), listLimit)
	case listSource != "":
		facts, err = repo.ListBySource(ctx, listSource, listLimit)
	default:
		facts, err = repo.List(ctx, listLimit, 0)
	}

	if err != nil {
		return fmt.Errorf("listing facts: %w", err)
	}

	if len(facts) == 0 {
		fmt.Println("No facts found.")
		return nil
	}

	count, _ := repo.Count(ctx)
	fmt.Printf("Showing %d of %d facts:\n\n", len(facts), count)

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
