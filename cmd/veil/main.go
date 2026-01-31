package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/exporter"
	"github.com/ossydotpy/veil/internal/generator"
	"github.com/ossydotpy/veil/internal/quick"
	"github.com/ossydotpy/veil/internal/store"
	"github.com/ossydotpy/veil/internal/store/sqlite"
)

const logo = `██╗   ██╗███████╗██╗██╗     
██║   ██║██╔════╝██║██║     
██║   ██║█████╗  ██║██║     
╚██╗ ██╔╝██╔══╝  ██║██║     
 ╚████╔╝ ███████╗██║███████╗
  ╚═══╝  ╚══════╝╚═╝╚══════╝`

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	command := os.Args[1]

	if command == "init" {
		cfg := config.LoadConfig()
		if _, err := os.Stat(cfg.DbPath); err == nil {
			fmt.Fprintf(os.Stderr, "Warning: A database already exists at %s\n", cfg.DbPath)
			fmt.Fprintf(os.Stderr, "Generating a new key and using it will make all existing secrets UNREADABLE.\n\n")
		}

		key, err := crypto.GenerateRandomKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to generate key: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(logo)
		fmt.Printf("\nYour new MASTER_KEY is:\n\n%s\n\nSAVE THIS KEY! If you lose it, your secrets are gone forever.\n", key)
		fmt.Println("Export it to your environment:\nexport MASTER_KEY=" + key)
		return
	}

	// Handle quick command early - it doesn't need master key or database
	if command == "quick" {
		runQuickCommand(os.Args[2:])
		return
	}

	if command == "--help" || command == "-h" || command == "help" {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	if command == "--version" || command == "-v" || command == "version" {
		fmt.Printf("veil version %s\n", version)
		os.Exit(0)
	}

	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	var s store.Store
	var err error

	switch cfg.StoreType {
	case "sqlite":
		s, err = sqlite.NewSqliteStore(cfg.DbPath)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unsupported store type: %s\n", cfg.StoreType)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	if command == "reset" {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: This will permanently DELETE ALL SECRETS in the database.\n")
		fmt.Fprintf(os.Stderr, "Are you sure? (type 'yes' to confirm): ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if confirmation != "yes" {
			fmt.Println("Aborted.")
			return
		}

		if err := s.Nuke(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database wiped successfully. You can now run 'veil init' to start over.")
		return
	}

	if err := cfg.ValidateMasterKey(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println("Run 'veil init' to generate a new key if you don't have one.")
		os.Exit(1)
	}

	engine, err := crypto.NewEngine(cfg.MasterKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize crypto: %v\n", err)
		os.Exit(1)
	}

	v := app.New(s, engine)

	switch command {
	case "set":
		if len(os.Args) != 5 {
			fmt.Fprintf(os.Stderr, "Usage: veil set <vault> <name> <value>\n")
			os.Exit(1)
		}
		if err := v.Set(os.Args[2], os.Args[3], os.Args[4]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "get":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "Usage: veil get <vault> <name>\n")
			os.Exit(1)
		}
		val, err := v.Get(os.Args[2], os.Args[3])
		if err != nil {
			if err == store.ErrNotFound {
				fmt.Fprintf(os.Stderr, "Error: secret not found\n")
			} else if isCryptoError(err) {
				fmt.Fprintf(os.Stderr, "Error: decryption failed (check your MASTER_KEY)\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
		fmt.Println(val)
	case "delete":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "Usage: veil delete <vault> <name>\n")
			os.Exit(1)
		}
		if err := v.Delete(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "Usage: veil list <vault>\n")
			os.Exit(1)
		}
		for name, err := range v.List(os.Args[2]) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(name)
		}
	case "vaults":
		for vault, err := range v.ListVaults() {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(vault)
		}
	case "export":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: veil export <vault> [--to <path>] [--force] [--append] [--dry-run]\n")
			os.Exit(1)
		}

		vault := os.Args[2]
		opts := parseExportFlags(os.Args[3:])

		preview, err := v.Export(vault, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if opts.DryRun {
			printPreview(preview, opts.TargetPath)
		} else {
			printExportResult(preview, opts, vault)
		}
	case "search":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: veil search <pattern>\n")
			os.Exit(1)
		}

		pattern := os.Args[2]
		results, err := v.Search(pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No matches found")
			return
		}

		fmt.Printf("Found %d match%s:\n", len(results), plural(len(results)))
		for _, ref := range results {
			fmt.Printf("  %s/%s\n", ref.Vault, ref.Name)
		}
	case "generate":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: veil generate <vault> <name> [--length N] [--no-symbols]\n")
			os.Exit(1)
		}

		opts := parseGenerateFlags(os.Args[4:])
		secret, err := v.Generate(os.Args[2], os.Args[3], opts)
		if err != nil {
			// Check if it's a warning about existing key in .env
			if errors.Is(err, app.ErrKeyExistsInEnv) {
				fmt.Printf("Generated secret: %s\n", secret)
				fmt.Printf("Stored in %s/%s\n", os.Args[2], os.Args[3])
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Generated secret: %s\n", secret)
			fmt.Printf("Stored in %s/%s\n", os.Args[2], os.Args[3])
			if opts.ToEnv != "" {
				if opts.Force {
					fmt.Printf("Updated in %s\n", opts.ToEnv)
				} else {
					fmt.Printf("Appended to %s\n", opts.ToEnv)
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command: %s\n", command)
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w *os.File) {
	fmt.Fprintln(w, logo)
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

func isCryptoError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, crypto.ErrDecryptionFailed) ||
		errors.Is(err, crypto.ErrCiphertextTooShort)
}

func parseExportFlags(args []string) exporter.ExportOptions {
	opts := exporter.ExportOptions{
		TargetPath: ".env",
		Format:     "env",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--to":
			if i+1 < len(args) {
				opts.TargetPath = args[i+1]
				i++
			}
		case "--force":
			opts.Force = true
		case "--append":
			opts.Append = true
		case "--dry-run":
			opts.DryRun = true
		case "--backup":
			opts.Backup = true
		case "--format":
			if i+1 < len(args) {
				opts.Format = args[i+1]
				i++
			}
		case "--include":
			if i+1 < len(args) {
				opts.Include = append(opts.Include, args[i+1])
				i++
			}
		case "--exclude":
			if i+1 < len(args) {
				opts.Exclude = append(opts.Exclude, args[i+1])
				i++
			}
		}
	}

	return opts
}

func printPreview(preview *exporter.Preview, targetPath string) {
	fmt.Println("DRY RUN - No files will be modified")

	if len(preview.NewKeys) > 0 {
		fmt.Printf("Would write to %s:\n", targetPath)
		for _, key := range preview.NewKeys {
			fmt.Printf("  + %s\n", key)
		}
	}

	if len(preview.UpdatedKeys) > 0 {
		fmt.Printf("\nWould update in %s:\n", targetPath)
		for _, key := range preview.UpdatedKeys {
			fmt.Printf("  ~ %s\n", key)
		}
	}

	if len(preview.SkippedKeys) > 0 {
		fmt.Printf("\nWould skip (already exist):\n")
		for _, key := range preview.SkippedKeys {
			fmt.Printf("  - %s\n", key)
		}
	}

	fmt.Printf("\nSummary: %s\n", preview.Summary())
}

func printExportResult(preview *exporter.Preview, opts exporter.ExportOptions, vault string) {
	if opts.Append {
		skipped := len(preview.SkippedKeys)
		added := len(preview.NewKeys) + len(preview.UpdatedKeys)

		if added == 0 {
			fmt.Printf("No changes to %s (all keys already present)\n", opts.TargetPath)
		} else if skipped > 0 {
			fmt.Printf("Appended %d secrets to %s (skipped %d already present)\n", added, opts.TargetPath, skipped)
		} else {
			fmt.Printf("Appended %d secrets to %s\n", added, opts.TargetPath)
		}
	} else {
		count := len(preview.NewKeys) + len(preview.UpdatedKeys)
		fmt.Printf("Exported %d secrets from '%s' to %s\n", count, vault, opts.TargetPath)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "es"
}

func parseGenerateFlags(args []string) generator.Options {
	opts := generator.Options{
		Type: "password",
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				opts.Type = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				opts.Format = args[i+1]
				i++
			}
		case "--prefix":
			if i+1 < len(args) {
				opts.Prefix = args[i+1]
				i++
			}
		case "--length":
			if i+1 < len(args) {
				length, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.Length = length
				}
				i++
			}
		case "--bits":
			if i+1 < len(args) {
				bits, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.Bits = bits
				}
				i++
			}
		case "--no-symbols":
			opts.NoSymbols = true
		case "--to-env":
			if i+1 < len(args) {
				opts.ToEnv = args[i+1]
				i++
			}
		case "--force":
			opts.Force = true
		}
	}

	return opts
}

