package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/application/handlers"
	"github.com/ersonp/lore-core/internal/domain/entities"
)

func newTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Manage entity types",
		Long:  "List, add, or remove custom entity types for this world.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypesList(cmd)
		},
	}

	cmd.AddCommand(newTypesListCmd())
	cmd.AddCommand(newTypesAddCmd())
	cmd.AddCommand(newTypesRemoveCmd())
	cmd.AddCommand(newTypesDescribeCmd())

	return cmd
}

func newTypesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all entity types",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypesList(cmd)
		},
	}
}

func runTypesList(cmd *cobra.Command) error {
	ctx := cmd.Context()

	return withInternalDeps(func(d *internalDeps) error {
		handler := handlers.NewEntityTypeHandler(d.entityTypeService)

		types, err := handler.HandleList(ctx)
		if err != nil {
			return fmt.Errorf("listing types: %w", err)
		}

		if len(types) == 0 {
			fmt.Println("No entity types found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION\tDEFAULT")
		for i := range types {
			isDefault := ""
			if entities.IsDefaultType(types[i].Name) {
				isDefault = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", types[i].Name, truncate(types[i].Description, 50), isDefault)
		}
		w.Flush()

		return nil
	})
}

func newTypesAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <description>",
		Short: "Add a custom entity type",
		Long:  "Add a new custom entity type. Name must be lowercase with underscores.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypesAdd(cmd, args[0], args[1])
		},
	}
}

func runTypesAdd(cmd *cobra.Command, name, description string) error {
	ctx := cmd.Context()

	return withInternalDeps(func(d *internalDeps) error {
		handler := handlers.NewEntityTypeHandler(d.entityTypeService)

		if err := handler.HandleAdd(ctx, name, description); err != nil {
			return fmt.Errorf("adding type: %w", err)
		}

		fmt.Printf("Added entity type: %s\n", name)
		return nil
	})
}

func newTypesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a custom entity type",
		Long:  "Remove a custom entity type. Default types cannot be removed.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypesRemove(cmd, args[0])
		},
	}
}

func runTypesRemove(cmd *cobra.Command, name string) error {
	ctx := cmd.Context()

	return withInternalDeps(func(d *internalDeps) error {
		handler := handlers.NewEntityTypeHandler(d.entityTypeService)

		if err := handler.HandleRemove(ctx, name); err != nil {
			return fmt.Errorf("removing type: %w", err)
		}

		fmt.Printf("Removed entity type: %s\n", name)
		return nil
	})
}

func newTypesDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <name>",
		Short: "Show details about an entity type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTypesDescribe(cmd, args[0])
		},
	}
}

func runTypesDescribe(cmd *cobra.Command, name string) error {
	ctx := cmd.Context()

	return withInternalDeps(func(d *internalDeps) error {
		handler := handlers.NewEntityTypeHandler(d.entityTypeService)

		et, err := handler.HandleDescribe(ctx, name)
		if err != nil {
			return fmt.Errorf("describing type: %w", err)
		}
		if et == nil {
			return fmt.Errorf("entity type %q not found", name)
		}

		fmt.Printf("Name:        %s\n", et.Name)
		fmt.Printf("Description: %s\n", et.Description)
		fmt.Printf("Default:     %v\n", entities.IsDefaultType(et.Name))
		if !et.CreatedAt.IsZero() {
			fmt.Printf("Created:     %s\n", et.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	})
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
