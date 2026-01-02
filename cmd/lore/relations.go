package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
)

//nolint:unused // Used by newRelationsCmd
type relationsFlags struct {
	relType string
	depth   int
	format  string
}

//nolint:unused // Will be registered in main.go (Task 11)
func newRelationsCmd() *cobra.Command {
	var flags relationsFlags

	cmd := &cobra.Command{
		Use:   "relations <fact-id>",
		Short: "List relationships for a fact",
		Long:  "Shows all relationships connected to a fact, with optional depth traversal.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelations(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.relType, "type", "", "Filter by relationship type")
	cmd.Flags().IntVar(&flags.depth, "depth", 1, "Traversal depth (1-5)")
	cmd.Flags().StringVar(&flags.format, "format", "tree", "Output format: tree, list, json")

	return cmd
}

//nolint:unused // Called by newRelationsCmd
func runRelations(cmd *cobra.Command, args []string, flags relationsFlags) error {
	ctx := cmd.Context()
	factID := args[0]

	// Validate depth
	if flags.depth < 1 || flags.depth > 5 {
		return errors.New("depth must be between 1 and 5")
	}

	// Validate format
	validFormats := map[string]bool{"tree": true, "list": true, "json": true}
	if !validFormats[flags.format] {
		return fmt.Errorf("invalid format: %s (valid: tree, list, json)", flags.format)
	}

	return withRelationshipHandler(func(handler *handlers.RelationshipHandler) error {
		opts := handlers.ListOptions{
			Type:  flags.relType,
			Depth: flags.depth,
		}

		result, err := handler.HandleList(ctx, factID, opts)
		if err != nil {
			return fmt.Errorf("listing relationships: %w", err)
		}

		if len(result.Relationships) == 0 {
			fmt.Printf("No relationships found for fact: %s\n", factID)
			return nil
		}

		return printRelations(factID, result, flags.format)
	})
}

//nolint:unused // Called by runRelations
func printRelations(factID string, result *handlers.ListResult, format string) error {
	switch format {
	case "json":
		return printRelationsJSON(result)
	case "list":
		return printRelationsList(factID, result)
	case "tree":
		return printRelationsTree(factID, result)
	default:
		return printRelationsTree(factID, result)
	}
}

//nolint:unused // Called by printRelations
func printRelationsJSON(result *handlers.ListResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

//nolint:unused // Called by printRelations
func printRelationsList(factID string, result *handlers.ListResult) error {
	fmt.Printf("Relationships for %s:\n", factID)
	fmt.Println(strings.Repeat("-", 60))

	for _, info := range result.Relationships {
		rel := info.Relationship
		sourceName := getFactName(info.SourceFact)
		targetName := getFactName(info.TargetFact)

		direction := "->"
		if rel.Bidirectional {
			direction = "<->"
		}

		fmt.Printf("%s (%s) %s [%s] %s %s (%s)\n",
			rel.SourceFactID, sourceName,
			direction,
			rel.Type,
			direction,
			rel.TargetFactID, targetName,
		)
	}
	return nil
}

//nolint:unused // Called by printRelations
func printRelationsTree(factID string, result *handlers.ListResult) error {
	// Find the root fact name
	rootName := factID
	for _, info := range result.Relationships {
		if info.SourceFact != nil && info.SourceFact.ID == factID {
			rootName = getFactName(info.SourceFact)
			break
		}
		if info.TargetFact != nil && info.TargetFact.ID == factID {
			rootName = getFactName(info.TargetFact)
			break
		}
	}

	fmt.Printf("%s (%s)\n", factID, rootName)

	for i, info := range result.Relationships {
		rel := info.Relationship
		isLast := i == len(result.Relationships)-1

		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		// Determine the "other" fact (not the root)
		var otherID, otherName string
		if rel.SourceFactID == factID {
			otherID = rel.TargetFactID
			otherName = getFactName(info.TargetFact)
		} else {
			otherID = rel.SourceFactID
			otherName = getFactName(info.SourceFact)
		}

		dirIndicator := ""
		if rel.Bidirectional {
			dirIndicator = " <->"
		}

		fmt.Printf("%s %s%s → %s (%s)\n", prefix, rel.Type, dirIndicator, otherID, otherName)
	}

	return nil
}

//nolint:unused // Helper function
func getFactName(fact *entities.Fact) string {
	if fact == nil {
		return "unknown"
	}
	if fact.Subject != "" {
		return fact.Subject
	}
	return fact.ID
}
