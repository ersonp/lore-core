// Package config provides configuration loading and management.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultConfigDir is the directory name for lore configuration.
	DefaultConfigDir = ".lore"
	// DefaultConfigFile is the default config file name.
	DefaultConfigFile = "config.yaml"
	// DefaultWorldsFile is the default worlds file name.
	DefaultWorldsFile = "worlds.yaml"
)

var (
	// reNonAlphanumeric matches characters that aren't alphanumeric or underscore.
	reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9_]`)
	// reMultipleUnderscores matches consecutive underscores.
	reMultipleUnderscores = regexp.MustCompile(`_+`)
)

// Config holds static infrastructure configuration (read-only after init).
type Config struct {
	LLM      LLMConfig      `yaml:"llm,omitempty"`
	Embedder EmbedderConfig `yaml:"embedder,omitempty"`
	Qdrant   QdrantConfig   `yaml:"qdrant,omitempty"`
	SQLite   SQLiteConfig   `yaml:"sqlite,omitempty"`
}

// LLMConfig holds configuration for the LLM provider.
type LLMConfig struct {
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	APIKey   string `yaml:"api_key,omitempty"`
}

// EmbedderConfig holds configuration for the embedding provider.
type EmbedderConfig struct {
	Provider string `yaml:"provider,omitempty"`
	Model    string `yaml:"model,omitempty"`
	APIKey   string `yaml:"api_key,omitempty"`
}

// QdrantConfig holds configuration for the Qdrant vector database.
type QdrantConfig struct {
	Host       string `yaml:"host,omitempty"`
	Port       int    `yaml:"port,omitempty"`
	Collection string `yaml:"collection,omitempty"`
	APIKey     string `yaml:"api_key,omitempty"`
}

// SQLiteConfig holds configuration for the SQLite relational database.
type SQLiteConfig struct {
	// Path is the file path to the SQLite database.
	// For per-world databases, this is computed dynamically using SQLitePathForWorld.
	Path string `yaml:"path,omitempty"`
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: "openai",
			Model:    "gpt-4o-mini",
		},
		Embedder: EmbedderConfig{
			Provider: "openai",
			Model:    "text-embedding-3-small",
		},
		Qdrant: QdrantConfig{
			Host: "localhost",
			Port: 6334,
		},
	}
}

// Load loads configuration from the .lore directory in the given path.
func Load(basePath string) (*Config, error) {
	configFile := filepath.Join(basePath, DefaultConfigDir, DefaultConfigFile)

	data, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s (run 'lore worlds create' first)", configFile)
	}
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Start with defaults
	cfg := Default()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Apply environment variable overrides
	cfg.applyEnvOverrides()

	return cfg, nil
}

// applyEnvOverrides applies environment variable overrides.
func (c *Config) applyEnvOverrides() {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		if c.LLM.APIKey == "" {
			c.LLM.APIKey = key
		}
		if c.Embedder.APIKey == "" {
			c.Embedder.APIKey = key
		}
	}
	if key := os.Getenv("QDRANT_API_KEY"); key != "" {
		if c.Qdrant.APIKey == "" {
			c.Qdrant.APIKey = key
		}
	}
}

// ConfigDir returns the path to the .lore config directory.
func ConfigDir(basePath string) string {
	return filepath.Join(basePath, DefaultConfigDir)
}

// ConfigFilePath returns the path to the config file.
func ConfigFilePath(basePath string) string {
	return filepath.Join(basePath, DefaultConfigDir, DefaultConfigFile)
}

// WorldsFilePath returns the path to the worlds file.
func WorldsFilePath(basePath string) string {
	return filepath.Join(basePath, DefaultConfigDir, DefaultWorldsFile)
}

// Exists checks if a lore config exists in the given path.
func Exists(basePath string) bool {
	configFile := filepath.Join(basePath, DefaultConfigDir, DefaultConfigFile)
	_, err := os.Stat(configFile)
	return err == nil
}

// SanitizeWorldName converts a world name to a valid collection suffix.
func SanitizeWorldName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and hyphens with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove any characters that aren't alphanumeric or underscore
	name = reNonAlphanumeric.ReplaceAllString(name, "")

	// Remove consecutive underscores
	name = reMultipleUnderscores.ReplaceAllString(name, "_")

	// Trim leading/trailing underscores
	name = strings.Trim(name, "_")

	if name == "" {
		return "default"
	}

	return name
}

// GenerateCollectionName creates a collection name for a world.
func GenerateCollectionName(worldName string) string {
	return "lore_" + SanitizeWorldName(worldName)
}

// SQLitePathForWorld returns the SQLite database path for a given world.
func SQLitePathForWorld(basePath, worldName string) string {
	return filepath.Join(basePath, DefaultConfigDir, "worlds", SanitizeWorldName(worldName), "lore.db")
}

// WorldDir returns the directory path for a given world.
func WorldDir(basePath, worldName string) string {
	return filepath.Join(basePath, DefaultConfigDir, "worlds", SanitizeWorldName(worldName))
}
