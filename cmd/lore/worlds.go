package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

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

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Worlds) == 0 {
		fmt.Println("No worlds configured.")
		fmt.Println("Use 'lore worlds create NAME' to create a world.")
		return nil
	}

	fmt.Printf("%-20s %-25s %s\n", "NAME", "COLLECTION", "DESCRIPTION")
	fmt.Printf("%-20s %-25s %s\n", "----", "----------", "-----------")

	for name, world := range cfg.Worlds {
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
		if _, exists := cfg.Worlds[name]; exists {
			return fmt.Errorf("world %q already exists", name)
		}

		newWorld := config.WorldConfig{
			Collection:  collection,
			Description: description,
		}

		if err := addWorldToConfig(cwd, name, newWorld); err != nil {
			return fmt.Errorf("adding world to config: %w", err)
		}
	}

	if err := createQdrantCollection(ctx, cfg, collection); err != nil {
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

	world, exists := cfg.Worlds[name]
	if !exists {
		return fmt.Errorf("world %q not found", name)
	}

	if !force {
		count, err := getCollectionCount(ctx, cfg, world.Collection)
		if err == nil && count > 0 {
			return fmt.Errorf("world %q contains %d facts, use --force to delete", name, count)
		}
	}

	if err := deleteQdrantCollection(ctx, cfg, world.Collection); err != nil {
		fmt.Printf("Warning: could not delete collection %q: %v\n", world.Collection, err)
	}

	if err := removeWorldFromConfig(cwd, name); err != nil {
		return fmt.Errorf("removing world from config: %w", err)
	}

	fmt.Printf("Deleted world %q\n", name)

	return nil
}

func addWorldToConfig(basePath string, name string, world config.WorldConfig) error {
	configFile := config.ConfigFilePath(basePath)

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	worlds, ok := rawConfig["worlds"].(map[string]interface{})
	if !ok {
		worlds = make(map[string]interface{})
	}

	worlds[name] = map[string]interface{}{
		"collection":  world.Collection,
		"description": world.Description,
	}
	rawConfig["worlds"] = worlds

	newData, err := yaml.Marshal(rawConfig)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, newData, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func removeWorldFromConfig(basePath string, name string) error {
	configFile := config.ConfigFilePath(basePath)

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	worlds, ok := rawConfig["worlds"].(map[string]interface{})
	if ok {
		delete(worlds, name)
		rawConfig["worlds"] = worlds
	}

	newData, err := yaml.Marshal(rawConfig)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, newData, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func createQdrantCollection(ctx context.Context, cfg *config.Config, collection string) error {
	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	return repo.EnsureCollection(ctx, embedder.VectorSize)
}

func getCollectionCount(ctx context.Context, cfg *config.Config, collection string) (uint64, error) {
	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return 0, err
	}
	defer repo.Close()

	return repo.Count(ctx)
}

func deleteQdrantCollection(ctx context.Context, cfg *config.Config, collection string) error {
	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return err
	}
	defer repo.Close()

	return repo.DeleteCollection(ctx)
}