// runQuickCommand handles ephemeral secret generation without database
func runQuickCommand(args []string) {
	opts := parseQuickFlags(args)

	qg := quick.New()

	// Handle batch mode
	if opts.BatchFile != "" {
		runQuickBatch(qg, opts)
		return
	}

	// Handle count > 1
	if opts.Count > 1 {
		results, err := qg.GenerateMultiple(opts.QuickOptions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Generated %d %ss:\n", len(results), opts.QuickOptions.Type)
		for i, result := range results {
			fmt.Printf("%d. %s\n", i+1, result.Value)
		}
		return
	}

	// Single generation
	result, err := qg.Generate(opts.QuickOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle file append
	if opts.QuickOptions.ToFile != "" {
		if err := quick.AppendToEnvFile(opts.QuickOptions.ToFile, opts.QuickOptions.EnvName, result.Value, opts.QuickOptions.Force); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated: %s\n", result.Value)
		if opts.QuickOptions.Force {
			fmt.Printf("Updated %s in %s\n", opts.QuickOptions.EnvName, opts.QuickOptions.ToFile)
		} else {
			fmt.Printf("Appended %s to %s\n", opts.QuickOptions.EnvName, opts.QuickOptions.ToFile)
		}
		return
	}

	// Display to terminal with template
	output := quick.FormatOutput(result, opts.QuickOptions.Template)
	fmt.Println(output)
}

func runQuickBatch(qg *quick.Generator, opts quickCommandOptions) {
	// Load batch config
	data, err := os.ReadFile(opts.BatchFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read batch file: %v\n", err)
		os.Exit(1)
	}

	var config quick.BatchConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse batch file: %v\n", err)
		os.Exit(1)
	}

	results, err := qg.GenerateBatch(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle file append
	if opts.QuickOptions.ToFile != "" {
		if err := quick.AppendBatchToEnvFile(opts.QuickOptions.ToFile, results, opts.QuickOptions.Force); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Batch: %s\n", opts.BatchFile)
		fmt.Printf("Generated %d secrets appended to %s:\n", len(results), opts.QuickOptions.ToFile)
		for _, r := range results {
			fmt.Printf("  - %s\n", r.EnvName)
		}
		return
	}

	// Output to terminal
	fmt.Printf("Batch: %s\n", opts.BatchFile)
	fmt.Printf("Generated %d secrets:\n", len(results))
	for _, r := range results {
		fmt.Printf("  %s: %s\n", r.EnvName, r.Value)
	}
}

type quickCommandOptions struct {
	QuickOptions quick.Options
	BatchFile    string
	Count        int
}

func parseQuickFlags(args []string) quickCommandOptions {
	opts := quickCommandOptions{
		QuickOptions: quick.Options{
			Type: "password",
		},
	}

	i := 0

	// First non-flag argument is the type (optional)
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		opts.QuickOptions.Type = args[0]

		// Handle shorthand types (hex, base64, uuid -> apikey with format)
		if format := quick.GetFormatFromType(args[0]); format != "" {
			opts.QuickOptions.Type = "apikey"
			opts.QuickOptions.Format = format
		}

		i = 1
	}

	for ; i < len(args); i++ {
		switch args[i] {
		case "--length":
			if i+1 < len(args) {
				length, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.QuickOptions.Length = length
				}
				i++
			}
		case "--format":
			if i+1 < len(args) {
				opts.QuickOptions.Format = args[i+1]
				i++
			}
		case "--prefix":
			if i+1 < len(args) {
				opts.QuickOptions.Prefix = args[i+1]
				i++
			}
		case "--bits":
			if i+1 < len(args) {
				bits, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.QuickOptions.Bits = bits
				}
				i++
			}
		case "--no-symbols":
			opts.QuickOptions.NoSymbols = true
		case "--count":
			if i+1 < len(args) {
				count, err := strconv.Atoi(args[i+1])
				if err == nil {
					opts.Count = count
				}
				i++
			}
		case "--to":
			if i+1 < len(args) {
				opts.QuickOptions.ToFile = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				opts.QuickOptions.EnvName = args[i+1]
				i++
			}
		case "--force":
			opts.QuickOptions.Force = true
		case "--template":
			if i+1 < len(args) {
				opts.QuickOptions.Template = args[i+1]
				i++
			}
		case "--batch":
			if i+1 < len(args) {
				opts.BatchFile = args[i+1]
				i++
			}
		}
	}

	return opts
}
