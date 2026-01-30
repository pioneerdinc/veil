package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/exporter"
	"github.com/ossydotpy/veil/internal/store"
	"github.com/ossydotpy/veil/internal/store/sqlite"
)

const logo = `██╗   ██╗███████╗██╗██╗     
██║   ██║██╔════╝██║██║     
██║   ██║█████╗  ██║██║     
╚██╗ ██╔╝██╔══╝  ██║██║     
 ╚████╔╝ ███████╗██║███████╗
  ╚═══╝  ╚══════╝╚═╝╚══════╝`

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	command := os.Args[1]

	if command == "init" {
		cfg := config.LoadConfig()
		if _, err := os.Stat(cfg.DbPath); err == nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: A database already exists at %s\n", cfg.DbPath)
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
	fmt.Fprintln(w, "  reset                       Delete all secrets and start fresh")
	fmt.Fprintln(w, "  set <vault> <name> <value>  Store a secret")
	fmt.Fprintln(w, "  get <vault> <name>          Retrieve a secret")
	fmt.Fprintln(w, "  delete <vault> <name>       Remove a secret")
	fmt.Fprintln(w, "  list <vault>                List all secret names in a vault")
	fmt.Fprintln(w, "  vaults                      List all vaults")
	fmt.Fprintln(w, "  export <vault>              Export vault secrets to .env file")
	fmt.Fprintln(w, "                              --to <path>     Output file path (default: .env)")
	fmt.Fprintln(w, "                              --force         Overwrite existing file")
	fmt.Fprintln(w, "                              --append        Append to existing file")
	fmt.Fprintln(w, "                              --dry-run       Preview without writing")
	fmt.Fprintln(w, "                              --backup        Create backup before overwriting")
	fmt.Fprintln(w, "                              --format <fmt>  Output format (env, json, yaml)")
}

func isCryptoError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "cipher: message authentication failed") ||
		strings.HasPrefix(s, "error decrypting")
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
