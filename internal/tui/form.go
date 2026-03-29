package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"rr/internal/config"
	"rr/internal/models"
)

// ── textField — minimal single-line editor ────────────────────────────────────

type textField struct {
	value       []rune
	cursor      int
	limit       int    // 0 = unlimited
	placeholder string
}

func newTextField(placeholder string, limit int) textField {
	return textField{placeholder: placeholder, limit: limit}
}

func (t *textField) setValue(s string) {
	t.value = []rune(s)
	t.cursor = len(t.value)
}

func (t textField) Value() string { return string(t.value) }

func (t *textField) insert(r rune) {
	if t.limit > 0 && len(t.value) >= t.limit {
		return
	}
	v := t.value
	t.value = append(v[:t.cursor:t.cursor], append([]rune{r}, v[t.cursor:]...)...)
	t.cursor++
}

func (t *textField) backspace() {
	if t.cursor > 0 {
		v := t.value
		t.value = append(v[:t.cursor-1:t.cursor-1], v[t.cursor:]...)
		t.cursor--
	}
}

func (t *textField) deleteFwd() {
	if t.cursor < len(t.value) {
		v := t.value
		t.value = append(v[:t.cursor:t.cursor], v[t.cursor+1:]...)
	}
}

func (t *textField) moveLeft()  { if t.cursor > 0 { t.cursor-- } }
func (t *textField) moveRight() { if t.cursor < len(t.value) { t.cursor++ } }
func (t *textField) moveHome()  { t.cursor = 0 }
func (t *textField) moveEnd()   { t.cursor = len(t.value) }

// view renders the field as a fixed-width string.
// focused=true shows a block cursor; focused=false shows the value dimmed or placeholder.
func (t textField) view(focused bool, width int) string {
	if !focused {
		val := string(t.value)
		if val == "" {
			val = dimStyle.Render(t.placeholder)
		}
		return lipgloss.NewStyle().Width(width).Render(val)
	}

	// Split around cursor for cursor rendering
	before := string(t.value[:t.cursor])
	var cursorChar, after string
	if t.cursor < len(t.value) {
		cursorChar = lipgloss.NewStyle().Reverse(true).Foreground(localAccent).Render(string(t.value[t.cursor]))
		after = string(t.value[t.cursor+1:])
	} else {
		cursorChar = lipgloss.NewStyle().Reverse(true).Render(" ")
	}

	content := lipgloss.NewStyle().Foreground(localAccent).Render(before) +
		cursorChar +
		lipgloss.NewStyle().Foreground(localAccent).Render(after)

	return lipgloss.NewStyle().Width(width).Render(content)
}

// ── Form mode & field indices ─────────────────────────────────────────────────

type formMode int

const (
	modeAdd  formMode = iota
	modeEdit
)

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
	original models.CommandEntry
	inputs   [3]textField
	focus    int
	scope    models.Scope
	errMsg   string
}

func initAddForm() formData {
	return formData{
		mode:  modeAdd,
		scope: models.ScopeLocal,
		inputs: [3]textField{
			newTextField("required", 80),
			newTextField("required", 256),
			newTextField("auto", 1),
		},
	}
}

func initEditForm(e models.CommandEntry) formData {
	f := initAddForm()
	f.mode = modeEdit
	f.original = e
	f.scope = e.Scope
	f.inputs[fieldName].setValue(e.Name)
	f.inputs[fieldCommand].setValue(e.Command)
	f.inputs[fieldShortcut].setValue(e.ShortcutKey)
	return f
}

func (f *formData) nextField() { f.focus = (f.focus + 1) % fieldCount }
func (f *formData) prevField() { f.focus = (f.focus - 1 + fieldCount) % fieldCount }

func (f *formData) toggleScope() {
	if f.scope == models.ScopeLocal {
		f.scope = models.ScopeGlobal
	} else {
		f.scope = models.ScopeLocal
	}
}

// ── Form update ───────────────────────────────────────────────────────────────

