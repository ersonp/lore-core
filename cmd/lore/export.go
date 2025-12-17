package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

type exportFlags struct {
	format   string
	output   string
	factType string
	source   string
	limit    int
}

type exporter struct {
	repo   ports.VectorDB
	format string
	output string
}

func newExportCmd() *cobra.Command {
	var flags exportFlags

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export facts to file",
		Long:  "Exports facts to JSON, CSV, or markdown format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.format, "format", "f", "json", "Output format (json, csv, markdown)")
	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVarP(&flags.factType, "type", "t", "", "Filter by fact type")
	cmd.Flags().StringVarP(&flags.source, "source", "s", "", "Filter by source file")
	cmd.Flags().IntVarP(&flags.limit, "limit", "l", DefaultExportLimit, "Maximum number of facts to export")

	return cmd
}

func runExport(cmd *cobra.Command, flags exportFlags) error {
	if !contains(validFormats, flags.format) {
		return fmt.Errorf("invalid format %q, valid formats: %v", flags.format, validFormats)
	}

	if flags.factType != "" && !isValidType(flags.factType) {
		return fmt.Errorf("invalid type %q, valid types: %v", flags.factType, validTypes)
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

	e := &exporter{
		repo:   repo,
		format: flags.format,
		output: flags.output,
	}

	ctx := cmd.Context()
	facts, err := e.fetchFacts(ctx, flags.factType, flags.source, flags.limit)
	if err != nil {
		return err
	}

	return e.export(facts)
}

func (e *exporter) fetchFacts(ctx context.Context, factType, source string, limit int) ([]entities.Fact, error) {
	var facts []entities.Fact
	var err error

	switch {
	case factType != "":
		facts, err = e.repo.ListByType(ctx, entities.FactType(factType), limit)
	case source != "":
		facts, err = e.repo.ListBySource(ctx, source, limit)
	default:
		facts, err = e.repo.List(ctx, limit, 0)
	}

	if err != nil {
		return nil, fmt.Errorf("listing facts: %w", err)
	}

	if len(facts) == 0 {
		return nil, fmt.Errorf("no facts found to export")
	}

	return facts, nil
}

func (e *exporter) export(facts []entities.Fact) error {
	output, err := e.formatFacts(facts)
	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if e.output != "" {
		if err := os.WriteFile(e.output, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		fmt.Printf("Exported %d facts to %s\n", len(facts), e.output)
		return nil
	}

	fmt.Print(output)
	return nil
}

func (e *exporter) formatFacts(facts []entities.Fact) (string, error) {
	switch e.format {
	case "json":
		return formatJSON(facts)
	case "csv":
		return formatCSV(facts)
	case "markdown":
		return formatMarkdown(facts)
	default:
		return "", fmt.Errorf("unknown format: %s", e.format)
	}
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
