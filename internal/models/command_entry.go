package models

// Scope indicates whether a command lives in the local or global config.
type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeGlobal Scope = "global"
)

// CommandEntry represents a single saved command.
// Scope is runtime-only and is never written to disk (json:"-").
type CommandEntry struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	ShortcutKey string `json:"shortcut_key"`
	Scope       Scope  `json:"-"`
}
