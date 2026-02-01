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
	copy := make(map[string]Command, len(registry))
	for k, v := range registry {
		copy[k] = v
	}
	return copy
}
