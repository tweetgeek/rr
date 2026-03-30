package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"rr/internal/config"
	"rr/internal/models"
)

// setupCLIEnv redirects the global config to a temp dir and changes CWD to a
// fresh temp dir so tests never touch the real ~/.config or the repo checkout.
func setupCLIEnv(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Linux: os.UserConfigDir() uses $XDG_CONFIG_HOME when set
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Chdir(t.TempDir())
}

// runCmd executes the root command with the given args and returns any error.
// Cobra output (usage/errors) is suppressed; fmt.Printf calls in handlers
// still go to os.Stdout (acceptable in test output).
func runCmd(t *testing.T, args ...string) error {
	t.Helper()
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	return root.Execute()
}

func localCommands(t *testing.T) []models.CommandEntry {
	t.Helper()
	cfg, err := config.LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal(): %v", err)
	}
	return cfg.Commands
}

func globalCommands(t *testing.T) []models.CommandEntry {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	return cfg.Commands
}

// ── findIndex ────────────────────────────────────────────────────────────────

func TestFindIndex_ByName(t *testing.T) {
	names := []string{"Alpha", "Beta", "Gamma"}
	nameOf := func(i int) string { return names[i] }

	if got := findIndex("Beta", len(names), nameOf); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestFindIndex_ByIndex_OneBased(t *testing.T) {
	names := []string{"A", "B", "C"}
	nameOf := func(i int) string { return names[i] }

	if got := findIndex("1", len(names), nameOf); got != 0 {
		t.Errorf("expected 0 for index '1', got %d", got)
	}
	if got := findIndex("3", len(names), nameOf); got != 2 {
		t.Errorf("expected 2 for index '3', got %d", got)
	}
}

func TestFindIndex_OutOfRange(t *testing.T) {
	names := []string{"A"}
	nameOf := func(i int) string { return names[i] }

	if got := findIndex("0", len(names), nameOf); got != -1 {
		t.Errorf("expected -1 for index 0 (out of range), got %d", got)
	}
	if got := findIndex("2", len(names), nameOf); got != -1 {
		t.Errorf("expected -1 for index 2 (out of range), got %d", got)
	}
}

func TestFindIndex_NotFound(t *testing.T) {
	names := []string{"A", "B"}
	nameOf := func(i int) string { return names[i] }

	if got := findIndex("Z", len(names), nameOf); got != -1 {
		t.Errorf("expected -1 for unknown name, got %d", got)
	}
}

func TestFindIndex_EmptyList(t *testing.T) {
	if got := findIndex("anything", 0, func(int) string { return "" }); got != -1 {
		t.Errorf("expected -1 for empty list, got %d", got)
	}
}

// ── add ──────────────────────────────────────────────────────────────────────

func TestAdd_LocalCommand(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "add", "-n", "Deploy", "-c", "make deploy"); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	cmds := localCommands(t)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 local command, got %d", len(cmds))
	}
	if cmds[0].Name != "Deploy" {
		t.Errorf("Name: expected 'Deploy', got %q", cmds[0].Name)
	}
	if cmds[0].Command != "make deploy" {
		t.Errorf("Command: expected 'make deploy', got %q", cmds[0].Command)
	}
	if cmds[0].ShortcutKey != "d" {
		t.Errorf("ShortcutKey: expected auto-shortcut 'd', got %q", cmds[0].ShortcutKey)
	}
}

func TestAdd_GlobalCommand(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "add", "--global", "-n", "Status", "-c", "git status"); err != nil {
		t.Fatalf("add --global failed: %v", err)
	}
	cmds := globalCommands(t)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 global command, got %d", len(cmds))
	}
	if cmds[0].Name != "Status" {
		t.Errorf("Name: expected 'Status', got %q", cmds[0].Name)
	}
}

func TestAdd_GlobalDoesNotPollutesLocal(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "--global", "-n", "X", "-c", "x")
	if len(localCommands(t)) != 0 {
		t.Error("global add should not affect local config")
	}
}

