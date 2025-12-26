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
	sourceFile string
	autoSave   bool
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

	cmd.Flags().StringVarP(&flags.sourceFile, "source", "s", "interactive", "Source name for facts")
	cmd.Flags().BoolVar(&flags.autoSave, "save", false, "Auto-save facts (with confirmation on conflicts)")

	return cmd
}

type watchState struct {
	pendingFacts      []entities.Fact
	pendingIssues     []ports.ConsistencyIssue
	extractionService *services.ExtractionService
	vectorDB          ports.VectorDB
	sourceFile        string
	autoSave          bool
}

func runWatch(cmd *cobra.Command, flags watchFlags) error {
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

	if globalWorld == "" {
		return fmt.Errorf("world is required (use --world flag)")
	}

	collection, err := worlds.GetCollection(globalWorld)
	if err != nil {
		return err
	}

	qdrantCfg := cfg.Qdrant
	qdrantCfg.Collection = collection

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

	state := &watchState{
		extractionService: services.NewExtractionService(llmClient, emb, repo),
		vectorDB:          repo,
		sourceFile:        flags.sourceFile,
		autoSave:          flags.autoSave,
	}

	return state.runInputLoop(cmd.Context())
}

func (s *watchState) runInputLoop(ctx context.Context) error {
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
		input := strings.ToLower(strings.TrimSpace(line))

		if handled, shouldExit := s.handleCommand(ctx, input, scanner); handled {
			if shouldExit {
				return nil
			}
			continue
		}

		emptyLineCount = s.handleInput(ctx, line, &inputBuffer, emptyLineCount)
	}

	return scanner.Err()
}

// handleCommand processes user commands. Returns (handled, shouldExit).
func (s *watchState) handleCommand(ctx context.Context, input string, scanner *bufio.Scanner) (bool, bool) {
	switch input {
	case "quit", "exit":
		return true, s.handleQuit(scanner)
	case "save":
		if err := s.savePendingFacts(ctx); err != nil {
			fmt.Printf("Error saving facts: %v\n", err)
		}
		return true, false
	case "discard":
		s.pendingFacts = nil
		s.pendingIssues = nil
		fmt.Println("Pending facts discarded.")
		return true, false
	case "list":
		s.showPendingFacts()
		return true, false
	case "help":
		s.showHelp()
		return true, false
	default:
		return false, false
	}
}

func (s *watchState) handleQuit(scanner *bufio.Scanner) bool {
	if len(s.pendingFacts) > 0 {
		fmt.Printf("Warning: %d pending facts will be lost. Type 'quit' again to confirm.\n", len(s.pendingFacts))
		fmt.Print("> ")
		if scanner.Scan() && strings.ToLower(strings.TrimSpace(scanner.Text())) == "quit" {
			fmt.Println("Goodbye!")
			return true
		}
		return false
	}
	fmt.Println("Goodbye!")
	return true
}

func (s *watchState) showHelp() {
	fmt.Println("Commands:")
	fmt.Println("  save    - Save all pending facts to database")
	fmt.Println("  discard - Discard all pending facts")
	fmt.Println("  list    - Show pending facts")
	fmt.Println("  quit    - Exit interactive mode")
	fmt.Println("  help    - Show this help")
	fmt.Println()
	fmt.Println("Enter text and press Enter twice to extract and check facts.")
}

func (s *watchState) handleInput(ctx context.Context, line string, inputBuffer *strings.Builder, emptyLineCount int) int {
	if line == "" {
		emptyLineCount++
		if emptyLineCount >= 1 && inputBuffer.Len() > 0 {
			if err := processBufferedInput(ctx, s, inputBuffer); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			return 0
		}
		return emptyLineCount
	}

	if inputBuffer.Len() > 0 {
		inputBuffer.WriteString("\n")
	}
	inputBuffer.WriteString(line)
	return 0
}

func processBufferedInput(ctx context.Context, state *watchState, inputBuffer *strings.Builder) error {
	text := strings.TrimSpace(inputBuffer.String())
	inputBuffer.Reset()
	if text == "" {
		return nil
	}
	return state.processInput(ctx, text)
}

func (s *watchState) processInput(ctx context.Context, text string) error {
	fmt.Println("\nChecking...")

	opts := services.ExtractionOptions{
		CheckConsistency: true,
		CheckOnly:        true, // Don't save yet
	}

	result, err := s.extractionService.ExtractAndStoreWithOptions(ctx, text, s.sourceFile, opts)
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
		response, _ := reader.ReadString('\n') // Error ignored: EOF/error treated as "no"
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
