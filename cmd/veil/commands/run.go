package commands

import (
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"

	"github.com/ossydotpy/veil/cmd/veil/flags"
)

type RunCommand struct {
	BaseCommand
}

func NewRunCommand() *RunCommand {
	return &RunCommand{
		BaseCommand: NewBaseCommand("run", "Run a command with vault secrets injected into its environment"),
	}
}

func (c *RunCommand) Execute(args []string, deps Dependencies) error {
	sepIdx := findSeparator(args)
	if sepIdx < 0 {
		return &UsageError{
			Command: "run",
			Usage:   "veil run <vault> [flags] -- <command> [args...]",
		}
	}

	veilArgs := args[:sepIdx]
	cmdArgs := args[sepIdx+1:]

	if len(veilArgs) < 1 {
		return &UsageError{
			Command: "run",
			Usage:   "veil run <vault> [flags] -- <command> [args...]",
		}
	}

	if len(cmdArgs) < 1 {
		return &UsageError{
			Command: "run",
			Usage:   "veil run <vault> [flags] -- <command> [args...]",
		}
	}

	vault := veilArgs[0]
	opts, err := flags.ParseRunFlags(veilArgs[1:])
	if err != nil {
		return err
	}

	stderr := deps.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if opts.ShowHelp {
		c.printHelp(stderr)
		return nil
	}

	secrets, err := deps.App.RunEnv(vault, opts.Include, opts.Exclude)
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		fmt.Fprintf(stderr, "Warning: vault %q contains no secrets\n", vault)
		fmt.Fprintf(stderr, "Running command with current environment (no secrets injected)\n")
	}

	mergedEnv := mergeEnv(os.Environ(), secrets)

	binary, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return fmt.Errorf("command not found: %s", cmdArgs[0])
	}

	return syscall.Exec(binary, cmdArgs, mergedEnv)
}

func findSeparator(args []string) int {
	for i, arg := range args {
		if arg == "--" {
			return i
		}
	}
	return -1
}

func mergeEnv(current []string, secrets map[string]string) []string {
	envMap := make(map[string]string, len(current)+len(secrets))
	for _, entry := range current {
		key, val, _ := strings.Cut(entry, "=")
		envMap[key] = val
	}
	maps.Copy(envMap, secrets)
	keys := make([]string, 0, len(envMap))
	for key := range envMap {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	result := make([]string, 0, len(envMap))
	for _, key := range keys {
		result = append(result, fmt.Sprintf("%s=%s", key, envMap[key]))
	}
	return result
}

func (c *RunCommand) printHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: veil run <vault> [flags] -- <command> [args...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run a command with vault secrets injected into its environment.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "The -- separator is required. It tells veil where flags end")
	fmt.Fprintln(w, "and the command begins.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --include <pattern>  Include only matching keys (repeatable)")
	fmt.Fprintln(w, "  --exclude <pattern>  Exclude matching keys (repeatable)")
	fmt.Fprintln(w, "  --help, -h           Show this help message")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  veil run production -- python app.py")
	fmt.Fprintln(w, "  veil run staging --include 'DB_*' -- docker compose up")
	fmt.Fprintln(w, "  veil run production --exclude 'DEBUG*' -- node server.js")
	fmt.Fprintln(w, "  veil run development -- env")
}

func init() {
	Register(NewRunCommand())
}
