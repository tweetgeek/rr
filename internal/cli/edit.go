package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"rr/internal/config"
)

func newEditCmd() *cobra.Command {
	var name, command, shortcut string
	var global bool

	cmd := &cobra.Command{
		Use:   "edit <name|index>",
		Short: "Edit a command entry by name or 1-based index",
		Args:  cobra.ExactArgs(1),
		Example: `  rr edit "Deploy" --command "make deploy-prod"
  rr edit 1 --name "Build" --shortcut b
  rr edit --global "Deploy" --command "make deploy-prod"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("command") && !cmd.Flags().Changed("shortcut") {
				return fmt.Errorf("provide at least one of --name, --command, --shortcut")
			}

			target := args[0]

			if global {
				return editIn(cmd, target, "global", name, command, shortcut)
			}
			// Try local first; fall back to global.
			err := editIn(cmd, target, "local", name, command, shortcut)
			if err != nil {
				return editIn(cmd, target, "global", name, command, shortcut)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "New display name")
	cmd.Flags().StringVarP(&command, "command", "c", "", "New command string")
	cmd.Flags().StringVarP(&shortcut, "shortcut", "s", "", "Override shortcut key (manually set)")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Edit in global config only")

	return cmd
}

func editIn(cmd *cobra.Command, target, scope, name, command, shortcut string) error {
	load := config.LoadLocal
	save := config.SaveLocal
	if scope == "global" {
		load = config.Load
		save = config.Save
	}

	c, err := load()
	if err != nil {
		return err
	}

	idx := findIndex(target, len(c.Commands), func(i int) string { return c.Commands[i].Name })
	if idx == -1 {
		return fmt.Errorf("no command found matching %q in %s config", target, scope)
	}

	entry := &c.Commands[idx]

	if cmd.Flags().Changed("name") {
		entry.Name = name
	}
	if cmd.Flags().Changed("command") {
		entry.Command = command
	}
	if cmd.Flags().Changed("shortcut") {
		// Load all entries to check uniqueness across both configs.
		all, err := config.LoadAll()
		if err != nil {
			return err
		}
		for _, e := range all {
			// Skip the entry being edited.
			if e.Name == entry.Name && string(e.Scope) == scope {
				continue
			}
			if e.ShortcutKey == shortcut {
				return fmt.Errorf("shortcut %q is already used by %q (%s)", shortcut, e.Name, e.Scope)
			}
		}
		entry.ShortcutKey = shortcut
	}

	if err := save(c); err != nil {
		return err
	}

	fmt.Printf("Updated %q [%s]\n", entry.Name, scope)
	return nil
}
