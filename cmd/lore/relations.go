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

type relationsFlags struct {
	relType string
	depth   int
	format  string
}

func newRelationsCmd() *cobra.Command {
	var flags relationsFlags

	cmd := &cobra.Command{
		Use:   "relations <entity-name>",
		Short: "List relationships for an entity",
		Long: `Shows all relationships connected to an entity, with optional filtering.

Examples:
  lore relations Alice
  lore relations Alice --type ally
  lore relations "Northern Kingdom" --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelations(cmd, args, flags)
		},
	}

	cmd.Flags().StringVar(&flags.relType, "type", "", "Filter by relationship type")
	cmd.Flags().IntVar(&flags.depth, "depth", 1, "Traversal depth (1-5)")
	cmd.Flags().StringVar(&flags.format, "format", "tree", "Output format: tree, list, json")

	return cmd
}

func runRelations(cmd *cobra.Command, args []string, flags relationsFlags) error {
	ctx := cmd.Context()
	entityName := args[0]

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

		result, err := handler.HandleList(ctx, globalWorld, entityName, opts)
		if err != nil {
			return fmt.Errorf("listing relationships: %w", err)
		}

		if len(result.Relationships) == 0 {
			fmt.Printf("No relationships found for entity: %s\n", entityName)
			return nil
		}

		return printRelations(entityName, result, flags.format)
	})
}

func printRelations(entityName string, result *handlers.ListResult, format string) error {
	switch format {
	case "json":
		return printRelationsJSON(result)
	case "list":
		return printRelationsList(entityName, result)
	case "tree":
		return printRelationsTree(entityName, result)
	default:
		return printRelationsTree(entityName, result)
	}
}

func printRelationsJSON(result *handlers.ListResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printRelationsList(entityName string, result *handlers.ListResult) error {
	fmt.Printf("Relationships for %s:\n", entityName)
	fmt.Println(strings.Repeat("-", 60))

	for _, info := range result.Relationships {
		rel := info.Relationship
		sourceName := getEntityName(info.SourceEntity)
		targetName := getEntityName(info.TargetEntity)

		direction := "->"
		if rel.Bidirectional {
			direction = "<->"
		}

		fmt.Printf("%s %s [%s] %s %s\n",
			sourceName,
			direction,
			rel.Type,
			direction,
			targetName,
		)
	}
	return nil
}

func printRelationsTree(entityName string, result *handlers.ListResult) error {
	fmt.Printf("%s\n", entityName)

	for i, info := range result.Relationships {
		rel := info.Relationship
		isLast := i == len(result.Relationships)-1

		prefix := "+-"
		if isLast {
			prefix = "\\-"
		}

		// Determine the "other" entity (not the root)
		var otherName string
		sourceName := getEntityName(info.SourceEntity)
		targetName := getEntityName(info.TargetEntity)

		if strings.EqualFold(sourceName, entityName) {
			otherName = targetName
		} else {
			otherName = sourceName
		}

		dirIndicator := ""
		if rel.Bidirectional {
			dirIndicator = " <->"
		}

		fmt.Printf("%s %s%s -> %s\n", prefix, rel.Type, dirIndicator, otherName)
	}

	return nil
}

func getEntityName(entity *entities.Entity) string {
	if entity == nil {
		return "unknown"
	}
	return entity.Name
}
