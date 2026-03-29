package cli

import (
	"github.com/spf13/cobra"

	"rr/internal/config"
	"rr/internal/tui"
)

// NewRootCmd builds and returns the root cobra command.
func NewRootCmd() *cobra.Command {
	var outputFile string

	root := &cobra.Command{
		Use:   "rr",
		Short: "Fast Command Runner — pick and run saved commands",
		Long: `rr lets you save frequently used commands and launch them instantly
via an interactive TUI or keyboard shortcuts.

Commands are scoped:
  local   — stored in .rr.json in the current directory (project-specific)
  global  — stored in ~/.config/rr/config.json (available everywhere)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := config.LoadAll()
			if err != nil {
				return err
			}
			tui.Run(entries, outputFile)
			return nil
		},
	}

	// Hidden flag used by the zsh widget to receive the selected command
	// without polluting stdout (which zle may capture or discard).
	root.Flags().StringVar(&outputFile, "output-file", "", "Write selected command to this file instead of stdout")
	_ = root.Flags().MarkHidden("output-file")

	root.AddCommand(newAddCmd())
	root.AddCommand(newRemoveCmd())
	root.AddCommand(newEditCmd())

	return root
}
