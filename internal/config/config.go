package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"rr/internal/models"
)

const appName = "rr"
const configFileName = "config.json"
const localConfigFileName = ".rr.json"

// Config holds all persisted application state.
type Config struct {
	Commands []models.CommandEntry `json:"commands"`
}

// ── Global config ─────────────────────────────────────────────────────────────

// globalConfigPath returns ~/.config/rr/config.json (or OS equivalent).
func globalConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(base, appName, configFileName), nil
}

// Load reads the global config. Returns an empty Config if the file does not exist.
func Load() (*Config, error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, err
	}
	return loadFile(path)
}

// Save writes the global config to disk.
func Save(cfg *Config) error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}
	return saveFile(path, cfg)
}

// ── Local config ──────────────────────────────────────────────────────────────

// LocalConfigPath returns the path to .rr.json in the current working directory.
func LocalConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	return filepath.Join(cwd, localConfigFileName), nil
}

// LoadLocal reads the local config from .rr.json in CWD.
// Returns an empty Config if the file does not exist.
func LoadLocal() (*Config, error) {
	path, err := LocalConfigPath()
	if err != nil {
		return nil, err
	}
	return loadFile(path)
}

// SaveLocal writes the local config to .rr.json in CWD.
func SaveLocal(cfg *Config) error {
	path, err := LocalConfigPath()
	if err != nil {
		return err
	}
	return saveFile(path, cfg)
}

// ── Merged view ───────────────────────────────────────────────────────────────

// LoadAll returns local commands (Scope=local) followed by global (Scope=global).
// This is the canonical list used by the TUI and by index-based operations.
func LoadAll() ([]models.CommandEntry, error) {
	local, err := LoadLocal()
	if err != nil {
		return nil, err
	}
	global, err := Load()
	if err != nil {
		return nil, err
	}

	out := make([]models.CommandEntry, 0, len(local.Commands)+len(global.Commands))
	for _, e := range local.Commands {
		e.Scope = models.ScopeLocal
		out = append(out, e)
	}
	for _, e := range global.Commands {
		e.Scope = models.ScopeGlobal
		out = append(out, e)
	}
	return out, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}

func saveFile(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config %s: %w", path, err)
	}
	return nil
}
