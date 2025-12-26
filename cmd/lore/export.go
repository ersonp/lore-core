package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
)

type exportFlags struct {
	format     string
	output     string
	factType   string
	sourceFile string
	limit      int
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
	cmd.Flags().StringVarP(&flags.sourceFile, "source", "s", "", "Filter by source file")
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

	ctx := cmd.Context()

	return withRepo(func(repo ports.VectorDB) error {
		e := &exporter{
			repo:   repo,
			format: flags.format,
			output: flags.output,
		}

		facts, err := e.fetchFacts(ctx, flags.factType, flags.sourceFile, flags.limit)
		if err != nil {
			return err
		}

		return e.export(facts)
	})
}

func (e *exporter) fetchFacts(ctx context.Context, factType, sourceFile string, limit int) ([]entities.Fact, error) {
	var facts []entities.Fact
	var err error

	switch {
	case factType != "":
		facts, err = e.repo.ListByType(ctx, entities.FactType(factType), limit)
	case sourceFile != "":
		facts, err = e.repo.ListBySource(ctx, sourceFile, limit)
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

func (e *exporter) export(facts []entities.Fact) (err error) {
	var w io.Writer
	var f *os.File

	if e.output != "" {
		f, err = os.OpenFile(e.output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("creating file: %w", err)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = fmt.Errorf("closing file: %w", cerr)
			}
		}()
		w = f
	} else {
		w = os.Stdout
	}

	if err := e.formatFacts(w, facts); err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if e.output != "" {
		fmt.Printf("Exported %d facts to %s\n", len(facts), e.output)
	}

	return nil
}

func (e *exporter) formatFacts(w io.Writer, facts []entities.Fact) error {
	switch e.format {
	case "json":
		return formatJSON(w, facts)
	case "csv":
		return formatCSV(w, facts)
	case "markdown":
		return formatMarkdown(w, facts)
	default:
		return fmt.Errorf("unknown format: %s", e.format)
	}
}

func formatJSON(w io.Writer, facts []entities.Fact) error {
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

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exportFacts)
}

func formatCSV(w io.Writer, facts []entities.Fact) error {
	writer := csv.NewWriter(w)

	header := []string{"id", "type", "subject", "predicate", "object", "context", "source_file", "confidence"}
	if err := writer.Write(header); err != nil {
		return err
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
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func formatMarkdown(w io.Writer, facts []entities.Fact) error {
	if _, err := fmt.Fprintf(w, "# Exported Facts\n\nTotal: %d facts\n\n", len(facts)); err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "| Type | Subject | Predicate | Object | Source |\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "|------|---------|-----------|--------|--------|\n"); err != nil {
		return err
	}

	for _, f := range facts {
		source := f.SourceFile
		if len(source) > 30 {
			source = "..." + source[len(source)-27:]
		}
		if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n",
			f.Type,
			escapeMarkdown(f.Subject),
			escapeMarkdown(f.Predicate),
			escapeMarkdown(f.Object),
			escapeMarkdown(source),
		); err != nil {
			return err
		}
	}

	return nil
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
