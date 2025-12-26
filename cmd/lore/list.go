package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

func newListCmd() *cobra.Command {
	var (
		limit      int
		factType   string
		sourceFile string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all facts",
		Long:  "Lists all facts stored in the database with optional filtering.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, limit, factType, sourceFile)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", DefaultListLimit, "Maximum number of facts to display")
	cmd.Flags().StringVarP(&factType, "type", "t", "", "Filter by fact type")
	cmd.Flags().StringVarP(&sourceFile, "source", "s", "", "Filter by source file")

	return cmd
}

func runList(cmd *cobra.Command, limit int, factType string, sourceFile string) error {
	ctx := cmd.Context()

	return withRepo(func(repo ports.VectorDB) error {
		var facts []entities.Fact
		var err error

		switch {
		case factType != "":
			if !isValidType(factType) {
				return fmt.Errorf("invalid type %q, valid types: %v", factType, validTypes)
			}
			facts, err = repo.ListByType(ctx, entities.FactType(factType), limit)
		case sourceFile != "":
			facts, err = repo.ListBySource(ctx, sourceFile, limit)
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

		count, _ := repo.Count(ctx)
		displayFacts(facts, count)
		return nil
	})
}

func displayFacts(facts []entities.Fact, totalCount uint64) {
	if totalCount > 0 {
		fmt.Printf("Showing %d of %d facts:\n\n", len(facts), totalCount)
	} else {
		fmt.Printf("Showing %d facts:\n\n", len(facts))
	}

	for _, fact := range facts {
		displayFact(fact)
	}
}

func displayFact(fact entities.Fact) {
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
