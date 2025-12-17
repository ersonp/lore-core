package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

type deleteFlags struct {
	source string
	all    bool
	force  bool
}

func newDeleteCmd() *cobra.Command {
	var flags deleteFlags

	cmd := &cobra.Command{
		Use:   "delete [fact-id]",
		Short: "Delete facts",
		Long:  "Deletes facts by ID, source file, or all facts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd, args, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.source, "source", "s", "", "Delete all facts from source file")
	cmd.Flags().BoolVarP(&flags.all, "all", "a", false, "Delete all facts")
	cmd.Flags().BoolVarP(&flags.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string, flags deleteFlags) error {
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
	case flags.all:
		if !flags.force {
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

	case flags.source != "":
		facts, _ := repo.ListBySource(ctx, flags.source, MaxDeleteBatchSize)
		if len(facts) == 0 {
			fmt.Printf("No facts found from source: %s\n", flags.source)
			return nil
		}
		if !flags.force {
			if !confirmAction(fmt.Sprintf("Delete %d facts from %s?", len(facts), flags.source)) {
				fmt.Println("Cancelled.")
				return nil
			}
		}
		if err := repo.DeleteBySource(ctx, flags.source); err != nil {
			return fmt.Errorf("deleting facts by source: %w", err)
		}
		fmt.Printf("Deleted %d facts from %s\n", len(facts), flags.source)

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
