package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ossydotpy/veil/internal/app"
	"github.com/ossydotpy/veil/internal/config"
	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/store"
	"github.com/ossydotpy/veil/internal/store/sqlite"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		key, err := crypto.GenerateRandomKey()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}
		fmt.Printf(`██╗   ██╗███████╗██╗██╗     
██║   ██║██╔════╝██║██║     
██║   ██║█████╗  ██║██║     
╚██╗ ██╔╝██╔══╝  ██║██║     
 ╚████╔╝ ███████╗██║███████╗
  ╚═══╝  ╚══════╝╚═╝╚══════╝
`)
		fmt.Printf("Your new MASTER_KEY is:\n\n%s\n\nSAVE THIS KEY! If you lose it, your secrets are gone forever.\n", key)
		fmt.Println("Export it to your environment:\nexport MASTER_KEY=" + key)
		return
	}

	cfg := config.LoadConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if err := cfg.ValidateMasterKey(); err != nil {
		fmt.Println("Error: " + err.Error())
		fmt.Println("Run 'veil init' to generate a new key if you don't have one.")
		os.Exit(1)
	}

	var s store.Store
	var err error

	switch cfg.StoreType {
	case "sqlite":
		s, err = sqlite.NewSqliteStore(cfg.DbPath)
	default:
		log.Fatalf("Unsupported store type: %s", cfg.StoreType)
	}

	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer s.Close()

	engine, err := crypto.NewEngine(cfg.MasterKey)
	if err != nil {
		log.Fatalf("Failed to initialize crypto: %v", err)
	}

	_ = app.New(s, engine)
}
