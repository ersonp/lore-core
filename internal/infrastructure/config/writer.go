package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WriteDefault creates the .lore directory and writes default config files.
func WriteDefault(basePath string) error {
	return WriteDefaultWithWorld(basePath, "default", "")
}

// WriteDefaultWithWorld creates the .lore directory and writes config files with the specified world.
func WriteDefaultWithWorld(basePath string, worldName string, description string) error {
	configDir := filepath.Join(basePath, DefaultConfigDir)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Write config.yaml (static infrastructure config)
	if err := writeDefaultConfig(basePath); err != nil {
		return err
	}

	// Write worlds.yaml with the initial world
	worlds := &WorldsConfig{
		Worlds: map[string]WorldEntry{
			worldName: {
				Collection:  GenerateCollectionName(worldName),
				Description: description,
			},
		},
	}

	if err := worlds.Save(basePath); err != nil {
		return fmt.Errorf("writing worlds file: %w", err)
	}

	return nil
}

// writeDefaultConfig writes the default config.yaml file.
func writeDefaultConfig(basePath string) error {
	configFile := filepath.Join(basePath, DefaultConfigDir, DefaultConfigFile)

	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	}

	cfg := Default()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// WriteConfig writes the given config to the config file.
func WriteConfig(basePath string, cfg *Config) error {
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
