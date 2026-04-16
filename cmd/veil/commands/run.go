package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ossydotpy/veil/cmd/veil/flags"
)

// RunCommand injects vault secrets into a process's environment.
type RunCommand struct {
	BaseCommand
}

func NewRunCommand() *RunCommand {
	return &RunCommand{
		BaseCommand: NewBaseCommand("run", "Run a command with secrets injected into its environment"),
	}
}

func (c *RunCommand) Execute(args []string, deps Dependencies) error {
	opts, err := flags.ParseRunFlags(args)
	if err != nil {
		return err
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	if opts.ShowHelp || len(args) < 1 {
		c.printHelp(stdout)
		return nil
	}

	if opts.Command == "" {
		return &UsageError{
			Command: "run",
			Usage:   "veil run <vault> -- <command> [args...]",
		}
	}

	// Fetch all secrets for the vault
	secrets, err := deps.App.GetAllSecrets(opts.Vault)
	if err != nil {
		return fmt.Errorf("failed to load vault '%s': %w", opts.Vault, err)
	}

	// Prepare child process
	cmd := exec.Command(opts.Command, opts.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Inject the current environment + secrets
	envMap := make(map[string]string)
	
	// Start with current environment
	for _, envStr := range os.Environ() {
		// We can't guarantee there aren't multiple '=' in the value
		for i := 0; i < len(envStr); i++ {
			if envStr[i] == '=' {
				envMap[envStr[:i]] = envStr[i+1:]
				break
			}
		}
	}

	// Override with vault secrets
	for k, v := range secrets {
		envMap[k] = v
	}

	// Rebuild into "K=V" slice
	var childEnv []string
	for k, v := range envMap {
		childEnv = append(childEnv, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Env = childEnv

	// Execute it
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (c *RunCommand) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: veil run <vault> -- <command> [args...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run a command with vault secrets injected directly into its environment variables.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  veil run production -- python main.py")
	fmt.Fprintln(w, "  veil run production -- npm start")
	fmt.Fprintln(w, "  veil run production -- go run main.go")
	fmt.Fprintln(w, "  veil run production -- printenv")
}

func init() {
	Register(NewRunCommand())
}
