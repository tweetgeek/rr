package tui

import (
	"strings"
	"testing"

	"rr/internal/models"
)

// ── textField ────────────────────────────────────────────────────────────────

func TestTextField_Insert_Append(t *testing.T) {
	f := newTextField("ph", 0)
	f.insert('h')
	f.insert('i')
	if f.Value() != "hi" {
		t.Errorf("expected 'hi', got %q", f.Value())
	}
	if f.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", f.cursor)
	}
}

func TestTextField_Insert_AtMiddle(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("ac")
	f.cursor = 1
	f.insert('b')
	if f.Value() != "abc" {
		t.Errorf("expected 'abc', got %q", f.Value())
	}
	if f.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", f.cursor)
	}
}

func TestTextField_Insert_LimitEnforced(t *testing.T) {
	f := newTextField("", 3)
	f.insert('a')
	f.insert('b')
	f.insert('c')
	f.insert('d') // should be ignored (limit=3)
	if f.Value() != "abc" {
		t.Errorf("expected 'abc' (limit 3), got %q", f.Value())
	}
}

func TestTextField_Backspace_RemovesPreviousChar(t *testing.T) {
	f := newTextField("", 0)
	f.insert('a')
	f.insert('b')
	f.backspace()
	if f.Value() != "a" {
		t.Errorf("expected 'a' after backspace, got %q", f.Value())
	}
	if f.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", f.cursor)
	}
}

func TestTextField_Backspace_AtStart_NoOp(t *testing.T) {
	f := newTextField("", 0)
	f.insert('x')
	f.moveHome()
	f.backspace() // cursor=0, nothing to delete
	if f.Value() != "x" {
		t.Errorf("expected 'x' to remain, got %q", f.Value())
	}
}

func TestTextField_DeleteFwd_RemovesCharAtCursor(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("hello")
	f.cursor = 2
	f.deleteFwd() // removes 'l' at position 2
	if f.Value() != "helo" {
		t.Errorf("expected 'helo', got %q", f.Value())
	}
	if f.cursor != 2 {
		t.Errorf("cursor should not move on deleteFwd, got %d", f.cursor)
	}
}

func TestTextField_DeleteFwd_AtEnd_NoOp(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("hi")
	f.deleteFwd() // cursor at end
	if f.Value() != "hi" {
		t.Errorf("expected 'hi' to remain, got %q", f.Value())
	}
}

func TestTextField_MoveLeft(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("abc")
	f.moveLeft()
	if f.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", f.cursor)
	}
	f.moveLeft()
	f.moveLeft()
	if f.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", f.cursor)
	}
	f.moveLeft() // at start, no-op
	if f.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", f.cursor)
	}
}

func TestTextField_MoveRight(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("ab")
	f.moveHome()
	f.moveRight()
	if f.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", f.cursor)
	}
	f.moveRight()
	if f.cursor != 2 {
		t.Errorf("expected cursor=2, got %d", f.cursor)
	}
	f.moveRight() // at end, no-op
	if f.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", f.cursor)
	}
}

func TestTextField_MoveHomeEnd(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("hello")
	f.moveHome()
	if f.cursor != 0 {
		t.Errorf("moveHome: expected cursor=0, got %d", f.cursor)
	}
	f.moveEnd()
	if f.cursor != 5 {
		t.Errorf("moveEnd: expected cursor=5, got %d", f.cursor)
	}
}

func TestTextField_SetValue_CursorAtEnd(t *testing.T) {
	f := newTextField("", 0)
	f.setValue("hello world")
	if f.Value() != "hello world" {
		t.Errorf("expected 'hello world', got %q", f.Value())
	}
	if f.cursor != len("hello world") {
		t.Errorf("expected cursor at end (%d), got %d", len("hello world"), f.cursor)
	}
}

func TestTextField_Value_Empty(t *testing.T) {
	f := newTextField("placeholder", 0)
	if f.Value() != "" {
		t.Errorf("expected '' for new field, got %q", f.Value())
	}
}

// ── formData — init ──────────────────────────────────────────────────────────

func TestInitAddForm_Defaults(t *testing.T) {
	f := initAddForm()
	if f.mode != modeAdd {
		t.Errorf("expected modeAdd, got %d", f.mode)
	}
	if f.scope != models.ScopeLocal {
		t.Errorf("expected ScopeLocal, got %q", f.scope)
	}
	if f.focus != 0 {
		t.Errorf("expected focus=0, got %d", f.focus)
	}
	if f.errMsg != "" {
		t.Errorf("expected empty errMsg, got %q", f.errMsg)
	}
}

