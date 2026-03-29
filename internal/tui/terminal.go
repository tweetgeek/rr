package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// openTTY opens /dev/tty directly so the TUI works even when
// stdout/stdin are redirected (e.g. inside a zsh widget).
func openTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

// withRawMode puts fd into raw mode and returns a restore function.
func withRawMode(fd int) (restore func(), err error) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func() { term.Restore(fd, state) }, nil //nolint:errcheck
}

// readKey reads one logical keypress from tty.
//
// We read up to 8 bytes in a single call. In raw mode (VMIN=1, VTIME=0) the
// OS blocks until at least 1 byte is ready, then returns ALL bytes currently
// in the kernel tty buffer — typically the full escape sequence for arrow keys
// (\x1b[A = 3 bytes) arrives atomically, so a single Read captures it whole.
// This avoids the unreliable SetNonblock dance on /dev/tty on macOS.
func readKey(tty *os.File) string {
	buf := make([]byte, 8)
	n, err := tty.Read(buf)
	if err != nil || n == 0 {
		return ""
	}
	b := buf[:n]

	// Multi-byte escape sequences
	if n >= 3 && b[0] == 0x1b && b[1] == '[' {
		switch b[2] {
		case 'A':
			return "up"
		case 'B':
			return "down"
		case 'C':
			return "right"
		case 'D':
			return "left"
		case 'Z':
			return "shift+tab"
		case '3':
			if n >= 4 && b[3] == '~' {
				return "delete"
			}
		}
	}

	// Single-byte controls and printable characters
	switch b[0] {
	case 0x03:
		return "ctrl+c"
	case 0x04:
		return "ctrl+d"
	case 0x05:
		return "ctrl+e"
	case 0x0e:
		return "ctrl+n"
	case 0x09:
		return "tab"
	case 0x0d:
		return "enter"
	case 0x08, 0x7f:
		return "backspace"
	case 0x1b:
		return "esc"
	default:
		if b[0] >= 32 && b[0] < 127 {
			return string(rune(b[0]))
		}
	}
	return ""
}

// renderFrame writes view to w, erasing the previous render first.
// prevLines tracks the line count of the last frame so we can
// move the cursor back up before overwriting.
func renderFrame(w io.Writer, view string, prevLines *int) {
	lines := strings.Split(strings.TrimRight(view, "\n"), "\n")
	if *prevLines > 0 {
		fmt.Fprintf(w, "\033[%dA", *prevLines)
	}
	for _, line := range lines {
		fmt.Fprintf(w, "\033[2K\r%s\n", line)
	}
	*prevLines = len(lines)
}

// clearFrame removes the last rendered frame from the terminal,
// leaving the cursor at the position before the first render.
func clearFrame(w io.Writer, lines int) {
	if lines > 0 {
		fmt.Fprintf(w, "\033[%dA", lines)
		for i := 0; i < lines; i++ {
			fmt.Fprintf(w, "\033[2K\r\n")
		}
		fmt.Fprintf(w, "\033[%dA", lines)
	}
}
