package commands

// SetCommand stores a secret in a vault.
type SetCommand struct {
	BaseCommand
}

func NewSetCommand() *SetCommand {
	return &SetCommand{
		BaseCommand: NewBaseCommand("set", "Store a secret"),
	}
}

func (c *SetCommand) Execute(args []string, deps Dependencies) error {
	if len(args) != 3 {
		return &UsageError{
			Command: "set",
			Usage:   "veil set <vault> <name> <value>",
		}
	}

	vault, name, value := args[0], args[1], args[2]

	if err := deps.App.Set(vault, name, value); err != nil {
		return err
	}

	return nil
}

func init() {
	Register(NewSetCommand())
}
