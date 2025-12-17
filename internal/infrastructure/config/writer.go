package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfigYAML is the default configuration content.
const DefaultConfigYAML = `# Lore-Core Configuration

llm:
  provider: openai
  model: gpt-4o-mini
  # api_key: your-api-key (or set OPENAI_API_KEY env var)

embedder:
  provider: openai
  model: text-embedding-3-small
  # api_key: your-api-key (or set OPENAI_API_KEY env var)

qdrant:
  host: localhost
  port: 6334
  collection: lore_facts
  # api_key: your-api-key (for Qdrant Cloud)
`

// WriteDefault creates the .lore directory and writes a default config file.
func WriteDefault(basePath string) error {
	return WriteDefaultWithWorld(basePath, "default", "")
}

// WriteDefaultWithWorld creates the .lore directory and writes a config file with the specified world.
func WriteDefaultWithWorld(basePath string, worldName string, description string) error {
	configDir := filepath.Join(basePath, DefaultConfigDir)
	configFile := filepath.Join(configDir, DefaultConfigFile)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	}

	collection := GenerateCollectionName(worldName)
	configContent := fmt.Sprintf(`# Lore-Core Configuration

llm:
  provider: openai
  model: gpt-4o-mini
  # api_key: your-api-key (or set OPENAI_API_KEY env var)

embedder:
  provider: openai
  model: text-embedding-3-small
  # api_key: your-api-key (or set OPENAI_API_KEY env var)

qdrant:
  host: localhost
  port: 6334
  # api_key: your-api-key (for Qdrant Cloud)

worlds:
  %s:
    collection: %s
    description: "%s"
`, worldName, collection, description)

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Write writes the given config to the config file.
func Write(basePath string, cfg *Config) error {
	configDir := filepath.Join(basePath, DefaultConfigDir)
	configFile := filepath.Join(configDir, DefaultConfigFile)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Exists checks if a lore config exists in the given path.
func Exists(basePath string) bool {
	configFile := filepath.Join(basePath, DefaultConfigDir, DefaultConfigFile)
	_, err := os.Stat(configFile)
	return err == nil
}
