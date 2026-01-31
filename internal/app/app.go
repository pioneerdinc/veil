package app

import (
	"bufio"
	"bytes"
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/exporter"
	"github.com/ossydotpy/veil/internal/generator"
	"github.com/ossydotpy/veil/internal/store"
)

type App struct {
	store  store.Store
	crypto *crypto.Engine
}

func New(s store.Store, c *crypto.Engine) *App {
	return &App{
		store:  s,
		crypto: c,
	}
}

func (a *App) Set(vault, name, value string) error {
	encrypted, err := a.crypto.Encrypt(value)
	if err != nil {
		return err
	}
	return a.store.Save(vault, name, encrypted)
}

func (a *App) Get(vault, name string) (string, error) {
	encrypted, err := a.store.Get(vault, name)
	if err != nil {
		return "", err
	}
	return a.crypto.Decrypt(encrypted)
}

func (a *App) Delete(vault, name string) error {
	return a.store.Delete(vault, name)
}

func (a *App) List(vault string) iter.Seq2[string, error] {
	return a.store.List(vault)
}

func (a *App) ListVaults() iter.Seq2[string, error] {
	return a.store.ListVaults()
}

func (a *App) Reset() error {
	return a.store.Nuke()
}

func (a *App) Search(pattern string) ([]store.SecretRef, error) {
	var results []store.SecretRef
	for ref, err := range a.store.Search(pattern) {
		if err != nil {
			return nil, err
		}
		results = append(results, ref)
	}
	return results, nil
}

func (a *App) GetAllSecrets(vault string) (map[string]string, error) {
	secrets := make(map[string]string)
	for name, err := range a.List(vault) {
		if err != nil {
			return nil, err
		}
		value, err := a.Get(vault, name)
		if err != nil {
			return nil, err
		}
		secrets[name] = value
	}

	return secrets, nil
}

func (a *App) Export(vault string, opts exporter.ExportOptions) (*exporter.Preview, error) {
	secrets, err := a.GetAllSecrets(vault)
	if err != nil {
		return nil, err
	}

	filtered := exporter.FilterSecrets(secrets, opts.Include, opts.Exclude)

	exp := exporter.Get(opts.Format)

	preview, err := exp.Preview(filtered, opts)
	if err != nil {
		return nil, err
	}

	if !opts.DryRun {
		if err := exp.Export(filtered, opts); err != nil {
			return nil, err
		}
	}

	return preview, nil
}

func (a *App) Generate(vault, name string, opts generator.Options) (string, error) {
	// Generate the secret
	secret, err := generator.Generate(opts)
	if err != nil {
		return "", err
	}

	// Store in vault
	if err := a.Set(vault, name, secret); err != nil {
		return "", err
	}

	if opts.ToEnv != "" {
		if err := a.appendToEnvFile(name, secret, opts.ToEnv, opts.Force); err != nil {
			return secret, err
		}
	}

	return secret, nil
}

// appendToEnvFile appends a key-value pair to an .env file
func (a *App) appendToEnvFile(key, value, path string, force bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	keyExists := false
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if existingKey, _, found := strings.Cut(line, "="); found {
			if existingKey == key {
				keyExists = true
				break
			}
		}
	}

	if keyExists && !force {
		return fmt.Errorf("%s already exists in %s, use --force to overwrite", key, path)
	}

	var newContent strings.Builder
	if keyExists && force {
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				newContent.WriteString(line)
				newContent.WriteString("\n")
				continue
			}
			if existingKey, _, found := strings.Cut(trimmed, "="); found {
				if existingKey == key {
					newContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
					continue
				}
			}
			newContent.WriteString(line)
			newContent.WriteString("\n")
		}
	} else {
		newContent.Write(data)
		if !strings.HasSuffix(newContent.String(), "\n") {
			newContent.WriteString("\n")
		}
		newContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	if err := os.WriteFile(path, []byte(newContent.String()), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}
