package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/ports"
)

type deleteFlags struct {
	sourceFile string
	all        bool
	force      bool
}

type deleter struct {
	repo  ports.VectorDB
	force bool
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

	cmd.Flags().StringVarP(&flags.sourceFile, "source", "s", "", "Delete all facts from source file")
	cmd.Flags().BoolVarP(&flags.all, "all", "a", false, "Delete all facts")
	cmd.Flags().BoolVarP(&flags.force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string, flags deleteFlags) error {
	ctx := cmd.Context()

	return withRepo(func(repo ports.VectorDB) error {
		d := &deleter{
			repo:  repo,
			force: flags.force,
		}

		switch {
		case flags.all:
			return d.deleteAll(ctx)
		case flags.sourceFile != "":
			return d.deleteBySource(ctx, flags.sourceFile)
		case len(args) > 0:
			return d.deleteByID(ctx, args[0])
		default:
			return fmt.Errorf("specify a fact ID, --source, or --all")
		}
	})
}

func (d *deleter) deleteAll(ctx context.Context) error {
	if !d.force {
		prompt := d.buildDeleteAllPrompt(ctx)
		if !confirmAction(prompt) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := d.repo.DeleteAll(ctx); err != nil {
		return fmt.Errorf("deleting all facts: %w", err)
	}
	fmt.Println("All facts deleted.")
	return nil
}

func (d *deleter) buildDeleteAllPrompt(ctx context.Context) string {
	count, err := d.repo.Count(ctx)
	if err != nil {
		return "Delete all facts?"
	}
	return fmt.Sprintf("Delete all %d facts?", count)
}

func (d *deleter) deleteBySource(ctx context.Context, sourceFile string) error {
	facts, err := d.repo.ListBySource(ctx, sourceFile, MaxDeleteBatchSize)
	if err != nil {
		return fmt.Errorf("listing facts by source: %w", err)
	}

	if len(facts) == 0 {
		fmt.Printf("No facts found from source: %s\n", sourceFile)
		return nil
	}

	if !d.force && !confirmAction(fmt.Sprintf("Delete %d facts from %s?", len(facts), sourceFile)) {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := d.repo.DeleteBySource(ctx, sourceFile); err != nil {
		return fmt.Errorf("deleting facts by source: %w", err)
	}
	fmt.Printf("Deleted %d facts from %s\n", len(facts), sourceFile)
	return nil
}

func (d *deleter) deleteByID(ctx context.Context, factID string) error {
	if err := d.repo.Delete(ctx, factID); err != nil {
		return fmt.Errorf("deleting fact: %w", err)
	}
	fmt.Printf("Deleted fact: %s\n", factID)
	return nil
}

func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, _ := reader.ReadString('\n') // Error ignored: EOF/error treated as "no"
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
