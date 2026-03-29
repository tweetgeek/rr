package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rr/internal/config"
	"rr/internal/models"
)

// ── Form mode ─────────────────────────────────────────────────────────────────

type formMode int

const (
	modeAdd  formMode = iota
	modeEdit
)

// ── Form styles ───────────────────────────────────────────────────────────────

var (
	formTitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(localAccent).PaddingLeft(2)
	labelStyle         = lipgloss.NewStyle().Width(11).Foreground(lipgloss.Color("250"))
	scopeActiveStyle   = lipgloss.NewStyle().Foreground(localAccent).Bold(true)
	scopeInactiveStyle = dimStyle
)

// fieldIdx constants for readability.
const (
	fieldName     = 0
	fieldCommand  = 1
	fieldShortcut = 2
	fieldScope    = 3
	fieldCount    = 4
)

// ── formData ──────────────────────────────────────────────────────────────────

type formData struct {
	mode     formMode
	original models.CommandEntry // for edit: the entry being replaced
	inputs   [3]textinput.Model
	focus    int          // 0–3
	scope    models.Scope // local or global
	errMsg   string
}

func newInput(placeholder string, limit int) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.CharLimit = limit
	t.Prompt = ""
	t.Width = 40
	t.PromptStyle = dimStyle
	t.PlaceholderStyle = dimStyle
	t.TextStyle = lipgloss.NewStyle()
	return t
}

func initAddForm() formData {
	f := formData{
		mode:  modeAdd,
		scope: models.ScopeLocal,
	}
	f.inputs[fieldName] = newInput("required", 80)
	f.inputs[fieldCommand] = newInput("required", 256)
	f.inputs[fieldShortcut] = newInput("auto", 1)
	f.inputs[fieldName].Focus()
	return f
}

func initEditForm(e models.CommandEntry) formData {
	f := initAddForm()
	f.mode = modeEdit
	f.original = e
	f.scope = e.Scope
	f.inputs[fieldName].SetValue(e.Name)
	f.inputs[fieldCommand].SetValue(e.Command)
	f.inputs[fieldShortcut].SetValue(e.ShortcutKey)
	return f
}

// ── Form update ───────────────────────────────────────────────────────────────

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, isKey := msg.(tea.KeyMsg)
	if !isKey {
		// Forward to focused textinput
		var cmd tea.Cmd
		if m.form.focus < 3 {
			m.form.inputs[m.form.focus], cmd = m.form.inputs[m.form.focus].Update(msg)
		}
		return m, cmd
	}

	switch key.String() {
	case "esc", "ctrl+c":
		m.state = stateList
		return m, nil

	case "tab", "down":
		m.form.errMsg = ""
		m.form.blur()
		m.form.focus = (m.form.focus + 1) % fieldCount
		m.form.focusCurrent()

	case "shift+tab", "up":
		m.form.errMsg = ""
		m.form.blur()
		m.form.focus = (m.form.focus - 1 + fieldCount) % fieldCount
		m.form.focusCurrent()

	case " ":
		// Toggle scope when on scope field.
		if m.form.focus == fieldScope {
			if m.form.scope == models.ScopeLocal {
				m.form.scope = models.ScopeGlobal
			} else {
				m.form.scope = models.ScopeLocal
			}
			return m, nil
		}

	case "enter":
		if m.form.focus == fieldScope || m.form.focus == fieldShortcut {
			// Save from last two fields.
			return m, m.form.save(m.entries)
		}
		// Advance to next field from name/command.
		m.form.errMsg = ""
		m.form.blur()
		m.form.focus = (m.form.focus + 1) % fieldCount
		m.form.focusCurrent()
	}

	// Forward to focused textinput.
	var cmd tea.Cmd
	if m.form.focus < 3 {
		m.form.inputs[m.form.focus], cmd = m.form.inputs[m.form.focus].Update(msg)
	}
	return m, cmd
}

func (f *formData) blur() {
	if f.focus < 3 {
		f.inputs[f.focus].Blur()
	}
}

func (f *formData) focusCurrent() {
	if f.focus < 3 {
		f.inputs[f.focus].Focus()
	}
}

// save validates the form and writes to config, then emits a reloadMsg.
func (f formData) save(existing []models.CommandEntry) tea.Cmd {
	name := strings.TrimSpace(f.inputs[fieldName].Value())
	cmd := strings.TrimSpace(f.inputs[fieldCommand].Value())
	shortcut := strings.TrimSpace(f.inputs[fieldShortcut].Value())

	// Inline validation — return an errMsg cmd if invalid.
	errCmd := func(msg string) tea.Cmd {
		return func() tea.Msg { return formErrMsg(msg) }
	}

	if name == "" {
		return errCmd("name is required")
	}
	if cmd == "" {
		return errCmd("command is required")
	}

	// Shortcut uniqueness check (exclude the entry being edited).
	if shortcut != "" {
		for _, e := range existing {
			if e.ShortcutKey != shortcut {
				continue
			}
			if f.mode == modeEdit && e.Name == f.original.Name && e.Scope == f.original.Scope {
				continue // same entry — OK
			}
			return errCmd(fmt.Sprintf("shortcut '%s' already used by '%s'", shortcut, e.Name))
		}
	} else if f.mode == modeAdd {
		shortcut = config.AutoShortcut(name, existing)
	}

	entry := models.CommandEntry{
		Name:        name,
		Command:     cmd,
		ShortcutKey: shortcut,
	}

	return func() tea.Msg {
		var err error
		if f.mode == modeAdd {
			err = addEntry(entry, f.scope)
		} else {
			err = editEntry(f.original, entry, f.scope)
		}
		if err != nil {
			return formErrMsg(err.Error())
		}
		entries, err := config.LoadAll()
		if err != nil {
			return formErrMsg(err.Error())
		}
		return reloadMsg{entries}
	}
}