func TestAdd_ManualShortcut(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "add", "-n", "Deploy", "-c", "make deploy", "-s", "x"); err != nil {
		t.Fatalf("add with shortcut failed: %v", err)
	}
	cmds := localCommands(t)
	if cmds[0].ShortcutKey != "x" {
		t.Errorf("expected shortcut 'x', got %q", cmds[0].ShortcutKey)
	}
}

func TestAdd_DuplicateShortcutFails(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "A", "-c", "echo a", "-s", "x")
	if err := runCmd(t, "add", "-n", "B", "-c", "echo b", "-s", "x"); err == nil {
		t.Error("expected error for duplicate shortcut, got nil")
	}
}

func TestAdd_MissingNameFails(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "add", "-c", "echo hi"); err == nil {
		t.Error("expected error for missing --name, got nil")
	}
}

func TestAdd_MissingCommandFails(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "add", "-n", "Test"); err == nil {
		t.Error("expected error for missing --command, got nil")
	}
}

func TestAdd_MultipleEntries_AccumulateInFile(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "A", "-c", "echo a")
	runCmd(t, "add", "-n", "B", "-c", "echo b")
	cmds := localCommands(t)
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}
}

func TestAdd_ShortcutCrossScope_ConflictDetected(t *testing.T) {
	setupCLIEnv(t)
	// Add local entry with shortcut 'g'
	runCmd(t, "add", "-n", "Local", "-c", "echo local", "-s", "g")
	// Try to add global with same shortcut 'g'
	if err := runCmd(t, "add", "--global", "-n", "Global", "-c", "echo global", "-s", "g"); err == nil {
		t.Error("expected cross-scope shortcut conflict error, got nil")
	}
}

// ── remove ────────────────────────────────────────────────────────────────────

func TestRemove_ByName(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")
	runCmd(t, "add", "-n", "Build", "-c", "go build")

	if err := runCmd(t, "remove", "Deploy"); err != nil {
		t.Fatalf("remove by name failed: %v", err)
	}
	cmds := localCommands(t)
	if len(cmds) != 1 || cmds[0].Name != "Build" {
		t.Errorf("expected only 'Build' to remain, got %+v", cmds)
	}
}

func TestRemove_ByIndex(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "A", "-c", "echo a")
	runCmd(t, "add", "-n", "B", "-c", "echo b")

	if err := runCmd(t, "remove", "1"); err != nil {
		t.Fatalf("remove by index failed: %v", err)
	}
	cmds := localCommands(t)
	if len(cmds) != 1 || cmds[0].Name != "B" {
		t.Errorf("expected 'B' to remain after removing index 1, got %+v", cmds)
	}
}

func TestRemove_NotFoundFails(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "remove", "NonExistent"); err == nil {
		t.Error("expected error for non-existent command, got nil")
	}
}

func TestRemove_GlobalByName(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "--global", "-n", "GlobalCmd", "-c", "echo global")

	if err := runCmd(t, "remove", "--global", "GlobalCmd"); err != nil {
		t.Fatalf("remove --global by name failed: %v", err)
	}
	if len(globalCommands(t)) != 0 {
		t.Error("expected 0 global commands after remove")
	}
}

func TestRemove_FallsBackToGlobal(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "--global", "-n", "GlobalCmd", "-c", "echo global")
	// No local config; remove should fall back to global
	if err := runCmd(t, "remove", "GlobalCmd"); err != nil {
		t.Fatalf("remove with global fallback failed: %v", err)
	}
	if len(globalCommands(t)) != 0 {
		t.Error("expected 0 global commands after fallback removal")
	}
}

func TestRemove_LocalPreferredOverGlobal(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Cmd", "-c", "echo local")
	runCmd(t, "add", "--global", "-n", "Cmd", "-c", "echo global")

	// Without --global, should remove the local one first
	runCmd(t, "remove", "Cmd")
	if len(localCommands(t)) != 0 {
		t.Error("expected local 'Cmd' to be removed")
	}
	if len(globalCommands(t)) != 1 {
		t.Error("expected global 'Cmd' to remain")
	}
}

