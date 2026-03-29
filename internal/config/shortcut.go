package config

import (
	"strings"
	"unicode"

	"rr/internal/models"
)

// takenShortcuts returns a set of all shortcut keys already in use.
func takenShortcuts(cmds []models.CommandEntry) map[string]bool {
	taken := make(map[string]bool, len(cmds))
	for _, e := range cmds {
		if e.ShortcutKey != "" {
			taken[strings.ToLower(e.ShortcutKey)] = true
		}
	}
	return taken
}

// AutoShortcut picks the first available single-character shortcut for name.
// Priority:
//  1. Characters from name (letters only, lowercased, in order)
//  2. Digits '1'–'9'
//  3. Letters 'a'–'z' not already tried via the name
//
// Returns "" if every candidate is taken.
func AutoShortcut(name string, existing []models.CommandEntry) string {
	taken := takenShortcuts(existing)

	tried := make(map[string]bool)
	try := func(ch string) (string, bool) {
		k := strings.ToLower(ch)
		if tried[k] || taken[k] {
			return "", false
		}
		tried[k] = true
		return k, true
	}

	// 1. Letters from name
	for _, r := range name {
		if unicode.IsLetter(r) {
			if k, ok := try(string(r)); ok {
				return k
			}
		}
	}

	// 2. Digits
	for r := '1'; r <= '9'; r++ {
		if k, ok := try(string(r)); ok {
			return k
		}
	}

	// 3. Remaining alphabet
	for r := 'a'; r <= 'z'; r++ {
		if k, ok := try(string(r)); ok {
			return k
		}
	}

	return ""
}
