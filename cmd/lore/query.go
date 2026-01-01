package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
)

func newQueryCmd() *cobra.Command {
	var (
		limit    int
		factType string
	)

	cmd := &cobra.Command{
		Use:   "query <question>",
		Short: "Search for facts",
		Long:  "Performs semantic search to find facts matching your question.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(cmd, args[0], limit, factType)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", DefaultQueryLimit, "Maximum number of results")
	cmd.Flags().StringVarP(&factType, "type", "t", "", "Filter by fact type (character, location, event, relationship, rule, timeline)")

	return cmd
}

func runQuery(cmd *cobra.Command, query string, limit int, factType string) error {
	ctx := cmd.Context()

	return withInternalDeps(func(d *internalDeps) error {
		// Validate type flag if provided
		if factType != "" {
			if !d.entityTypeService.IsValid(ctx, factType) {
				validTypes, err := d.entityTypeService.GetValidTypes(ctx)
				if err != nil {
					return fmt.Errorf("getting valid types: %w", err)
				}
				return fmt.Errorf("invalid type %q, valid types: %s", factType, strings.Join(validTypes, ", "))
			}
		}

		result, err := executeQuery(ctx, &d.Deps, query, limit, factType)
		if err != nil {
			return fmt.Errorf("querying facts: %w", err)
		}

		printQueryResults(result)
		return nil
	})
}

func executeQuery(ctx context.Context, d *Deps, query string, limit int, factType string) (*handlers.QueryResult, error) {
	if factType != "" {
		return d.QueryHandler.HandleByType(ctx, query, entities.FactType(factType), limit)
	}
	return d.QueryHandler.Handle(ctx, query, limit)
}

func printQueryResults(result *handlers.QueryResult) {
	if len(result.Facts) == 0 {
		fmt.Println("No facts found.")
		return
	}

	fmt.Printf("Found %d facts:\n\n", len(result.Facts))

	for i := range result.Facts {
		printFact(i+1, &result.Facts[i])
	}
}

func printFact(num int, fact *entities.Fact) {
	fmt.Printf("%d. [%s] %s %s %s\n", num, fact.Type, fact.Subject, fact.Predicate, fact.Object)
	if fact.Context != "" {
		fmt.Printf("   Context: %s\n", fact.Context)
	}
	if fact.SourceFile != "" {
		fmt.Printf("   Source: %s\n", fact.SourceFile)
	}
	fmt.Println()
}
