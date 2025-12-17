package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var (
	deleteSource string
	deleteAll    bool
	deleteForce  bool
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [fact-id]",
		Short: "Delete facts",
		Long:  "Deletes facts by ID, source file, or all facts.",
		RunE:  runDelete,
	}

	cmd.Flags().StringVarP(&deleteSource, "source", "s", "", "Delete all facts from source file")
	cmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all facts")
	cmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	switch {
	case deleteAll:
		if !deleteForce {
			count, _ := repo.Count(ctx)
			if !confirmAction(fmt.Sprintf("Delete all %d facts?", count)) {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		if err := repo.DeleteAll(ctx); err != nil {
			return fmt.Errorf("deleting all facts: %w", err)
		}
		fmt.Println("All facts deleted.")

	case deleteSource != "":
		facts, _ := repo.ListBySource(ctx, deleteSource, 1000)
		if len(facts) == 0 {
			fmt.Printf("No facts found from source: %s\n", deleteSource)
			return nil
		}
		if !deleteForce {
			if !confirmAction(fmt.Sprintf("Delete %d facts from %s?", len(facts), deleteSource)) {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		if err := repo.DeleteBySource(ctx, deleteSource); err != nil {
			return fmt.Errorf("deleting facts by source: %w", err)
		}
		fmt.Printf("Deleted %d facts from %s\n", len(facts), deleteSource)

	case len(args) > 0:
		factID := args[0]
		if err := repo.Delete(ctx, factID); err != nil {
			return fmt.Errorf("deleting fact: %w", err)
		}
		fmt.Printf("Deleted fact: %s\n", factID)

	default:
		return fmt.Errorf("specify a fact ID, --source, or --all")
	}

	return nil
}

func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
