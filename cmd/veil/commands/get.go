package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/store"
)

// GetCommand retrieves a secret from a vault.
type GetCommand struct {
	BaseCommand
}

func NewGetCommand() *GetCommand {
	return &GetCommand{
		BaseCommand: NewBaseCommand("get", "Retrieve a secret"),
	}
}

func (c *GetCommand) Execute(args []string, deps Dependencies) error {
	if len(args) != 2 {
		return &UsageError{
			Command: "get",
			Usage:   "veil get <vault> <name>",
		}
	}

	stdout := deps.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	vault, name := args[0], args[1]

	val, err := deps.App.Get(vault, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("secret not found")
		}
		if isCryptoError(err) {
			return fmt.Errorf("decryption failed (check your MASTER_KEY)")
		}
		return err
	}

	fmt.Fprintln(stdout, val)
	return nil
}

// isCryptoError checks if an error is related to cryptographic operations.
func isCryptoError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, crypto.ErrDecryptionFailed) ||
		errors.Is(err, crypto.ErrCiphertextTooShort)
}

func init() {
	Register(NewGetCommand())
}
