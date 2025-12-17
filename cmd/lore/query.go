package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var validTypes = []string{"character", "location", "event", "relationship", "rule", "timeline"}

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

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Maximum number of results")
	cmd.Flags().StringVarP(&factType, "type", "t", "", "Filter by fact type (character, location, event, relationship, rule, timeline)")

	return cmd
}

func runQuery(cmd *cobra.Command, query string, limit int, factType string) error {
	ctx := cmd.Context()

	if factType != "" && !isValidType(factType) {
		return fmt.Errorf("invalid type %q, valid types: %v", factType, validTypes)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	_, queryHandler, repo, err := buildDependencies(cfg, globalWorld)
	if err != nil {
		return err
	}
	defer repo.Close()

	var result *handlers.QueryResult
	if factType != "" {
		result, err = queryHandler.HandleByType(ctx, query, entities.FactType(factType), limit)
	} else {
		result, err = queryHandler.Handle(ctx, query, limit)
	}
	if err != nil {
		return fmt.Errorf("querying facts: %w", err)
	}

	if len(result.Facts) == 0 {
		fmt.Println("No facts found.")
		return nil
	}

	fmt.Printf("Found %d facts:\n\n", len(result.Facts))

	for i, fact := range result.Facts {
		fmt.Printf("%d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
		if fact.Context != "" {
			fmt.Printf("   Context: %s\n", fact.Context)
		}
		if fact.SourceFile != "" {
			fmt.Printf("   Source: %s\n", fact.SourceFile)
		}
		fmt.Println()
	}

	return nil
}

func isValidType(t string) bool {
	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}
	return false
}
