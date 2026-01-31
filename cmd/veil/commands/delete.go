package commands

// DeleteCommand removes a secret from a vault.
type DeleteCommand struct {
	BaseCommand
}

func NewDeleteCommand() *DeleteCommand {
	return &DeleteCommand{
		BaseCommand: NewBaseCommand("delete", "Remove a secret"),
	}
}

func (c *DeleteCommand) Execute(args []string, deps Dependencies) error {
	if len(args) != 2 {
		return &UsageError{
			Command: "delete",
			Usage:   "veil delete <vault> <name>",
		}
	}

	vault, name := args[0], args[1]

	if err := deps.App.Delete(vault, name); err != nil {
		return err
	}

	return nil
}

func init() {
	Register(NewDeleteCommand())
}
