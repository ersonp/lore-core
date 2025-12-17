package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

var (
	exportFormat string
	exportOutput string
	exportType   string
	exportSource string
	exportLimit  int
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export facts to file",
		Long:  "Exports facts to JSON, CSV, or markdown format.",
		RunE:  runExport,
	}

	cmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format (json, csv, markdown)")
	cmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&exportType, "type", "t", "", "Filter by fact type")
	cmd.Flags().StringVarP(&exportSource, "source", "s", "", "Filter by source file")
	cmd.Flags().IntVarP(&exportLimit, "limit", "l", 1000, "Maximum number of facts to export")

	return cmd
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	validFormats := []string{"json", "csv", "markdown"}
	if !contains(validFormats, exportFormat) {
		return fmt.Errorf("invalid format %q, valid formats: %v", exportFormat, validFormats)
	}

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

	var facts []entities.Fact

	switch {
	case exportType != "":
		if !isValidType(exportType) {
			return fmt.Errorf("invalid type %q, valid types: %v", exportType, validTypes)
		}
		facts, err = repo.ListByType(ctx, entities.FactType(exportType), exportLimit)
	case exportSource != "":
		facts, err = repo.ListBySource(ctx, exportSource, exportLimit)
	default:
		facts, err = repo.List(ctx, exportLimit, 0)
	}

	if err != nil {
		return fmt.Errorf("listing facts: %w", err)
	}

	if len(facts) == 0 {
		return fmt.Errorf("no facts found to export")
	}

	var output string
	switch exportFormat {
	case "json":
		output, err = formatJSON(facts)
	case "csv":
		output, err = formatCSV(facts)
	case "markdown":
		output, err = formatMarkdown(facts)
	}

	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		fmt.Printf("Exported %d facts to %s\n", len(facts), exportOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

func formatJSON(facts []entities.Fact) (string, error) {
	type exportFact struct {
		ID         string  `json:"id"`
		Type       string  `json:"type"`
		Subject    string  `json:"subject"`
		Predicate  string  `json:"predicate"`
		Object     string  `json:"object"`
		Context    string  `json:"context,omitempty"`
		SourceFile string  `json:"source_file,omitempty"`
		Confidence float64 `json:"confidence"`
	}

	exportFacts := make([]exportFact, 0, len(facts))
	for _, f := range facts {
		exportFacts = append(exportFacts, exportFact{
			ID:         f.ID,
			Type:       string(f.Type),
			Subject:    f.Subject,
			Predicate:  f.Predicate,
			Object:     f.Object,
			Context:    f.Context,
			SourceFile: f.SourceFile,
			Confidence: f.Confidence,
		})
	}

	data, err := json.MarshalIndent(exportFacts, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data) + "\n", nil
}

func formatCSV(facts []entities.Fact) (string, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	header := []string{"id", "type", "subject", "predicate", "object", "context", "source_file", "confidence"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	for _, f := range facts {
		row := []string{
			f.ID,
			string(f.Type),
			f.Subject,
			f.Predicate,
			f.Object,
			f.Context,
			f.SourceFile,
			fmt.Sprintf("%.2f", f.Confidence),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return buf.String(), writer.Error()
}

func formatMarkdown(facts []entities.Fact) (string, error) {
	var buf strings.Builder

	buf.WriteString("# Exported Facts\n\n")
	buf.WriteString(fmt.Sprintf("Total: %d facts\n\n", len(facts)))

	buf.WriteString("| Type | Subject | Predicate | Object | Source |\n")
	buf.WriteString("|------|---------|-----------|--------|--------|\n")

	for _, f := range facts {
		source := f.SourceFile
		if len(source) > 30 {
			source = "..." + source[len(source)-27:]
		}
		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			f.Type,
			escapeMarkdown(f.Subject),
			escapeMarkdown(f.Predicate),
			escapeMarkdown(f.Object),
			escapeMarkdown(source),
		))
	}

	return buf.String(), nil
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
