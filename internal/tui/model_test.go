package tui

import (
	"testing"

	"rr/internal/models"
)

// ── buildRows ────────────────────────────────────────────────────────────────

func TestBuildRows_Empty(t *testing.T) {
	rows, firstEntry := buildRows(nil)
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
	if firstEntry != 0 {
		t.Errorf("expected firstEntry=0, got %d", firstEntry)
	}
}

func TestBuildRows_OnlyLocal_NoSeparators(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Command: "a", Scope: models.ScopeLocal},
		{Name: "B", Command: "b", Scope: models.ScopeLocal},
	}
	rows, firstEntry := buildRows(entries)
	if firstEntry != 0 {
		t.Errorf("expected firstEntry=0, got %d", firstEntry)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for i, r := range rows {
		if _, ok := r.(entryRow); !ok {
			t.Errorf("rows[%d]: expected entryRow, got %T", i, r)
		}
	}
}

func TestBuildRows_OnlyGlobal_NoSeparators(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "G", Command: "g", Scope: models.ScopeGlobal},
	}
	rows, firstEntry := buildRows(entries)
	if firstEntry != 0 {
		t.Errorf("expected firstEntry=0, got %d", firstEntry)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if _, ok := rows[0].(entryRow); !ok {
		t.Errorf("expected entryRow, got %T", rows[0])
	}
}

func TestBuildRows_BothScopes_InsertsSeparators(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Local", Command: "l", Scope: models.ScopeLocal},
		{Name: "Global", Command: "g", Scope: models.ScopeGlobal},
	}
	rows, firstEntry := buildRows(entries)
	// Expected: [sepRow(local), entryRow(Local), sepRow(global), entryRow(Global)]
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	if firstEntry != 1 {
		t.Errorf("expected firstEntry=1 (skip local sep), got %d", firstEntry)
	}
	if _, ok := rows[0].(sepRow); !ok {
		t.Errorf("rows[0]: expected sepRow, got %T", rows[0])
	}
	if _, ok := rows[1].(entryRow); !ok {
		t.Errorf("rows[1]: expected entryRow, got %T", rows[1])
	}
	if _, ok := rows[2].(sepRow); !ok {
		t.Errorf("rows[2]: expected sepRow, got %T", rows[2])
	}
	if _, ok := rows[3].(entryRow); !ok {
		t.Errorf("rows[3]: expected entryRow, got %T", rows[3])
	}
}

func TestBuildRows_SeparatorLabels(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "L", Command: "l", Scope: models.ScopeLocal},
		{Name: "G", Command: "g", Scope: models.ScopeGlobal},
	}
	rows, _ := buildRows(entries)
	if rows[0].(sepRow).label != "local" {
		t.Errorf("expected local sep label='local', got %q", rows[0].(sepRow).label)
	}
	if rows[2].(sepRow).label != "global" {
		t.Errorf("expected global sep label='global', got %q", rows[2].(sepRow).label)
	}
}

// ── New / cursor ─────────────────────────────────────────────────────────────

func TestNew_SingleEntry_CursorAtZero(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Test", Command: "go test", Scope: models.ScopeLocal},
	}
	m := New(entries)
	if m.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", m.cursor)
	}
}

func TestNew_BothScopes_CursorSkipsSep(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Local", Scope: models.ScopeLocal},
		{Name: "Global", Scope: models.ScopeGlobal},
	}
	m := New(entries)
	// rows[0] = sepRow(local), rows[1] = entryRow(Local), …
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 (after local separator), got %d", m.cursor)
	}
}

// ── Navigation ───────────────────────────────────────────────────────────────

func TestMove_DownAndUp(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Scope: models.ScopeLocal},
		{Name: "B", Scope: models.ScopeLocal},
		{Name: "C", Scope: models.ScopeLocal},
	}
	m := New(entries)
	m.move(1)
	if m.cursor != 1 {
		t.Errorf("after move(1): expected cursor=1, got %d", m.cursor)
	}
	m.move(1)
	if m.cursor != 2 {
		t.Errorf("after move(1): expected cursor=2, got %d", m.cursor)
	}
	m.move(-1)
	if m.cursor != 1 {
		t.Errorf("after move(-1): expected cursor=1, got %d", m.cursor)
	}
}

func TestMove_CannotGoAboveFirst(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Scope: models.ScopeLocal},
	}
	m := New(entries)
	m.move(-1)
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 when moving up from top, got %d", m.cursor)
	}
}

func TestMove_CannotGoBelowLast(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Scope: models.ScopeLocal},
		{Name: "B", Scope: models.ScopeLocal},
	}
	m := New(entries)
	m.move(1) // to B
	m.move(1) // past end — should stay
	if m.cursor != 1 {
		t.Errorf("cursor should stay at last entry, got %d", m.cursor)
	}
}

