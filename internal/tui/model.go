package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rr/internal/models"
)

// ── Exit codes ────────────────────────────────────────────────────────────────

type exitCode int

const (
	exitClean exitCode = 0
	exitTab   exitCode = 2
)

type selectedMsg struct {
	command string
	code    exitCode
}

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	localAccent  = lipgloss.Color("205")
	globalAccent = lipgloss.Color("69")

	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	sepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))

	localSelStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(localAccent).
			Foreground(localAccent).
			Bold(true)
	globalSelStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(globalAccent).
			Foreground(globalAccent).
			Bold(true)

	nameCol = lipgloss.NewStyle().Width(24)
	cmdCol  = lipgloss.NewStyle().Width(32).Foreground(lipgloss.Color("240"))

	localBadge  = lipgloss.NewStyle().Foreground(localAccent).Faint(true).Render("local")
	globalBadge = lipgloss.NewStyle().Foreground(globalAccent).Faint(true).Render("global")
)

// ── Row types ─────────────────────────────────────────────────────────────────

type row interface{ rowType() }

type entryRow struct{ entry models.CommandEntry }
type sepRow struct{ label string }

func (entryRow) rowType() {}
func (sepRow) rowType()   {}

// ── Model ─────────────────────────────────────────────────────────────────────

const maxVisible = 14

type Model struct {
	rows     []row
	entries  []models.CommandEntry // selectable only, for shortcut lookup
	cursor   int                   // index into rows (always points at an entryRow)
	offset   int                   // first visible row index
	quitting bool
	chosen   *selectedMsg
}

func buildRows(entries []models.CommandEntry) ([]row, int) {
	var locals, globals []models.CommandEntry
	for _, e := range entries {
		if e.Scope == models.ScopeLocal {
			locals = append(locals, e)
		} else {
			globals = append(globals, e)
		}
	}

	both := len(locals) > 0 && len(globals) > 0
	var rows []row
	firstEntry := 0

	if len(locals) > 0 {
		if both {
			rows = append(rows, sepRow{"local"})
			firstEntry = 1
		}
		for _, e := range locals {
			rows = append(rows, entryRow{e})
		}
	}
	if len(globals) > 0 {
		if both {
			rows = append(rows, sepRow{"global"})
		}
		for _, e := range globals {
			rows = append(rows, entryRow{e})
		}
	}
	return rows, firstEntry
}

func New(entries []models.CommandEntry) Model {
	rows, firstEntry := buildRows(entries)
	return Model{
		rows:    rows,
		entries: entries,
		cursor:  firstEntry,
		offset:  0,
	}
}

// ── Navigation helpers ────────────────────────────────────────────────────────

func (m *Model) move(dir int) {
	idx := m.cursor + dir
	for idx >= 0 && idx < len(m.rows) {
		if _, isSep := m.rows[idx].(sepRow); !isSep {
			m.cursor = idx
			m.clampOffset()
			return
		}
		idx += dir
	}
}

func (m *Model) clampOffset() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+maxVisible {
		m.offset = m.cursor - maxVisible + 1
	}
}

func (m Model) selectedEntry() (models.CommandEntry, bool) {
	if m.cursor < len(m.rows) {
		if r, ok := m.rows[m.cursor].(entryRow); ok {
			return r.entry, true
		}
	}
	return models.CommandEntry{}, false
}

// ── Bubble Tea interface ──────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case selectedMsg:
		m.chosen = &msg
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			m.move(-1)

		case "down", "j":
			m.move(1)

		case "enter":
			if e, ok := m.selectedEntry(); ok {
				cmd := e.Command
				return m, func() tea.Msg { return selectedMsg{cmd, exitClean} }
			}

		case "tab":
			if e, ok := m.selectedEntry(); ok {
				cmd := e.Command
				return m, func() tea.Msg { return selectedMsg{cmd, exitTab} }
			}

		default:
			if len(msg.String()) == 1 {
				key := msg.String()
				for _, e := range m.entries {
					if strings.EqualFold(e.ShortcutKey, key) {
						cmd := e.Command
						return m, func() tea.Msg { return selectedMsg{cmd, exitClean} }
					}
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().
		Foreground(localAccent).Bold(true).PaddingLeft(2).
		Render("rr") + "\n")

	end := m.offset + maxVisible
	if end > len(m.rows) {
		end = len(m.rows)
	}

	for i := m.offset; i < end; i++ {
		switch v := m.rows[i].(type) {

		case sepRow:
			label := strings.ToUpper(v.label)
			line := "  " + label + " " + strings.Repeat("─", 40-len(label))
			sb.WriteString(sepStyle.Render(line) + "\n")

		case entryRow:
			e := v.entry
			shortcut := "    "
			if e.ShortcutKey != "" {
				shortcut = fmt.Sprintf("[%s] ", e.ShortcutKey)
			}
			badge := localBadge
			selStyle := localSelStyle
			if e.Scope == models.ScopeGlobal {
				badge = globalBadge
				selStyle = globalSelStyle
			}
			name := nameCol.Render(shortcut + e.Name)
			cmd := cmdCol.Render(e.Command)

			if i == m.cursor {
				sb.WriteString(selStyle.Render(name+cmd+" "+badge) + "\n")
			} else {
				sb.WriteString("  " + dimStyle.Render(name) + cmd + " " + badge + "\n")
			}
		}
	}

	sb.WriteString(dimStyle.Render("  ↑/↓  enter run  tab insert  esc quit") + "\n")
	return sb.String()
}

// ── Run ───────────────────────────────────────────────────────────────────────

func Run(entries []models.CommandEntry, outputFile string) {
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No commands configured.")
		fmt.Fprintln(os.Stderr, "  local:   rr add -n \"Name\" -c \"command\"")
		fmt.Fprintln(os.Stderr, "  global:  rr add --global -n \"Name\" -c \"command\"")
		os.Exit(1)
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	m := New(entries)
	p := tea.NewProgram(m,
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}

	final, ok := result.(Model)
	if !ok || final.chosen == nil {
		os.Exit(0)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(final.chosen.command), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "cannot write output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(final.chosen.command)
	}

	os.Exit(int(final.chosen.code))
}
