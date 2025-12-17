package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/domain/services"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
	embedder "github.com/ersonp/lore-core/internal/infrastructure/embedder/openai"
	llm "github.com/ersonp/lore-core/internal/infrastructure/llm/openai"
	"github.com/ersonp/lore-core/internal/infrastructure/vectordb/qdrant"
)

type watchFlags struct {
	source   string
	autoSave bool
}

func newWatchCmd() *cobra.Command {
	var flags watchFlags

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Interactive mode with real-time consistency checking",
		Long:  "Enter text interactively and get real-time fact extraction and consistency feedback.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(cmd, flags)
		},
	}

	cmd.Flags().StringVarP(&flags.source, "source", "s", "interactive", "Source name for facts")
	cmd.Flags().BoolVar(&flags.autoSave, "save", false, "Auto-save facts (with confirmation on conflicts)")

	return cmd
}

type watchState struct {
	pendingFacts      []entities.Fact
	pendingIssues     []ports.ConsistencyIssue
	extractionService *services.ExtractionService
	vectorDB          ports.VectorDB
	source            string
	autoSave          bool
}

func runWatch(cmd *cobra.Command, flags watchFlags) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if globalWorld == "" {
		return fmt.Errorf("world is required (use --world flag)")
	}

	collection, err := cfg.GetCollectionForWorld(globalWorld)
	if err != nil {
		return err
	}

	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

	// Build dependencies for watch mode
	repo, err := qdrant.NewRepository(qdrantCfg)
	if err != nil {
		return fmt.Errorf("creating qdrant repository: %w", err)
	}
	defer repo.Close()

	emb, err := embedder.NewEmbedder(cfg.Embedder)
	if err != nil {
		return fmt.Errorf("creating embedder: %w", err)
	}

	llmClient, err := llm.NewClient(cfg.LLM)
	if err != nil {
		return fmt.Errorf("creating llm client: %w", err)
	}

	extractionSvc := services.NewExtractionService(llmClient, emb, repo)

	state := &watchState{
		extractionService: extractionSvc,
		vectorDB:          repo,
		source:            flags.source,
		autoSave:          flags.autoSave,
	}

	fmt.Println("Lore interactive mode. Enter text and press Enter twice to check.")
	fmt.Println("Commands: 'save' to save pending facts, 'discard' to clear, 'list' to show pending, 'quit' to exit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	var inputBuffer strings.Builder
	emptyLineCount := 0

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// Handle commands
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "quit", "exit":
			if len(state.pendingFacts) > 0 {
				fmt.Printf("Warning: %d pending facts will be lost. Type 'quit' again to confirm.\n", len(state.pendingFacts))
				fmt.Print("> ")
				if scanner.Scan() && strings.ToLower(strings.TrimSpace(scanner.Text())) == "quit" {
					fmt.Println("Goodbye!")
					return nil
				}
				continue
			}
			fmt.Println("Goodbye!")
			return nil

		case "save":
			if err := state.savePendingFacts(ctx); err != nil {
				fmt.Printf("Error saving facts: %v\n", err)
			}
			continue

		case "discard":
			state.pendingFacts = nil
			state.pendingIssues = nil
			fmt.Println("Pending facts discarded.")
			continue

		case "list":
			state.showPendingFacts()
			continue

		case "help":
			fmt.Println("Commands:")
			fmt.Println("  save    - Save all pending facts to database")
			fmt.Println("  discard - Discard all pending facts")
			fmt.Println("  list    - Show pending facts")
			fmt.Println("  quit    - Exit interactive mode")
			fmt.Println("  help    - Show this help")
			fmt.Println()
			fmt.Println("Enter text and press Enter twice to extract and check facts.")
			continue
		}

		// Handle empty lines for double-enter detection
		if line == "" {
			emptyLineCount++
			if emptyLineCount >= 1 && inputBuffer.Len() > 0 {
				// Process the input
				text := strings.TrimSpace(inputBuffer.String())
				if text != "" {
					if err := state.processInput(ctx, text); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}
				inputBuffer.Reset()
				emptyLineCount = 0
			}
			continue
		}

		emptyLineCount = 0
		if inputBuffer.Len() > 0 {
			inputBuffer.WriteString("\n")
		}
		inputBuffer.WriteString(line)
	}

	return scanner.Err()
}

func (s *watchState) processInput(ctx context.Context, text string) error {
	fmt.Println("\nChecking...")

	opts := services.ExtractionOptions{
		CheckConsistency: true,
		CheckOnly:        true, // Don't save yet
	}

	result, err := s.extractionService.ExtractAndStoreWithOptions(ctx, text, s.source, opts)
	if err != nil {
		return fmt.Errorf("extracting facts: %w", err)
	}

	if len(result.Facts) == 0 {
		fmt.Println("No facts found in input.")
		return nil
	}

	fmt.Printf("Found %d facts:\n", len(result.Facts))
	for i, fact := range result.Facts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
	}

	// Display consistency issues
	if len(result.Issues) > 0 {
		fmt.Println()
		for _, issue := range result.Issues {
			severityLabel := formatSeverity(issue.Severity)
			fmt.Printf("%s: %s\n", severityLabel, issue.Description)
			fmt.Printf("  New:      %s %s %s\n", issue.NewFact.Subject, issue.NewFact.Predicate, issue.NewFact.Object)
			fmt.Printf("  Existing: %s %s %s (%s)\n", issue.ExistingFact.Subject, issue.ExistingFact.Predicate, issue.ExistingFact.Object, issue.ExistingFact.SourceFile)
		}
	}

	// Add to pending
	s.pendingFacts = append(s.pendingFacts, result.Facts...)
	s.pendingIssues = append(s.pendingIssues, result.Issues...)

	fmt.Printf("\nFacts queued (%d total pending). Use 'save' to save or 'discard' to clear.\n", len(s.pendingFacts))

	// Auto-save if enabled and no critical issues
	if s.autoSave && !hasCriticalIssues(result.Issues) {
		return s.savePendingFacts(ctx)
	}

	return nil
}

func (s *watchState) savePendingFacts(ctx context.Context) error {
	if len(s.pendingFacts) == 0 {
		fmt.Println("No pending facts to save.")
		return nil
	}

	// Check for critical issues
	if hasCriticalIssues(s.pendingIssues) {
		fmt.Print("Warning: There are CRITICAL consistency issues. Save anyway? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Save cancelled.")
			return nil
		}
	}

	if err := s.vectorDB.SaveBatch(ctx, s.pendingFacts); err != nil {
		return fmt.Errorf("saving facts: %w", err)
	}

	fmt.Printf("Saved %d facts.\n", len(s.pendingFacts))
	s.pendingFacts = nil
	s.pendingIssues = nil
	return nil
}

func (s *watchState) showPendingFacts() {
	if len(s.pendingFacts) == 0 {
		fmt.Println("No pending facts.")
		return
	}

	fmt.Printf("Pending facts (%d):\n", len(s.pendingFacts))
	for i, fact := range s.pendingFacts {
		fmt.Printf("  %d. [%s] %s %s %s\n", i+1, fact.Type, fact.Subject, fact.Predicate, fact.Object)
	}

	if len(s.pendingIssues) > 0 {
		fmt.Printf("\nPending issues (%d):\n", len(s.pendingIssues))
		for _, issue := range s.pendingIssues {
			fmt.Printf("  - %s: %s\n", formatSeverity(issue.Severity), issue.Description)
		}
	}
}

func hasCriticalIssues(issues []ports.ConsistencyIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return true
		}
	}
	return false
}