func TestMove_SkipsSeparators(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Local", Scope: models.ScopeLocal},
		{Name: "Global", Scope: models.ScopeGlobal},
	}
	m := New(entries)
	// cursor=1 (Local entry); rows[2] is sepRow(global), rows[3] is Global entry
	m.move(1)
	if m.cursor != 3 {
		t.Errorf("expected cursor=3 (skipped global sep), got %d", m.cursor)
	}
	m.move(-1)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 (skipped global sep going up), got %d", m.cursor)
	}
}

// ── selectedEntry ────────────────────────────────────────────────────────────

func TestSelectedEntry_ReturnsCurrentEntry(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Test", Command: "go test", Scope: models.ScopeLocal},
	}
	m := New(entries)
	e, ok := m.selectedEntry()
	if !ok {
		t.Fatal("selectedEntry() returned ok=false")
	}
	if e.Name != "Test" || e.Command != "go test" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestSelectedEntry_EmptyModel_ReturnsFalse(t *testing.T) {
	m := New(nil)
	_, ok := m.selectedEntry()
	if ok {
		t.Error("expected ok=false for empty model")
	}
}

// ── updateList key handling ───────────────────────────────────────────────────

func TestUpdateList_EscQuits(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "t", Scope: models.ScopeLocal}})
	m.update("esc")
	if !m.quitting {
		t.Error("expected quitting=true after esc")
	}
	if m.chosen != nil {
		t.Error("expected chosen=nil after esc")
	}
}

func TestUpdateList_CtrlCQuits(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "t", Scope: models.ScopeLocal}})
	m.update("ctrl+c")
	if !m.quitting {
		t.Error("expected quitting=true after ctrl+c")
	}
}

func TestUpdateList_EnterReturnsSelectedMsgExitClean(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "go test", Scope: models.ScopeLocal}})
	cmd := m.update("enter")
	if cmd == nil {
		t.Fatal("expected non-nil Cmd after enter")
	}
	msg := cmd()
	sel, ok := msg.(selectedMsg)
	if !ok {
		t.Fatalf("expected selectedMsg, got %T", msg)
	}
	if sel.command != "go test" {
		t.Errorf("expected 'go test', got %q", sel.command)
	}
	if sel.code != exitClean {
		t.Errorf("expected exitClean, got %d", sel.code)
	}
}

func TestUpdateList_TabReturnsSelectedMsgExitTab(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "go test", Scope: models.ScopeLocal}})
	cmd := m.update("tab")
	if cmd == nil {
		t.Fatal("expected non-nil Cmd after tab")
	}
	msg := cmd()
	sel, ok := msg.(selectedMsg)
	if !ok {
		t.Fatalf("expected selectedMsg, got %T", msg)
	}
	if sel.code != exitTab {
		t.Errorf("expected exitTab, got %d", sel.code)
	}
}

func TestUpdateList_ShortcutKeyDispatch(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Build", Command: "go build", ShortcutKey: "b", Scope: models.ScopeLocal},
		{Name: "Test", Command: "go test", ShortcutKey: "t", Scope: models.ScopeLocal},
	}
	m := New(entries)
	cmd := m.update("t")
	if cmd == nil {
		t.Fatal("expected non-nil Cmd after shortcut 't'")
	}
	msg := cmd()
	sel, ok := msg.(selectedMsg)
	if !ok {
		t.Fatalf("expected selectedMsg, got %T", msg)
	}
	if sel.command != "go test" {
		t.Errorf("expected 'go test', got %q", sel.command)
	}
	if sel.code != exitClean {
		t.Errorf("expected exitClean, got %d", sel.code)
	}
}

func TestUpdateList_UnknownShortcutDoesNothing(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "t", ShortcutKey: "t", Scope: models.ScopeLocal}})
	cmd := m.update("z") // no entry with shortcut 'z'
	if cmd != nil {
		t.Error("expected nil Cmd for unknown single-char key")
	}
}

func TestUpdateList_CtrlNOpensAddForm(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "t", Scope: models.ScopeLocal}})
	m.update("ctrl+n")
	if m.state != stateForm {
		t.Errorf("expected stateForm, got %d", m.state)
	}
	if m.form.mode != modeAdd {
		t.Errorf("expected modeAdd, got %d", m.form.mode)
	}
}

func TestUpdateList_CtrlEOpensEditForm(t *testing.T) {
	entries := []models.CommandEntry{{Name: "Deploy", Command: "make deploy", Scope: models.ScopeLocal}}
	m := New(entries)
	m.update("ctrl+e")
	if m.state != stateForm {
		t.Errorf("expected stateForm, got %d", m.state)
	}
	if m.form.mode != modeEdit {
		t.Errorf("expected modeEdit, got %d", m.form.mode)
	}
	if m.form.inputs[fieldName].Value() != "Deploy" {
		t.Errorf("edit form should pre-fill name 'Deploy', got %q", m.form.inputs[fieldName].Value())
	}
}

