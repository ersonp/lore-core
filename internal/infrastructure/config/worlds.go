package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorldsConfig holds dynamic world definitions (read/write).
type WorldsConfig struct {
	Worlds map[string]WorldEntry `yaml:"worlds,omitempty"`
}

// WorldEntry holds configuration for a specific world.
type WorldEntry struct {
	Collection  string `yaml:"collection"`
	Description string `yaml:"description,omitempty"`
}

// LoadWorlds loads world configuration from the .lore directory.
func LoadWorlds(basePath string) (*WorldsConfig, error) {
	worldsFile := filepath.Join(basePath, DefaultConfigDir, DefaultWorldsFile)

	data, err := os.ReadFile(worldsFile)
	if os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &WorldsConfig{
			Worlds: make(map[string]WorldEntry),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading worlds file: %w", err)
	}

	var cfg WorldsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing worlds file: %w", err)
	}

	if cfg.Worlds == nil {
		cfg.Worlds = make(map[string]WorldEntry)
	}

	return &cfg, nil
}

// Save writes the worlds configuration to the worlds file.
func (w *WorldsConfig) Save(basePath string) error {
	configDir := filepath.Join(basePath, DefaultConfigDir)
	worldsFile := filepath.Join(configDir, DefaultWorldsFile)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(w)
	if err != nil {
		return fmt.Errorf("marshaling worlds config: %w", err)
	}

	if err := os.WriteFile(worldsFile, data, 0600); err != nil {
		return fmt.Errorf("writing worlds file: %w", err)
	}

	return nil
}

// Add adds a world to the configuration.
func (w *WorldsConfig) Add(name string, entry WorldEntry) {
	if w.Worlds == nil {
		w.Worlds = make(map[string]WorldEntry)
	}
	w.Worlds[name] = entry
}

// Remove removes a world from the configuration.
func (w *WorldsConfig) Remove(name string) {
	if w.Worlds != nil {
		delete(w.Worlds, name)
	}
}

// Get returns the configuration for a specific world.
func (w *WorldsConfig) Get(name string) (*WorldEntry, error) {
	if len(w.Worlds) == 0 {
		return nil, errors.New("no worlds configured")
	}

	entry, ok := w.Worlds[name]
	if !ok {
		var b strings.Builder
		count := 0
		for k := range w.Worlds {
			if count > 0 {
				b.WriteString(", ")
			}
			b.WriteString(k)
			count++
			if count >= 5 {
				b.WriteString(", ...")
				break
			}
		}
		return nil, fmt.Errorf("world %q not found (available: %s)", name, b.String())
	}

	return &entry, nil
}

// GetCollection returns the Qdrant collection name for a world.
func (w *WorldsConfig) GetCollection(name string) (string, error) {
	entry, err := w.Get(name)
	if err != nil {
		return "", err
	}
	return entry.Collection, nil
}

// Exists checks if a world exists in the configuration.
func (w *WorldsConfig) Exists(name string) bool {
	if w.Worlds == nil {
		return false
	}
	_, ok := w.Worlds[name]
	return ok
}

// WorldsExists checks if a worlds config file exists in the given path.
func WorldsExists(basePath string) bool {
	worldsFile := filepath.Join(basePath, DefaultConfigDir, DefaultWorldsFile)
	_, err := os.Stat(worldsFile)
	return err == nil
}
