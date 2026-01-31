package commands

var registry = make(map[string]Command)

func Register(cmd Command) {
	registry[cmd.Name()] = cmd
}

func Get(name string) (Command, error) {
	cmd, ok := registry[name]
	if !ok {
		return nil, ErrUnknownCommand
	}
	return cmd, nil
}

func All() map[string]Command {
	return registry
}
