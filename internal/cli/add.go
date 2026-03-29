package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"rr/internal/config"
	"rr/internal/models"
)

func newAddCmd() *cobra.Command {
	var name, command, shortcut string
	var global bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new command entry",
		Example: `  rr add -n "Deploy" -c "make deploy"          # local (project)
  rr add -n "Git status" -c "git status" --global  # global (everywhere)
  rr add -n "Tests" -c "go test ./..." -s t        # local with manual shortcut`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if command == "" {
				return fmt.Errorf("--command is required")
			}

			// Load both configs to check shortcut uniqueness across all entries.
			all, err := config.LoadAll()
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("shortcut") {
				for _, e := range all {
					if e.ShortcutKey == shortcut {
						return fmt.Errorf("shortcut %q is already used by %q (%s)", shortcut, e.Name, e.Scope)
					}
				}
			} else {
				shortcut = config.AutoShortcut(name, all)
			}

			entry := models.CommandEntry{
				Name:        name,
				Command:     command,
				ShortcutKey: shortcut,
			}

			var scope string
			if global {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				cfg.Commands = append(cfg.Commands, entry)
				if err := config.Save(cfg); err != nil {
					return err
				}
				scope = "global"
			} else {
				cfg, err := config.LoadLocal()
				if err != nil {
					return err
				}
				cfg.Commands = append(cfg.Commands, entry)
				if err := config.SaveLocal(cfg); err != nil {
					return err
				}
				scope = "local"
			}

			if shortcut != "" {
				fmt.Printf("Added %q [%s] (shortcut: %s)\n", name, scope, shortcut)
			} else {
				fmt.Printf("Added %q [%s] (no shortcut available)\n", name, scope)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Display name for the command (required)")
	cmd.Flags().StringVarP(&command, "command", "c", "", "The command string to execute (required)")
	cmd.Flags().StringVarP(&shortcut, "shortcut", "s", "", "Override auto-assigned shortcut key")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Add to global config (default: local project config)")

	return cmd
}
