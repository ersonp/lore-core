package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

// worldManager handles qdrant collection operations for worlds.
type worldManager struct {
	cfg *config.Config
}

func newWorldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worlds",
		Short: "Manage worlds",
		RunE:  runWorldsList,
	}

	cmd.AddCommand(
		newWorldsListCmd(),
		newWorldsCreateCmd(),
		newWorldsDeleteCmd(),
	)

	return cmd
}

func newWorldsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all worlds",
		RunE:  runWorldsList,
	}
}

func runWorldsList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	worlds, err := config.LoadWorlds(cwd)
	if err != nil {
		return fmt.Errorf("loading worlds: %w", err)
	}

	if len(worlds.Worlds) == 0 {
		fmt.Println("No worlds configured.")
		fmt.Println("Use 'lore worlds create NAME' to create a world.")
		return nil
	}

	fmt.Printf("%-20s %-25s %s\n", "NAME", "COLLECTION", "DESCRIPTION")
	fmt.Printf("%-20s %-25s %s\n", "----", "----------", "-----------")

	for name, world := range worlds.Worlds {
		fmt.Printf("%-20s %-25s %s\n", name, world.Collection, world.Description)
	}

	return nil
}

func newWorldsCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new world",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorldsCreate(cmd, args[0], description)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "World description")

	return cmd
}

func runWorldsCreate(cmd *cobra.Command, name string, description string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	collection := config.GenerateCollectionName(name)
	initialized := false

	// Check if config exists, if not initialize
	if !config.Exists(cwd) {
		if err := config.WriteDefaultWithWorld(cwd, name, description); err != nil {
			return fmt.Errorf("initializing config: %w", err)
		}
		fmt.Printf("Initialized lore in %s\n", config.ConfigDir(cwd))
		initialized = true
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// If not initialized, add world to existing config
	if !initialized {
		worlds, err := config.LoadWorlds(cwd)
		if err != nil {
			return fmt.Errorf("loading worlds: %w", err)
		}

		if worlds.Exists(name) {
			return fmt.Errorf("world %q already exists", name)
		}

		worlds.Add(name, config.WorldEntry{
			Collection:  collection,
			Description: description,
		})

		if err := worlds.Save(cwd); err != nil {
			return fmt.Errorf("saving worlds: %w", err)
		}
	}

	mgr := &worldManager{cfg: cfg}
	if err := mgr.createCollection(ctx, collection); err != nil {
		return fmt.Errorf("creating qdrant collection: %w", err)
	}

	fmt.Printf("Created world %q with collection %q\n", name, collection)

	return nil
}

func newWorldsDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a world",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorldsDelete(cmd, args[0], force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete even if world contains facts")

	return cmd
}

func runWorldsDelete(cmd *cobra.Command, name string, force bool) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	worlds, err := config.LoadWorlds(cwd)
	if err != nil {
		return fmt.Errorf("loading worlds: %w", err)
	}

	world, err := worlds.Get(name)
	if err != nil {
		return fmt.Errorf("world %q not found", name)
	}

	mgr := &worldManager{cfg: cfg}

	if !force {
		count, err := mgr.getCollectionCount(ctx, world.Collection)
		if err == nil && count > 0 {
			return fmt.Errorf("world %q contains %d facts, use --force to delete", name, count)
		}
	}

	if err := mgr.deleteCollection(ctx, world.Collection); err != nil {
		fmt.Printf("Warning: could not delete collection %q: %v\n", world.Collection, err)
	}

	worlds.Remove(name)

	if err := worlds.Save(cwd); err != nil {
		return fmt.Errorf("saving worlds: %w", err)
	}

	fmt.Printf("Deleted world %q\n", name)

	return nil
}

func (m *worldManager) createCollection(ctx context.Context, collection string) error {
	qdrantCfg := m.cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	return repo.EnsureCollection(ctx, embedder.VectorSize)
}

func (m *worldManager) getCollectionCount(ctx context.Context, collection string) (uint64, error) {
	qdrantCfg := m.cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return 0, err
	}
	defer repo.Close()

	return repo.Count(ctx)
}

func (m *worldManager) deleteCollection(ctx context.Context, collection string) error {
	qdrantCfg := m.cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	return repo.DeleteCollection(ctx)
}