// formErrMsg carries a validation or save error back into the Update loop.
type formErrMsg string

// Make sure Model handles formErrMsg in its Update.
func init() {
	// Handled in model.go Update via type switch on reloadMsg / formErrMsg.
}

// ── Config mutations ──────────────────────────────────────────────────────────

func addEntry(e models.CommandEntry, scope models.Scope) error {
	if scope == models.ScopeLocal {
		cfg, err := config.LoadLocal()
		if err != nil {
			return err
		}
		cfg.Commands = append(cfg.Commands, e)
		return config.SaveLocal(cfg)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Commands = append(cfg.Commands, e)
	return config.Save(cfg)
}

func editEntry(original, updated models.CommandEntry, scope models.Scope) error {
	// Remove from original scope, add to (possibly different) scope.
	if err := removeEntryFromConfig(original); err != nil {
		return err
	}
	return addEntry(updated, scope)
}

func deleteEntry(e models.CommandEntry) tea.Cmd {
	return func() tea.Msg {
		if err := removeEntryFromConfig(e); err != nil {
			return formErrMsg(err.Error())
		}
		entries, err := config.LoadAll()
		if err != nil {
			return formErrMsg(err.Error())
		}
		return reloadMsg{entries}
	}
}

func removeEntryFromConfig(e models.CommandEntry) error {
	if e.Scope == models.ScopeLocal {
		cfg, err := config.LoadLocal()
		if err != nil {
			return err
		}
		cfg.Commands = removeByName(cfg.Commands, e.Name)
		return config.SaveLocal(cfg)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Commands = removeByName(cfg.Commands, e.Name)
	return config.Save(cfg)
}

func removeByName(cmds []models.CommandEntry, name string) []models.CommandEntry {
	out := cmds[:0]
	for _, c := range cmds {
		if c.Name != name {
			out = append(out, c)
		}
	}
	return out
}

// ── Form view ─────────────────────────────────────────────────────────────────

func (m Model) viewForm() string {
	f := m.form
	title := "add command"
	if f.mode == modeEdit {
		title = "edit  " + boldStyle.Render(f.original.Name)
	}

	marker := func(focused bool) string {
		if focused {
			return lipgloss.NewStyle().Foreground(localAccent).Render("▸ ")
		}
		return "  "
	}

	underline := func(focused bool, value string) string {
		w := 40
		padded := value + strings.Repeat(" ", max(0, w-len([]rune(value))))
		if focused {
			return lipgloss.NewStyle().
				Foreground(localAccent).
				Underline(true).
				Render(padded)
		}
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Underline(true).
			Render(padded)
	}

	var sb strings.Builder
	sb.WriteString(formTitleStyle.Render("rr  ·  "+title) + "\n")
	sb.WriteString(dimStyle.Render("  " + strings.Repeat("─", 54)) + "\n")

	// Name
	sb.WriteString(
		labelStyle.Render("name") +
			marker(f.focus == fieldName) +
			underline(f.focus == fieldName, f.inputs[fieldName].Value()) +
			"\n",
	)
	// Command
	sb.WriteString(
		labelStyle.Render("command") +
			marker(f.focus == fieldCommand) +
			underline(f.focus == fieldCommand, f.inputs[fieldCommand].Value()) +
			"\n",
	)
	// Shortcut
	scVal := f.inputs[fieldShortcut].Value()
	scDisplay := scVal
	if scDisplay == "" {
		scDisplay = dimStyle.Render("auto")
	}
	sb.WriteString(
		labelStyle.Render("shortcut") +
			marker(f.focus == fieldShortcut) +
			underline(f.focus == fieldShortcut, scVal) +
			"  " + dimStyle.Render("(empty = auto-assign)") +
			"\n",
	)
	_ = scDisplay

	// Scope toggle
	localStr := scopeInactiveStyle.Render("local")
	globalStr := scopeInactiveStyle.Render("global")
	if f.scope == models.ScopeLocal {
		localStr = scopeActiveStyle.Render("● local")
	} else {
		globalStr = lipgloss.NewStyle().Foreground(globalAccent).Bold(true).Render("● global")
	}
	sb.WriteString(
		labelStyle.Render("scope") +
			marker(f.focus == fieldScope) +
			localStr + "   " + globalStr +
			dimStyle.Render("  (space to toggle)") +
			"\n",
	)

	sb.WriteString(dimStyle.Render("  " + strings.Repeat("─", 54)) + "\n")

	if f.errMsg != "" {
		sb.WriteString(errorStyle.Render("  ✖ "+f.errMsg) + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  tab/↑↓ navigate   enter save   esc cancel") + "\n")
	}

	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
