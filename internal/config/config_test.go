package config

import (
	"path/filepath"
	"testing"

	"rr/internal/models"
)

// setupHome redirects os.UserConfigDir() to a temp dir so tests don't touch
// the real ~/.config or ~/Library/Application Support.
func setupHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Linux: os.UserConfigDir() uses $XDG_CONFIG_HOME when set
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
}

// ── Load / Save (global) ─────────────────────────────────────────────────────

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	setupHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(cfg.Commands))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	setupHome(t)

	original := &Config{
		Commands: []models.CommandEntry{
			{Name: "Deploy", Command: "make deploy", ShortcutKey: "d"},
		},
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(got.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(got.Commands))
	}
	if got.Commands[0].Name != "Deploy" {
		t.Errorf("Name: expected 'Deploy', got %q", got.Commands[0].Name)
	}
	if got.Commands[0].Command != "make deploy" {
		t.Errorf("Command: expected 'make deploy', got %q", got.Commands[0].Command)
	}
	if got.Commands[0].ShortcutKey != "d" {
		t.Errorf("ShortcutKey: expected 'd', got %q", got.Commands[0].ShortcutKey)
	}
}

func TestSave_ScopeNotPersisted(t *testing.T) {
	setupHome(t)

	original := &Config{
		Commands: []models.CommandEntry{
			{Name: "X", Command: "x", Scope: models.ScopeGlobal},
		},
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Commands[0].Scope != "" {
		t.Errorf("Scope should not be persisted (json:\"-\"), got %q", got.Commands[0].Scope)
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	setupHome(t)

	cfg := &Config{Commands: []models.CommandEntry{{Name: "A", Command: "a"}}}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() should create parent dirs: %v", err)
	}
}

// ── LoadLocal / SaveLocal ────────────────────────────────────────────────────

func TestLoadLocal_MissingFileReturnsEmpty(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error: %v", err)
	}
	if len(cfg.Commands) != 0 {
		t.Errorf("expected 0 commands, got %d", len(cfg.Commands))
	}
}

func TestSaveLocalAndLoadLocal_RoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())

	original := &Config{
		Commands: []models.CommandEntry{
			{Name: "Test", Command: "go test ./...", ShortcutKey: "t"},
		},
	}
	if err := SaveLocal(original); err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}

	got, err := LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error: %v", err)
	}
	if len(got.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(got.Commands))
	}
	if got.Commands[0].Name != "Test" {
		t.Errorf("Name: expected 'Test', got %q", got.Commands[0].Name)
	}
}

func TestSaveLocal_WritesToCWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := &Config{Commands: []models.CommandEntry{{Name: "A", Command: "a"}}}
	if err := SaveLocal(cfg); err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}

	path, _ := LocalConfigPath()
	if path != filepath.Join(dir, ".rr.json") {
		t.Errorf("expected path in CWD, got %q", path)
	}
}

// ── LoadAll ──────────────────────────────────────────────────────────────────

func TestLoadAll_BothScopes(t *testing.T) {
	setupHome(t)
	t.Chdir(t.TempDir())

	if err := SaveLocal(&Config{Commands: []models.CommandEntry{
		{Name: "LocalCmd", Command: "echo local", ShortcutKey: "l"},
	}}); err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}
	if err := Save(&Config{Commands: []models.CommandEntry{
		{Name: "GlobalCmd", Command: "echo global", ShortcutKey: "g"},
	}}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Local comes first
	if all[0].Name != "LocalCmd" || all[0].Scope != models.ScopeLocal {
		t.Errorf("entry 0: expected LocalCmd/local, got %+v", all[0])
	}
	if all[1].Name != "GlobalCmd" || all[1].Scope != models.ScopeGlobal {
		t.Errorf("entry 1: expected GlobalCmd/global, got %+v", all[1])
	}
}

func TestLoadAll_OnlyLocal(t *testing.T) {
	setupHome(t)
	t.Chdir(t.TempDir())

	if err := SaveLocal(&Config{Commands: []models.CommandEntry{
		{Name: "LocalCmd", Command: "echo local"},
	}}); err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(all) != 1 || all[0].Scope != models.ScopeLocal {
		t.Errorf("expected 1 local entry, got %+v", all)
	}
}

func TestLoadAll_OnlyGlobal(t *testing.T) {
	setupHome(t)
	t.Chdir(t.TempDir())

	if err := Save(&Config{Commands: []models.CommandEntry{
		{Name: "GlobalCmd", Command: "echo global"},
	}}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(all) != 1 || all[0].Scope != models.ScopeGlobal {
		t.Errorf("expected 1 global entry, got %+v", all)
	}
}

func TestLoadAll_EmptyConfigs(t *testing.T) {
	setupHome(t)
	t.Chdir(t.TempDir())

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 entries, got %d", len(all))
	}
}

func TestLoadAll_MultipleLocalAndGlobal(t *testing.T) {
	setupHome(t)
	t.Chdir(t.TempDir())

	if err := SaveLocal(&Config{Commands: []models.CommandEntry{
		{Name: "L1", Command: "l1"},
		{Name: "L2", Command: "l2"},
	}}); err != nil {
		t.Fatalf("SaveLocal() error: %v", err)
	}
	if err := Save(&Config{Commands: []models.CommandEntry{
		{Name: "G1", Command: "g1"},
	}}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	for _, e := range all[:2] {
		if e.Scope != models.ScopeLocal {
			t.Errorf("expected local scope for %q, got %q", e.Name, e.Scope)
		}
	}
	if all[2].Scope != models.ScopeGlobal {
		t.Errorf("expected global scope for %q, got %q", all[2].Name, all[2].Scope)
	}
}