func TestRemove_Aliases(t *testing.T) {
	for _, alias := range []string{"rm", "delete", "del"} {
		t.Run(alias, func(t *testing.T) {
			setupCLIEnv(t)
			runCmd(t, "add", "-n", "X", "-c", "x")
			if err := runCmd(t, alias, "X"); err != nil {
				t.Errorf("alias %q failed: %v", alias, err)
			}
			if len(localCommands(t)) != 0 {
				t.Errorf("alias %q: expected 0 commands after remove", alias)
			}
		})
	}
}

// ── edit ──────────────────────────────────────────────────────────────────────

func TestEdit_Name(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")

	if err := runCmd(t, "edit", "Deploy", "--name", "Release"); err != nil {
		t.Fatalf("edit --name failed: %v", err)
	}
	cmds := localCommands(t)
	if len(cmds) != 1 || cmds[0].Name != "Release" {
		t.Errorf("expected name 'Release', got %+v", cmds)
	}
}

func TestEdit_Command(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")

	if err := runCmd(t, "edit", "Deploy", "--command", "make deploy-prod"); err != nil {
		t.Fatalf("edit --command failed: %v", err)
	}
	cmds := localCommands(t)
	if cmds[0].Command != "make deploy-prod" {
		t.Errorf("expected 'make deploy-prod', got %q", cmds[0].Command)
	}
}

func TestEdit_Shortcut(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")

	if err := runCmd(t, "edit", "Deploy", "--shortcut", "z"); err != nil {
		t.Fatalf("edit --shortcut failed: %v", err)
	}
	cmds := localCommands(t)
	if cmds[0].ShortcutKey != "z" {
		t.Errorf("expected shortcut 'z', got %q", cmds[0].ShortcutKey)
	}
}

func TestEdit_PreservesUnchangedFields(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy", "-s", "d")

	runCmd(t, "edit", "Deploy", "--name", "Release")
	cmds := localCommands(t)
	// Command and shortcut should be unchanged
	if cmds[0].Command != "make deploy" {
		t.Errorf("command should be preserved, got %q", cmds[0].Command)
	}
	if cmds[0].ShortcutKey != "d" {
		t.Errorf("shortcut should be preserved, got %q", cmds[0].ShortcutKey)
	}
}

func TestEdit_ShortcutConflictFails(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "A", "-c", "echo a", "-s", "a")
	runCmd(t, "add", "-n", "B", "-c", "echo b", "-s", "b")

	if err := runCmd(t, "edit", "B", "--shortcut", "a"); err == nil {
		t.Error("expected error for shortcut conflict, got nil")
	}
}

func TestEdit_NoFlagsFails(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")

	if err := runCmd(t, "edit", "Deploy"); err == nil {
		t.Error("expected error when no flags provided, got nil")
	}
}

func TestEdit_ByIndex(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "-n", "Deploy", "-c", "make deploy")

	if err := runCmd(t, "edit", "1", "--command", "make deploy-staging"); err != nil {
		t.Fatalf("edit by index failed: %v", err)
	}
	cmds := localCommands(t)
	if cmds[0].Command != "make deploy-staging" {
		t.Errorf("expected 'make deploy-staging', got %q", cmds[0].Command)
	}
}

func TestEdit_NotFoundFails(t *testing.T) {
	setupCLIEnv(t)
	if err := runCmd(t, "edit", "NonExistent", "--name", "X"); err == nil {
		t.Error("expected error for non-existent command, got nil")
	}
}

func TestEdit_GlobalCommand(t *testing.T) {
	setupCLIEnv(t)
	runCmd(t, "add", "--global", "-n", "GCmd", "-c", "echo g")

	if err := runCmd(t, "edit", "--global", "GCmd", "--command", "echo global-updated"); err != nil {
		t.Fatalf("edit --global failed: %v", err)
	}
	cmds := globalCommands(t)
	if cmds[0].Command != "echo global-updated" {
		t.Errorf("expected 'echo global-updated', got %q", cmds[0].Command)
	}
}
