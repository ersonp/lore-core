package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

	cfg := Default()
	cfg.Worlds[worldName] = WorldConfig{
		Collection:  GenerateCollectionName(worldName),
		Description: description,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
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

	if err := os.WriteFile(configFile, data, 0600); err != nil {
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
