package config

import (
	"testing"

	"rr/internal/models"
)

func TestAutoShortcut_EmptyList(t *testing.T) {
	got := AutoShortcut("Deploy", nil)
	if got != "d" {
		t.Errorf("expected 'd', got %q", got)
	}
}

func TestAutoShortcut_FirstLetterTaken(t *testing.T) {
	existing := []models.CommandEntry{{ShortcutKey: "b"}}
	got := AutoShortcut("Build", existing)
	if got != "u" {
		t.Errorf("expected 'u' (first free letter), got %q", got)
	}
}

func TestAutoShortcut_AllNameLettersTaken_FallsBackToDigits(t *testing.T) {
	existing := []models.CommandEntry{
		{ShortcutKey: "a"},
		{ShortcutKey: "b"},
	}
	got := AutoShortcut("ab", existing)
	if got != "1" {
		t.Errorf("expected '1' (first digit), got %q", got)
	}
}

func TestAutoShortcut_CaseInsensitive(t *testing.T) {
	existing := []models.CommandEntry{{ShortcutKey: "D"}} // uppercase D blocks lowercase d
	got := AutoShortcut("Deploy", existing)
	if got != "e" {
		t.Errorf("expected 'e' (d taken case-insensitively), got %q", got)
	}
}

func TestAutoShortcut_DuplicateLettersInName(t *testing.T) {
	// "aaa" — only tries 'a' once, then digits
	existing := []models.CommandEntry{{ShortcutKey: "a"}}
	got := AutoShortcut("aaa", existing)
	if got != "1" {
		t.Errorf("expected '1', got %q", got)
	}
}

func TestAutoShortcut_FallsBackToAlphabet(t *testing.T) {
	// Name has no usable letters (digits only in name), digits 1–9 taken → falls back to 'a'
	var existing []models.CommandEntry
	for r := '1'; r <= '9'; r++ {
		existing = append(existing, models.CommandEntry{ShortcutKey: string(r)})
	}
	got := AutoShortcut("123", existing)
	if got != "a" {
		t.Errorf("expected 'a' from alphabet fallback, got %q", got)
	}
}

func TestAutoShortcut_NoneAvailable(t *testing.T) {
	var existing []models.CommandEntry
	for r := 'a'; r <= 'z'; r++ {
		existing = append(existing, models.CommandEntry{ShortcutKey: string(r)})
	}
	for r := '1'; r <= '9'; r++ {
		existing = append(existing, models.CommandEntry{ShortcutKey: string(r)})
	}
	got := AutoShortcut("Test", existing)
	if got != "" {
		t.Errorf("expected '' when all taken, got %q", got)
	}
}

func TestAutoShortcut_DigitInName(t *testing.T) {
	// Digits in name are not tried (only letters from name are tried first)
	// "1deploy" — '1' is not a letter so it's skipped; 'd' should be chosen
	got := AutoShortcut("1deploy", nil)
	if got != "d" {
		t.Errorf("expected 'd' (first letter from name), got %q", got)
	}
}
