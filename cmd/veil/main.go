package main

import (
	"fmt"
	"os"

	"github.com/ossydotpy/veil/cmd/veil/commands"
	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/store"
	"github.com/ossydotpy/veil/internal/store/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		return fmt.Errorf("no command provided")
	}

	cmdName := os.Args[1]

	// Handle help flags
	if cmdName == "--help" || cmdName == "-h" || cmdName == "help" {
		printUsage(os.Stdout)
		return nil
	}

	// Handle version flags (allow both flag and command style)
	if cmdName == "--version" || cmdName == "-v" {
		cmdName = "version"
	}

	// Look up the command in the registry
	cmd, err := commands.Get(cmdName)
	if err != nil {
		printUsage(os.Stderr)
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	// Build dependencies based on what the command needs
	deps, cleanup, err := buildDependencies(cmd)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Execute the command
	args := os.Args[2:]
	if err := cmd.Execute(args, deps); err != nil {
		// Handle UsageError specially - print without "Error:" prefix
		if _, ok := err.(*commands.UsageError); ok {
			return err
		}
		return err
	}

	return nil
}

// buildDependencies creates the dependencies struct based on what the command needs.
func buildDependencies(cmd commands.Command) (commands.Dependencies, func(), error) {
	deps := commands.Dependencies{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	// If the command doesn't need dependencies, return early
	if !cmd.NeedsDeps() {
		return deps, nil, nil
	}

	// Load and validate config
	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		return deps, nil, fmt.Errorf("configuration error: %w", err)
	}
	deps.Config = cfg

	// Initialize store
	var s store.Store
	var err error

	switch cfg.StoreType {
	case "sqlite":
		s, err = sqlite.NewSqliteStore(cfg.DbPath)
	default:
		return deps, nil, fmt.Errorf("unsupported store type: %s", cfg.StoreType)
	}

	if err != nil {
		return deps, nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	deps.Store = s

	cleanup := func() {
		s.Close()
	}

	// If the command doesn't need master key, we're done
	if !cmd.NeedsMasterKey() {
		return deps, cleanup, nil
	}

	// Validate master key and initialize crypto
	if cfg.MasterKey == "" {
		s.Close()
		return commands.Dependencies{}, nil, fmt.Errorf("MASTER_KEY environment variable is not set")
	}
	if err := cfg.ValidateMasterKey(); err != nil {
		s.Close()
		return commands.Dependencies{}, nil, fmt.Errorf("invalid MASTER_KEY: %w (run 'veil init' if you need a new key)", err)
	}

	engine, err := crypto.NewEngine(cfg.MasterKey)
	if err != nil {
		s.Close()
		return deps, nil, fmt.Errorf("failed to initialize crypto: %w", err)
	}
	deps.Engine = engine

	// Initialize app
	deps.App = app.New(s, engine)

	return deps, cleanup, nil
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, commands.Logo)
	fmt.Fprintln(w, "Usage: veil <command> [arguments]")
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  init                        Generate a new master key")
	fmt.Fprintln(w, "  version                     Show version information")
	fmt.Fprintln(w, "  reset                       Delete all secrets and start fresh")
	fmt.Fprintln(w, "  set <vault> <name> <value>  Store a secret")
	fmt.Fprintln(w, "  get <vault> <name>          Retrieve a secret")
	fmt.Fprintln(w, "  delete <vault> <name>       Remove a secret")
	fmt.Fprintln(w, "  list <vault>                List all secret names in a vault")
	fmt.Fprintln(w, "  vaults                      List all vaults")
	fmt.Fprintln(w, "  search <pattern>            Search secrets across all vaults")
	fmt.Fprintln(w, "                              Supports * wildcard (e.g., DB_*)")
	fmt.Fprintln(w, "  generate <vault> <name>     Generate and store a secret")
	fmt.Fprintln(w, "                              --type <type>   Secret type: password|apikey|jwt")
	fmt.Fprintln(w, "                              --length N      Password length (default: 32)")
	fmt.Fprintln(w, "                              --no-symbols    Alphanumeric only")
	fmt.Fprintln(w, "                              --format <fmt>  API key format: uuid|hex|base64")
	fmt.Fprintln(w, "                              --prefix <str>  Prefix for API keys (e.g., sk_live_)")
	fmt.Fprintln(w, "                              --bits N        JWT secret bits: 256|512 (default: 256)")
	fmt.Fprintln(w, "                              --to-env <path> Append to .env file")
	fmt.Fprintln(w, "                              --force         Overwrite existing key in .env")
	fmt.Fprintln(w, "  export <vault>              Export vault secrets to .env file")
	fmt.Fprintln(w, "                              --to <path>     Output file path (default: .env)")
	fmt.Fprintln(w, "                              --force         Overwrite existing file")
	fmt.Fprintln(w, "                              --append        Append to existing file")
	fmt.Fprintln(w, "                              --dry-run       Preview without writing")
	fmt.Fprintln(w, "                              --backup        Create backup before overwriting")
	fmt.Fprintln(w, "                              --format <fmt>  Output format (env, json, yaml)")
	fmt.Fprintln(w, "  quick [type]                Generate ephemeral secret (no storage)")
	fmt.Fprintln(w, "                              Types: password|apikey|jwt|hex|base64|uuid")
	fmt.Fprintln(w, "                              --length N      Password length (default: 32)")
	fmt.Fprintln(w, "                              --no-symbols    Alphanumeric only")
	fmt.Fprintln(w, "                              --format <fmt>  API key format: uuid|hex|base64")
	fmt.Fprintln(w, "                              --prefix <str>  Prefix for generated value")
	fmt.Fprintln(w, "                              --bits N        JWT secret bits (default: 256)")
	fmt.Fprintln(w, "                              --count N       Generate multiple secrets")
	fmt.Fprintln(w, "                              --to <path>     Append to .env file")
	fmt.Fprintln(w, "                              --name <KEY>    Env variable name (required with --to)")
	fmt.Fprintln(w, "                              --force         Overwrite existing key in .env")
	fmt.Fprintln(w, "                              --template <s>  Custom output format (use {value})")
	fmt.Fprintln(w, "                              --batch <file>  Generate from JSON config file")
}