func (m *Model) updateForm(key string) Cmd {
	f := &m.form
	f.errMsg = ""

	switch key {
	case "esc", "ctrl+c":
		m.state = stateList
		return nil

	case "tab", "down":
		f.nextField()
	case "shift+tab", "up":
		f.prevField()

	case "enter":
		if f.focus >= fieldShortcut {
			return f.save(m.entries)
		}
		f.nextField()

	case " ":
		if f.focus == fieldScope {
			f.toggleScope()
		} else {
			f.inputs[f.focus].insert(' ')
		}

	case "backspace":
		if f.focus < fieldScope {
			f.inputs[f.focus].backspace()
		}
	case "delete":
		if f.focus < fieldScope {
			f.inputs[f.focus].deleteFwd()
		}
	case "left":
		if f.focus < fieldScope {
			f.inputs[f.focus].moveLeft()
		}
	case "right":
		if f.focus < fieldScope {
			f.inputs[f.focus].moveRight()
		}

	default:
		if len(key) == 1 && f.focus < fieldScope {
			f.inputs[f.focus].insert([]rune(key)[0])
		}
	}
	return nil
}

// ── Form save ─────────────────────────────────────────────────────────────────

func (f *formData) save(existing []models.CommandEntry) Cmd {
	name := strings.TrimSpace(f.inputs[fieldName].Value())
	cmd := strings.TrimSpace(f.inputs[fieldCommand].Value())
	shortcut := strings.TrimSpace(f.inputs[fieldShortcut].Value())

	errCmd := func(msg string) Cmd {
		return func() Msg { return formErrMsg(msg) }
	}

	if name == "" {
		return errCmd("name is required")
	}
	if cmd == "" {
		return errCmd("command is required")
	}

	if shortcut != "" {
		for _, e := range existing {
			if e.ShortcutKey != shortcut {
				continue
			}
			if f.mode == modeEdit && e.Name == f.original.Name && e.Scope == f.original.Scope {
				continue
			}
			return errCmd(fmt.Sprintf("shortcut '%s' already used by '%s'", shortcut, e.Name))
		}
	} else if f.mode == modeAdd {
		shortcut = config.AutoShortcut(name, existing)
	}

	entry := models.CommandEntry{Name: name, Command: cmd, ShortcutKey: shortcut}
	scope := f.scope
	mode := f.mode
	original := f.original

	return func() Msg {
		var err error
		if mode == modeAdd {
			err = addEntry(entry, scope)
		} else {
			err = editEntry(original, entry, scope)
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
	if err := removeEntryFromConfig(original); err != nil {
		return err
	}
	return addEntry(updated, scope)
}

func deleteEntry(e models.CommandEntry) Cmd {
	return func() Msg {
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

var (
	formTitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(localAccent).PaddingLeft(2)
	labelStyle         = lipgloss.NewStyle().Width(11).Foreground(lipgloss.Color("250"))
	scopeActiveStyle   = lipgloss.NewStyle().Foreground(localAccent).Bold(true)
	scopeInactiveStyle = dimStyle
)

func (m *Model) viewForm() string {
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

	const fieldW = 40
	var sb strings.Builder

	sb.WriteString(formTitleStyle.Render("rr  ·  "+title) + "\n")
	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", 54)) + "\n")

	// Text input fields
	defs := [3]struct{ label string }{{"name"}, {"command"}, {"shortcut"}}
	for i, d := range defs {
		focused := f.focus == i
		hint := ""
		if i == fieldShortcut {
			hint = "  " + dimStyle.Render("(empty = auto)")
		}
		sb.WriteString(
			labelStyle.Render(d.label) +
				marker(focused) +
				f.inputs[i].view(focused, fieldW) +
				hint + "\n",
		)
	}

	// Scope toggle
	localStr := scopeInactiveStyle.Render("  local")
	globalStr := scopeInactiveStyle.Render("  global")
	if f.scope == models.ScopeLocal {
		localStr = scopeActiveStyle.Render("● local")
	} else {
		globalStr = lipgloss.NewStyle().Foreground(globalAccent).Bold(true).Render("● global")
	}
	sb.WriteString(
		labelStyle.Render("scope") +
			marker(f.focus == fieldScope) +
			localStr + "   " + globalStr +
			"  " + dimStyle.Render("(space)") + "\n",
	)

	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", 54)) + "\n")

	if f.errMsg != "" {
		sb.WriteString(errorStyle.Render("  ✖ "+f.errMsg) + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  tab/↑↓ navigate   enter save   esc cancel") + "\n")
	}

	return sb.String()
}
