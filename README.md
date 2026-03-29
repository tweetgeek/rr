# rr — Fast Command Runner

A keyboard-driven command launcher with an interactive TUI. Save your frequently used commands, run them instantly from anywhere, or insert them into your shell buffer for editing.

## Installation

```bash
git clone <repo>
cd rr
go build -o rr ./cmd/rr

# Place the binary somewhere on your $PATH (no sudo needed)
mkdir -p ~/.local/bin
cp rr ~/.local/bin/rr

# Make sure ~/.local/bin is in your PATH (add to ~/.zshrc if missing)
export PATH="$HOME/.local/bin:$PATH"
```

## Quick start

```bash
rr add -n "Deploy" -c "make deploy"                  # add a local command
rr add --global -n "Git status" -c "git status"      # add a global command
rr                                                    # open the TUI
```

---

## Command scopes

Commands live in one of two config files:

| Scope | File | When to use |
|-------|------|-------------|
| **local** | `.rr.json` in the current directory | project-specific commands |
| **global** | `~/.config/rr/config.json` (macOS: `~/Library/Application Support/rr/config.json`) | commands available everywhere |

The TUI lists local commands first, then global. Each entry is colour-coded accordingly.

---

## CLI reference

### `rr add`

```bash
rr add -n "Name" -c "command"              # local (default)
rr add --global -n "Name" -c "command"     # global
rr add -n "Name" -c "command" -s x        # override auto-assigned shortcut
```

Shortcut keys are **auto-assigned** from the command name (first free letter, then digits, then the full alphabet). Pass `-s` only when you want a specific key.

### `rr remove`

```bash
rr remove "Name"             # search local first, then global
rr remove 2                  # by 1-based index within the searched scope
rr remove --global "Name"    # global scope only
```

Aliases: `rm`, `delete`, `del`

### `rr edit`

```bash
rr edit "Name" --command "new command"
rr edit 1 --shortcut x
rr edit --global "Name" --name "New name"
```

Only the flags you pass are updated; everything else stays the same.

---

## TUI

Run `rr` without arguments to open the interactive interface.

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `↓` or `k` / `j` | move up / down |
| `1`, `a`, `d`, … | jump instantly to the command with that shortcut key |
| `Enter` | run the selected command (exits with code 0) |
| `Tab` | insert the selected command into the shell buffer (exits with code 2) |
| `Esc` / `Ctrl+C` | quit without doing anything |

### Managing commands from within the TUI

| Key | Action |
|-----|--------|
| `Ctrl+N` | add a new command |
| `Ctrl+E` | edit the selected command |
| `Ctrl+D` | delete the selected command (asks for confirmation) |

### Add / Edit form

| Key | Action |
|-----|--------|
| `Tab` / `↓` | next field |
| `Shift+Tab` / `↑` | previous field |
| `Space` | toggle scope (`local` ↔ `global`) when on the scope field |
| `Enter` | save |
| `Esc` | cancel, return to the list |

Shortcut field: leave empty to auto-assign, or type a single character to set it manually.

---

## zsh integration

Add the snippet below to `~/.zshrc` to open `rr` with a keyboard shortcut directly from the command line.

| TUI key | Shell behaviour |
|---------|-----------------|
| `Enter` | command is executed immediately |
| `Tab` | command is placed in the buffer so you can edit it first |
| `Esc` | nothing happens, prompt is restored |

```zsh
# ── rr integration ────────────────────────────────────────────────────────────
function _rr_widget() {
  # rr renders its TUI directly on /dev/tty so the terminal is not
  # affected by zle's stdout capture. The selected command is written
  # to a temp file and read back after rr exits.
  # exit 0 (Enter)  → execute immediately
  # exit 2 (Tab)    → place in buffer for editing
  # other           → do nothing (Esc / Ctrl+C)
  local tmpfile
  tmpfile=$(mktemp)

  rr --output-file "$tmpfile"
  local ret=$?

  local selected_command
  [[ -s "$tmpfile" ]] && selected_command=$(<"$tmpfile")
  rm -f "$tmpfile"

  if [[ -n "$selected_command" ]]; then
    BUFFER="$selected_command"
    CURSOR=$#BUFFER
    if [[ $ret -eq 0 ]]; then
      zle accept-line   # Enter: execute right away
    fi
    # ret == 2 (Tab): leave in buffer, do not execute
  fi

  zle reset-prompt
}

zle -N _rr_widget
bindkey '^x' _rr_widget   # Ctrl+X — change to any key you prefer
# ─────────────────────────────────────────────────────────────────────────────
```

Reload your config after editing:

```bash
source ~/.zshrc
```

### Alternative key bindings

```zsh
bindkey '^[r'  _rr_widget   # Alt+R
bindkey '^[^R' _rr_widget   # Alt+Shift+R
bindkey '^x^r' _rr_widget   # Ctrl+X, Ctrl+R (chord)
```

> **Note:** if `rr` is not on your `$PATH`, use the full path:
> `rr --output-file "$tmpfile"` → `/path/to/rr --output-file "$tmpfile"`

---

## Project layout

```
cmd/rr/main.go          entry point
internal/
  cli/                  Cobra subcommands (add, remove, edit, root)
  tui/                  Bubble Tea model + form modal
  config/               load / save ~/.config/rr/config.json and .rr.json
  models/               CommandEntry domain type
```

## Building from source

Requires Go 1.21+.

```bash
go build -o rr ./cmd/rr        # build
go vet ./...                   # lint
```
