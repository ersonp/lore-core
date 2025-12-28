// Package openai provides an LLMClient implementation using OpenAI.
package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/ersonp/lore-core/internal/domain/entities"
	"github.com/ersonp/lore-core/internal/domain/ports"
	"github.com/ersonp/lore-core/internal/infrastructure/config"
)

const extractionPrompt = `You are a fact extractor for fictional worlds. Extract facts from the given text.

For each fact, identify:
- type: character, location, event, relationship, rule, timeline
- subject: What/who the fact is about
- predicate: The property or relationship
- object: The value or target
- context: Any relevant context (optional)
- confidence: How confident you are (0.0-1.0)

Return ONLY a valid JSON array, no other text.

Example:
Input: "Frodo has blue eyes and lives in the Shire."
Output: [
  {"type": "character", "subject": "Frodo", "predicate": "eye_color", "object": "blue", "confidence": 0.95},
  {"type": "character", "subject": "Frodo", "predicate": "lives_in", "object": "the Shire", "confidence": 0.95}
]`

const consistencyPrompt = `Compare these new facts against existing facts. Identify any inconsistencies or contradictions.

New facts:
%s

Existing facts:
%s

For each inconsistency found, return:
- new_fact_index: Index of the conflicting new fact (0-based)
- existing_fact_index: Index of the contradicted existing fact (0-based)
- description: What the conflict is
- severity: "minor", "major", or "critical"

Return ONLY a valid JSON array, no other text. Return empty array [] if no inconsistencies found.`

// Client implements the LLMClient interface using OpenAI.
type Client struct {
	client *openai.Client
	model  string
}

// NewClient creates a new OpenAI LLM client.
func NewClient(cfg config.LLMConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	client := openai.NewClient(cfg.APIKey)

	model := "gpt-4o-mini"
	if cfg.Model != "" {
		model = cfg.Model
	}

	return &Client{
		client: client,
		model:  model,
	}, nil
}

// ExtractFacts extracts facts from the given text.
func (c *Client) ExtractFacts(ctx context.Context, text string) ([]entities.Fact, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: extractionPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("calling OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var rawFacts []rawFact
	if err := json.Unmarshal([]byte(content), &rawFacts); err != nil {
		return nil, fmt.Errorf("parsing facts JSON: %w (response: %s)", err, content)
	}

	facts := make([]entities.Fact, 0, len(rawFacts))
	for _, rf := range rawFacts {
		fact := entities.Fact{
			Type:       entities.FactType(rf.Type),
			Subject:    rf.Subject,
			Predicate:  rf.Predicate,
			Object:     objectToString(rf.Object),
			Context:    rf.Context,
			Confidence: rf.Confidence,
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

// CheckConsistency checks if new facts are consistent with existing facts.
func (c *Client) CheckConsistency(ctx context.Context, newFacts []entities.Fact, existingFacts []entities.Fact) ([]ports.ConsistencyIssue, error) {
	if len(newFacts) == 0 || len(existingFacts) == 0 {
		return nil, nil
	}

	newFactsJSON, err := json.Marshal(factsToRaw(newFacts))
	if err != nil {
		return nil, fmt.Errorf("marshaling new facts: %w", err)
	}

	existingFactsJSON, err := json.Marshal(factsToRaw(existingFacts))
	if err != nil {
		return nil, fmt.Errorf("marshaling existing facts: %w", err)
	}

	prompt := fmt.Sprintf(consistencyPrompt, string(newFactsJSON), string(existingFactsJSON))

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("calling OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	content = cleanJSONResponse(content)

	var rawIssues []rawConsistencyIssue
	if err := json.Unmarshal([]byte(content), &rawIssues); err != nil {
		return nil, fmt.Errorf("parsing consistency JSON: %w (response: %s)", err, content)
	}

	issues := make([]ports.ConsistencyIssue, 0, len(rawIssues))
	for _, ri := range rawIssues {
		if ri.NewFactIndex >= len(newFacts) || ri.ExistingFactIndex >= len(existingFacts) {
			continue
		}

		issue := ports.ConsistencyIssue{
			NewFact:      newFacts[ri.NewFactIndex],
			ExistingFact: existingFacts[ri.ExistingFactIndex],
			Description:  ri.Description,
			Severity:     ri.Severity,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// rawFact is the JSON structure for extracted facts.
type rawFact struct {
	Type       string      `json:"type"`
	Subject    string      `json:"subject"`
	Predicate  string      `json:"predicate"`
	Object     interface{} `json:"object"`
	Context    string      `json:"context,omitempty"`
	Confidence float64     `json:"confidence"`
}

// objectToString converts the object field to string (handles numbers from LLM).
func objectToString(obj interface{}) string {
	switch v := obj.(type) {
	case string:
		return v
	case float64:
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'g', -1, 64)
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// rawConsistencyIssue is the JSON structure for consistency issues.
type rawConsistencyIssue struct {
	NewFactIndex      int    `json:"new_fact_index"`
	ExistingFactIndex int    `json:"existing_fact_index"`
	Description       string `json:"description"`
	Severity          string `json:"severity"`
}

// factsToRaw converts entities to raw format for JSON.
func factsToRaw(facts []entities.Fact) []rawFact {
	raw := make([]rawFact, 0, len(facts))
	for i := range facts {
		raw = append(raw, rawFact{
			Type:       string(facts[i].Type),
			Subject:    facts[i].Subject,
			Predicate:  facts[i].Predicate,
			Object:     facts[i].Object,
			Context:    facts[i].Context,
			Confidence: facts[i].Confidence,
		})
	}
	return raw
}

// cleanJSONResponse removes markdown code blocks if present.
func cleanJSONResponse(content string) string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	return strings.TrimSpace(content)
}
