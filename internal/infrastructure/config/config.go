// Package config provides configuration loading and management.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigDir is the directory name for lore configuration.
	DefaultConfigDir = ".lore"
	// DefaultConfigFile is the default config file name.
	DefaultConfigFile = "config.yaml"
)

// Config holds all configuration for lore-core.
type Config struct {
	LLM      LLMConfig      `mapstructure:"llm"`
	Embedder EmbedderConfig `mapstructure:"embedder"`
	Qdrant   QdrantConfig   `mapstructure:"qdrant"`
}

// LLMConfig holds configuration for the LLM provider.
type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
}

// EmbedderConfig holds configuration for the embedding provider.
type EmbedderConfig struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
}

// QdrantConfig holds configuration for the Qdrant vector database.
type QdrantConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Collection string `mapstructure:"collection"`
	APIKey     string `mapstructure:"api_key"`
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
			Host:       "localhost",
			Port:       6334,
			Collection: "lore_facts",
		},
	}
}

// Load loads configuration from the .lore directory in the given path.
func Load(basePath string) (*Config, error) {
	configPath := filepath.Join(basePath, DefaultConfigDir)
	configFile := filepath.Join(configPath, DefaultConfigFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s (run 'lore init' first)", configFile)
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v)

	// Bind environment variables
	bindEnvVars(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Override with environment variables if set
	cfg.applyEnvOverrides()

	return &cfg, nil
}

// setDefaults sets default values in viper.
func setDefaults(v *viper.Viper) {
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("embedder.provider", "openai")
	v.SetDefault("embedder.model", "text-embedding-3-small")
	v.SetDefault("qdrant.host", "localhost")
	v.SetDefault("qdrant.port", 6334)
	v.SetDefault("qdrant.collection", "lore_facts")
}

// bindEnvVars binds environment variables to config keys.
func bindEnvVars(v *viper.Viper) {
	v.BindEnv("llm.api_key", "OPENAI_API_KEY")
	v.BindEnv("embedder.api_key", "OPENAI_API_KEY")
	v.BindEnv("qdrant.api_key", "QDRANT_API_KEY")
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
