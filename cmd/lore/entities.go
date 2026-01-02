package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
)

func newEntitiesCmd() *cobra.Command {
	var searchQuery string
	var limit int

	cmd := &cobra.Command{
		Use:   "entities",
		Short: "List entities in a world",
		Long: `List all tracked entities in a world.

Entities are subjects that have been used in relationships.
Use --search to filter by name.

Examples:
  lore entities
  lore entities --search "Ali"
  lore entities --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEntities(cmd, searchQuery, limit)
		},
	}

	cmd.Flags().StringVar(&searchQuery, "search", "", "Search entities by name")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of entities to return")

	return cmd
}

func runEntities(cmd *cobra.Command, searchQuery string, limit int) error {
	ctx := cmd.Context()

	return withEntityHandler(func(handler *handlers.EntityHandler) error {
		var result *handlers.EntityListResult
		var err error

		if searchQuery != "" {
			result, err = handler.HandleSearch(ctx, globalWorld, searchQuery, limit)
		} else {
			result, err = handler.HandleList(ctx, globalWorld, limit, 0)
		}

		if err != nil {
			return fmt.Errorf("listing entities: %w", err)
		}

		if len(result.Entities) == 0 {
			fmt.Println("No entities found.")
			return nil
		}

		fmt.Printf("Entities (%d total):\n", result.Total)
		fmt.Println()

		for _, entity := range result.Entities {
			fmt.Printf("  %-40s %s\n", entity.ID[:8]+"...", entity.Name)
		}

		return nil
	})
}
