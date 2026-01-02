package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
)

func newRelateCmd() *cobra.Command {
	var bidirectional bool

	cmd := &cobra.Command{
		Use:   "relate <source-entity> <type> <target-entity>",
		Short: "Create a relationship between two entities",
		Long: `Creates a relationship link between two entities.
Entities are created automatically if they don't exist.
Use quotes for entity names with spaces.

Valid relationship types:
  - parent, child, sibling, spouse
  - ally, enemy
  - located_in, owns, member_of, created

Examples:
  lore relate Alice ally Bob
  lore relate "Northern Kingdom" located_in "The Realm"
  lore relate Alice enemy "Dark Lord" --bidirectional=false`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelate(cmd, args, bidirectional)
		},
	}

	cmd.Flags().BoolVar(&bidirectional, "bidirectional", true, "Create bidirectional relationship")

	cmd.AddCommand(newRelateDeleteCmd())

	return cmd
}

func runRelate(cmd *cobra.Command, args []string, bidirectional bool) error {
	ctx := cmd.Context()
	sourceEntity := args[0]
	relType := args[1]
	targetEntity := args[2]

	return withRelationshipHandler(func(handler *handlers.RelationshipHandler) error {
		rel, err := handler.HandleCreate(ctx, globalWorld, sourceEntity, relType, targetEntity, bidirectional)
		if err != nil {
			return fmt.Errorf("creating relationship: %w", err)
		}

		fmt.Printf("Created relationship: %s\n", rel.ID)
		fmt.Printf("  %s -[%s]-> %s\n", sourceEntity, rel.Type, targetEntity)
		if rel.Bidirectional {
			fmt.Println("  (bidirectional)")
		}

		return nil
	})
}

func newRelateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <relationship-id>",
		Short: "Delete a relationship",
		Long:  "Deletes an existing relationship by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRelateDelete,
	}
}

func runRelateDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	relID := args[0]

	return withRelationshipHandler(func(handler *handlers.RelationshipHandler) error {
		if err := handler.HandleDelete(ctx, relID); err != nil {
			return fmt.Errorf("deleting relationship: %w", err)
		}

		fmt.Printf("Deleted relationship: %s\n", relID)
		return nil
	})
}
