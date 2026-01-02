package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
)

//nolint:unused // Will be registered in main.go (Task 11)
func newRelateCmd() *cobra.Command {
	var bidirectional bool

	cmd := &cobra.Command{
		Use:   "relate <source-fact-id> <type> <target-fact-id>",
		Short: "Create a relationship between two facts",
		Long: `Creates a relationship link between two existing facts.

Valid relationship types:
  - parent, child, sibling, spouse
  - ally, enemy
  - located_in, owns, member_of, created`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelate(cmd, args, bidirectional)
		},
	}

	cmd.Flags().BoolVar(&bidirectional, "bidirectional", true, "Create bidirectional relationship")

	cmd.AddCommand(newRelateDeleteCmd())

	return cmd
}

//nolint:unused // Called by newRelateCmd
func runRelate(cmd *cobra.Command, args []string, bidirectional bool) error {
	ctx := cmd.Context()
	sourceID := args[0]
	relType := args[1]
	targetID := args[2]

	return withRelationshipHandler(func(handler *handlers.RelationshipHandler) error {
		rel, err := handler.HandleCreate(ctx, sourceID, relType, targetID, bidirectional)
		if err != nil {
			return fmt.Errorf("creating relationship: %w", err)
		}

		fmt.Printf("Created relationship: %s\n", rel.ID)
		fmt.Printf("  %s -[%s]-> %s\n", rel.SourceFactID, rel.Type, rel.TargetFactID)
		if rel.Bidirectional {
			fmt.Println("  (bidirectional)")
		}

		return nil
	})
}

//nolint:unused // Added as subcommand to relate
func newRelateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <relationship-id>",
		Short: "Delete a relationship",
		Long:  "Deletes an existing relationship by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE:  runRelateDelete,
	}
}

//nolint:unused // Called by newRelateDeleteCmd
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