func TestUpdateList_CtrlDOpensDeleteConfirm(t *testing.T) {
	entries := []models.CommandEntry{{Name: "Deploy", Command: "make deploy", Scope: models.ScopeLocal}}
	m := New(entries)
	m.update("ctrl+d")
	if m.state != stateDeleteConfirm {
		t.Errorf("expected stateDeleteConfirm, got %d", m.state)
	}
	if m.delEntry.Name != "Deploy" {
		t.Errorf("expected delEntry.Name='Deploy', got %q", m.delEntry.Name)
	}
}

func TestUpdateList_UpDownKeys(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Scope: models.ScopeLocal},
		{Name: "B", Scope: models.ScopeLocal},
	}
	m := New(entries)
	m.update("down")
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.cursor)
	}
	m.update("up")
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after up, got %d", m.cursor)
	}
	m.update("j") // vim down
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after 'j', got %d", m.cursor)
	}
	m.update("k") // vim up
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after 'k', got %d", m.cursor)
	}
}

// ── updateDelete ─────────────────────────────────────────────────────────────

func TestUpdateDelete_NReturnsList(t *testing.T) {
	entries := []models.CommandEntry{{Name: "X", Command: "x", Scope: models.ScopeLocal}}
	m := New(entries)
	m.update("ctrl+d")
	m.update("n")
	if m.state != stateList {
		t.Errorf("expected stateList after 'n', got %d", m.state)
	}
}

func TestUpdateDelete_EscReturnsList(t *testing.T) {
	entries := []models.CommandEntry{{Name: "X", Command: "x", Scope: models.ScopeLocal}}
	m := New(entries)
	m.update("ctrl+d")
	m.update("esc")
	if m.state != stateList {
		t.Errorf("expected stateList after esc, got %d", m.state)
	}
}

// ── dispatch ─────────────────────────────────────────────────────────────────

func TestDispatch_NilCmd_NoOp(t *testing.T) {
	m := New(nil)
	m.dispatch(nil) // must not panic
}

func TestDispatch_SelectedMsg_SetsChosen(t *testing.T) {
	m := New(nil)
	m.dispatch(func() Msg { return selectedMsg{"echo hi", exitClean} })
	if m.chosen == nil {
		t.Fatal("expected chosen to be set")
	}
	if m.chosen.command != "echo hi" {
		t.Errorf("expected 'echo hi', got %q", m.chosen.command)
	}
	if !m.quitting {
		t.Error("expected quitting=true")
	}
}

func TestDispatch_ReloadMsg_UpdatesEntries(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "Old", Scope: models.ScopeLocal}})
	m.state = stateForm
	newEntries := []models.CommandEntry{{Name: "New", Command: "new", Scope: models.ScopeLocal}}
	m.dispatch(func() Msg { return reloadMsg{newEntries} })
	if m.state != stateList {
		t.Errorf("expected stateList after reload, got %d", m.state)
	}
	if len(m.entries) != 1 || m.entries[0].Name != "New" {
		t.Errorf("entries not updated: %+v", m.entries)
	}
}

func TestDispatch_FormErrMsg_SetsErrMsg(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.dispatch(func() Msg { return formErrMsg("something went wrong") })
	if m.form.errMsg != "something went wrong" {
		t.Errorf("expected errMsg='something went wrong', got %q", m.form.errMsg)
	}
}

// ── reload ───────────────────────────────────────────────────────────────────

func TestReload_CursorClamped(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "A", Scope: models.ScopeLocal},
		{Name: "B", Scope: models.ScopeLocal},
		{Name: "C", Scope: models.ScopeLocal},
	}
	m := New(entries)
	m.cursor = 2 // on C
	// Reload with only 1 entry — cursor should be clamped to 0
	m.reload([]models.CommandEntry{{Name: "X", Scope: models.ScopeLocal}})
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after reload with fewer entries, got %d", m.cursor)
	}
}

func TestReload_CursorSkipsSeparator(t *testing.T) {
	entries := []models.CommandEntry{
		{Name: "Local", Scope: models.ScopeLocal},
		{Name: "Global", Scope: models.ScopeGlobal},
	}
	m := New(entries)
	m.cursor = 0 // would land on sepRow after reload
	m.reload(entries)
	// cursor should advance past the separator
	if _, isSep := m.rows[m.cursor].(sepRow); isSep {
		t.Errorf("cursor=%d is on a sepRow after reload", m.cursor)
	}
}