func TestInitEditForm_PreFillsFields(t *testing.T) {
	e := models.CommandEntry{
		Name:        "Deploy",
		Command:     "make deploy",
		ShortcutKey: "d",
		Scope:       models.ScopeGlobal,
	}
	f := initEditForm(e)
	if f.mode != modeEdit {
		t.Errorf("expected modeEdit, got %d", f.mode)
	}
	if f.inputs[fieldName].Value() != "Deploy" {
		t.Errorf("name field: expected 'Deploy', got %q", f.inputs[fieldName].Value())
	}
	if f.inputs[fieldCommand].Value() != "make deploy" {
		t.Errorf("command field: expected 'make deploy', got %q", f.inputs[fieldCommand].Value())
	}
	if f.inputs[fieldShortcut].Value() != "d" {
		t.Errorf("shortcut field: expected 'd', got %q", f.inputs[fieldShortcut].Value())
	}
	if f.scope != models.ScopeGlobal {
		t.Errorf("scope: expected ScopeGlobal, got %q", f.scope)
	}
	if f.original.Name != "Deploy" {
		t.Errorf("original.Name: expected 'Deploy', got %q", f.original.Name)
	}
}

// ── formData — field navigation ──────────────────────────────────────────────

func TestFormData_NextField_Wraps(t *testing.T) {
	f := initAddForm()
	for i := 0; i < fieldCount; i++ {
		f.nextField()
	}
	if f.focus != 0 {
		t.Errorf("expected wrap to focus=0 after %d nextField calls, got %d", fieldCount, f.focus)
	}
}

func TestFormData_PrevField_Wraps(t *testing.T) {
	f := initAddForm()
	f.prevField()
	if f.focus != fieldCount-1 {
		t.Errorf("expected wrap to focus=%d after prevField from 0, got %d", fieldCount-1, f.focus)
	}
}

func TestFormData_NextThenPrev_Identity(t *testing.T) {
	f := initAddForm()
	f.nextField()
	f.prevField()
	if f.focus != 0 {
		t.Errorf("next then prev should return to 0, got %d", f.focus)
	}
}

// ── formData — scope toggle ──────────────────────────────────────────────────

func TestFormData_ToggleScope(t *testing.T) {
	f := initAddForm()
	if f.scope != models.ScopeLocal {
		t.Fatal("expected initial scope=local")
	}
	f.toggleScope()
	if f.scope != models.ScopeGlobal {
		t.Errorf("expected ScopeGlobal after first toggle, got %q", f.scope)
	}
	f.toggleScope()
	if f.scope != models.ScopeLocal {
		t.Errorf("expected ScopeLocal after second toggle, got %q", f.scope)
	}
}

// ── formData — save validation ───────────────────────────────────────────────

