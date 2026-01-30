package app

import (
	"iter"

	"github.com/ossydotpy/veil/internal/crypto"
	"github.com/ossydotpy/veil/internal/exporter"
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
