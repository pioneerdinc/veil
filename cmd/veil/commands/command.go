package commands

import (
	"io"

	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/store"
)

// Command represents a CLI command that can be executed.
type Command interface {
	// Name returns the command name as used on the command line.
	Name() string

	// Usage returns a short description of the command for help text.
	Usage() string

	// Execute runs the command with the given arguments and dependencies.
	Execute(args []string, deps Dependencies) error

	// NeedsDeps returns true if the command requires initialized dependencies
	// (store, crypto engine, app). Commands like 'init', 'version', and 'quick'
	// can run without these.
	NeedsDeps() bool

	// NeedsMasterKey returns true if the command requires the master key.
	// Some commands like 'reset' need the store but not the master key.
	NeedsMasterKey() bool
}

// Dependencies holds all injectable dependencies for commands.
type Dependencies struct {
	Store  store.Store
	Engine *crypto.Engine
	Config *config.Config
	App    *app.App

	// IO dependencies for testability
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// BaseCommand provides default implementations for the Command interface.
// Embed this in command structs to get sensible defaults.
type BaseCommand struct {
	name  string
	usage string
}

func NewBaseCommand(name, usage string) BaseCommand {
	return BaseCommand{name: name, usage: usage}
}

func (b BaseCommand) Name() string  { return b.name }
func (b BaseCommand) Usage() string { return b.usage }

// Default: most commands need full dependencies
func (b BaseCommand) NeedsDeps() bool      { return true }
func (b BaseCommand) NeedsMasterKey() bool { return true }
