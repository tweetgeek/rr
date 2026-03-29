package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"rr/internal/config"
)

func newRemoveCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:     "remove <name|index>",
		Aliases: []string{"rm", "delete", "del"},
		Short:   "Remove a command entry by name or 1-based index",
		Args:    cobra.ExactArgs(1),
		Example: `  rr remove "Deploy"           # search local then global
  rr remove 2                  # index within local list
  rr remove --global "Deploy"  # search only global
  rr remove --global 1         # index within global list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			if global {
				return removeFrom(target, "global")
			}
			// Try local first; fall back to global if not found locally.
			err := removeFrom(target, "local")
			if err != nil {
				return removeFrom(target, "global")
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove from global config only")
	return cmd
}

func removeFrom(target, scope string) error {
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

	removed := c.Commands[idx].Name
	c.Commands = append(c.Commands[:idx], c.Commands[idx+1:]...)

	if err := save(c); err != nil {
		return err
	}
	fmt.Printf("Removed %q [%s]\n", removed, scope)
	return nil
}

// findIndex resolves target (numeric 1-based index or name) against a list of size n.
// nameOf returns the name for position i.
func findIndex(target string, n int, nameOf func(int) string) int {
	if num, err := strconv.Atoi(target); err == nil {
		num-- // 1-based → 0-based
		if num < 0 || num >= n {
			return -1
		}
		return num
	}
	for i := 0; i < n; i++ {
		if nameOf(i) == target {
			return i
		}
	}
	return -1
}