func TestFormData_Save_EmptyName_ReturnsError(t *testing.T) {
	f := initAddForm()
	f.inputs[fieldCommand].setValue("echo hi")
	// name is empty
	cmd := f.save(nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	errMsg, ok := msg.(formErrMsg)
	if !ok {
		t.Fatalf("expected formErrMsg, got %T", msg)
	}
	if !strings.Contains(string(errMsg), "name is required") {
		t.Errorf("expected 'name is required', got %q", string(errMsg))
	}
}

func TestFormData_Save_EmptyCommand_ReturnsError(t *testing.T) {
	f := initAddForm()
	f.inputs[fieldName].setValue("Deploy")
	// command is empty
	cmd := f.save(nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	errMsg, ok := msg.(formErrMsg)
	if !ok {
		t.Fatalf("expected formErrMsg, got %T", msg)
	}
	if !strings.Contains(string(errMsg), "command is required") {
		t.Errorf("expected 'command is required', got %q", string(errMsg))
	}
}

func TestFormData_Save_ShortcutConflict_ReturnsError(t *testing.T) {
	existing := []models.CommandEntry{
		{Name: "Build", Command: "go build", ShortcutKey: "b", Scope: models.ScopeLocal},
	}
	f := initAddForm()
	f.inputs[fieldName].setValue("Test")
	f.inputs[fieldCommand].setValue("go test")
	f.inputs[fieldShortcut].setValue("b") // conflicts with Build
	cmd := f.save(existing)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	errMsg, ok := msg.(formErrMsg)
	if !ok {
		t.Fatalf("expected formErrMsg, got %T", msg)
	}
	if !strings.Contains(string(errMsg), "already used") {
		t.Errorf("expected 'already used' in error, got %q", string(errMsg))
	}
}

func TestFormData_Save_EditOwnShortcut_NoConflict(t *testing.T) {
	// Editing an entry should be allowed to keep its own shortcut
	original := models.CommandEntry{
		Name:        "Deploy",
		Command:     "make deploy",
		ShortcutKey: "d",
		Scope:       models.ScopeLocal,
	}
	existing := []models.CommandEntry{original}

	f := initEditForm(original)
	f.inputs[fieldShortcut].setValue("d") // same shortcut as original
	// Should NOT return a conflict error (it returns a Cmd that does file IO)
	cmd := f.save(existing)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	// The cmd will try to do IO — we can't easily test it without a filesystem,
	// but we confirm it's NOT a formErrMsg (conflict check passed)
	msg := cmd()
	if _, isErr := msg.(formErrMsg); isErr {
		t.Errorf("unexpected error for editing own shortcut: %q", string(msg.(formErrMsg)))
	}
}

func TestFormData_Save_WhitespaceOnlyName_ReturnsError(t *testing.T) {
	f := initAddForm()
	f.inputs[fieldName].setValue("   ")
	f.inputs[fieldCommand].setValue("echo hi")
	cmd := f.save(nil)
	msg := cmd()
	errMsg, ok := msg.(formErrMsg)
	if !ok {
		t.Fatalf("expected formErrMsg, got %T", msg)
	}
	if !strings.Contains(string(errMsg), "name is required") {
		t.Errorf("expected 'name is required' for whitespace-only name, got %q", string(errMsg))
	}
}

// ── updateForm key handling ──────────────────────────────────────────────────

func TestUpdateForm_EscReturnsToList(t *testing.T) {
	m := New([]models.CommandEntry{{Name: "T", Command: "t", Scope: models.ScopeLocal}})
	m.state = stateForm
	m.form = initAddForm()
	m.update("esc")
	if m.state != stateList {
		t.Errorf("expected stateList after esc, got %d", m.state)
	}
}

func TestUpdateForm_TabAdvancesField(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.update("tab")
	if m.form.focus != 1 {
		t.Errorf("expected focus=1 after tab, got %d", m.form.focus)
	}
}

func TestUpdateForm_ShiftTabGoesBack(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.form.focus = 2
	m.update("shift+tab")
	if m.form.focus != 1 {
		t.Errorf("expected focus=1 after shift+tab, got %d", m.form.focus)
	}
}

func TestUpdateForm_TypeableCharInsertsIntoField(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.update("d")
	m.update("e")
	m.update("p")
	if m.form.inputs[0].Value() != "dep" {
		t.Errorf("expected 'dep', got %q", m.form.inputs[0].Value())
	}
}

func TestUpdateForm_SpaceOnScopeFieldTogglesScope(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.form.focus = fieldScope
	m.update(" ")
	if m.form.scope != models.ScopeGlobal {
		t.Errorf("expected ScopeGlobal after space on scope field, got %q", m.form.scope)
	}
}

func TestUpdateForm_BackspaceDeletesChar(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.form.inputs[0].setValue("abc")
	m.update("backspace")
	if m.form.inputs[0].Value() != "ab" {
		t.Errorf("expected 'ab' after backspace, got %q", m.form.inputs[0].Value())
	}
}

func TestUpdateForm_ErrMsgClearedOnNextKey(t *testing.T) {
	m := New(nil)
	m.state = stateForm
	m.form = initAddForm()
	m.form.errMsg = "previous error"
	m.update("a")
	if m.form.errMsg != "" {
		t.Errorf("expected errMsg to be cleared, got %q", m.form.errMsg)
	}
}

// ── removeByName ─────────────────────────────────────────────────────────────

func TestRemoveByName_RemovesMatchingEntry(t *testing.T) {
	cmds := []models.CommandEntry{
		{Name: "A", Command: "a"},
		{Name: "B", Command: "b"},
		{Name: "C", Command: "c"},
	}
	result := removeByName(cmds, "B")
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	for _, e := range result {
		if e.Name == "B" {
			t.Error("B should have been removed")
		}
	}
}

func TestRemoveByName_NotFound_ReturnsAll(t *testing.T) {
	cmds := []models.CommandEntry{
		{Name: "A", Command: "a"},
	}
	result := removeByName(cmds, "Z")
	if len(result) != 1 {
		t.Errorf("expected 1 entry when not found, got %d", len(result))
	}
}

func TestRemoveByName_Empty_ReturnsEmpty(t *testing.T) {
	result := removeByName(nil, "X")
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}
