package importer

import (
	"errors"
	"fmt"
)

var importers = map[string]Importer{
	"env": &EnvImporter{},
}

var ErrUnsupportedFormat = errors.New("unsupported import format")

func Register(name string, i Importer) {
	importers[name] = i
}

func Get(name string) (Importer, error) {
	if i, ok := importers[name]; ok {
		return i, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, name)
}

type Importer interface {
	Import(opts ImportOptions) (map[string]string, error)
	Format() string
}

type ImportOptions struct {
	SourcePath string
	Include    []string
	Exclude    []string
	Format     string
	DryRun     bool
	Force      bool
}

type Preview struct {
	NewKeys     []string
	UpdatedKeys []string
	SkippedKeys []string
}

func (p *Preview) Summary() string {
	return fmt.Sprintf("%d new, %d updates, %d skipped", len(p.NewKeys), len(p.UpdatedKeys), len(p.SkippedKeys))
}
