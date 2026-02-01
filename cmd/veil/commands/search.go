package commands

import (
	"fmt"
	"os"
)

// SearchCommand searches for secrets matching a pattern across all vaults.
type SearchCommand struct {
	BaseCommand
}

func NewSearchCommand() *SearchCommand {
	return &SearchCommand{
		BaseCommand: NewBaseCommand("search", "Search secrets across all vaults"),
	}
}

func (c *SearchCommand) Execute(args []string, deps Dependencies) error {
	if len(args) < 1 {
		return &UsageError{
			Command: "search",
			Usage:   "veil search <pattern>",
		}
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	pattern := args[0]

	results, err := deps.App.Search(pattern)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(stdout, "No matches found")
		return nil
	}

	fmt.Fprintf(stdout, "Found %d match%s:\n", len(results), plural(len(results)))
	for _, ref := range results {
		fmt.Fprintf(stdout, "  %s/%s\n", ref.Vault, ref.Name)
	}

	return nil
}

// plural returns "es" for counts != 1, empty string otherwise.
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "es"
}

func init() {
	Register(NewSearchCommand())
}
