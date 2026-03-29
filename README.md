# rr — Fast Command Runner

Błyskawiczny launcher zapisanych komend z interaktywnym TUI i skrótami klawiszowymi.

## Instalacja

```bash
go build -o rr ./cmd/rr
# Opcjonalnie: skopiuj binarny do miejsca w $PATH
rr
```

## Użycie

```bash
rr                      # otwiera TUI
rr add -n "Deploy" -c "make deploy"              # dodaj lokalną komendę
rr add --global -n "Git status" -c "git status"  # dodaj globalną komendę
rr remove "Deploy"
rr edit "Deploy" --command "make deploy-prod"
```

## Konfiguracja

### Zakresy komend

| Zakres | Plik | Kiedy używać |
|--------|------|--------------|
| **local** | `.rr.json` w bieżącym katalogu | komendy specyficzne dla projektu |
| **global** | `~/.config/rr/config.json` | komendy dostępne wszędzie |

TUI wyświetla najpierw lokalne, potem globalne. Każda pozycja oznaczona jest kolorowym badge'em (`local` / `global`).

### Subkomendy

```bash
# Dodawanie
rr add -n "Nazwa" -c "komenda"              # lokalny (domyślnie)
rr add --global -n "Nazwa" -c "komenda"     # globalny
rr add -n "Nazwa" -c "komenda" -s x        # ręczny skrót (domyślnie auto)

# Usuwanie — szuka lokalnie, potem globalnie
rr remove "Nazwa"
rr remove 2                   # wg indeksu (1-based) w danym zakresie
rr remove --global "Nazwa"    # tylko globalny

# Edycja
rr edit "Nazwa" --command "nowa komenda"
rr edit 1 --shortcut x
rr edit --global "Nazwa" --name "Nowa nazwa"
```

### TUI — klawiatura

| Klawisz | Akcja |
|---------|-------|
| `↑` / `↓` | nawigacja |
| `1`, `a`, `d`, … | natychmiastowy wybór po skrócie |
| `Enter` | uruchom komendę (exit 0) |
| `Tab` | wstaw do bufora (exit 2) |
| `Esc` / `Ctrl+C` | wyjdź bez akcji |

---

## Integracja z zsh

Dodaj poniższy snippet do `~/.zshrc`, aby wywołać `rr` skrótem klawiszowym.

- **`Enter`** — komenda zostaje natychmiast wykonana
- **`Tab`** — komenda trafia do bufora wiersza poleceń (możesz ją edytować przed wykonaniem)
- **`Esc`** — nic się nie dzieje, wracasz do prompta

```zsh
# ── rr integration ────────────────────────────────────────────────────────────
function _rr_widget() {
  # TUI renderuje się na /dev/tty (nie przechwytywane przez shell).
  # Wynik trafia do pliku tymczasowego przez --output-file,
  # co omija wszelkie problemy z stdout w kontekście zle.
  # Exit 0  → Enter  → wykonaj komendę od razu
  # Exit 2  → Tab    → wstaw do bufora do edycji
  # Inne    → Esc/Ctrl+C → nic nie rób
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
      zle accept-line   # Enter: od razu wykonaj
    fi
    # ret == 2 (Tab): zostaw w buforze do edycji, nie wykonuj
  fi

  zle reset-prompt
}

zle -N _rr_widget
bindkey '^x' _rr_widget   # Ctrl+X — zmień na dowolny inny skrót
# ─────────────────────────────────────────────────────────────────────────────
```

> **Uwaga:** jeśli `rr` nie jest w `$PATH`, podaj pełną ścieżkę do binarki,
> np. `selected_command=$(/usr/local/bin/rr)`.

### Alternatywne skróty

```zsh
bindkey '^[r'  _rr_widget   # Alt+R
bindkey '^[^R' _rr_widget   # Alt+Shift+R
bindkey '^x^r' _rr_widget   # Ctrl+X, Ctrl+R (chord)
```

Po edycji `.zshrc` przeładuj konfigurację:

```bash
source ~/.zshrc
```
