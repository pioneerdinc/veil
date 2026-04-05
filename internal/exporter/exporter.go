package exporter

import (
	"errors"
	"fmt"
)

var formats = map[string]Exporter{
	"env": &EnvExporter{},
}

var ErrUnsupportedFormat = errors.New("unsupported export format")

func Register(name string, e Exporter) {
	formats[name] = e
}

func Get(name string) (Exporter, error) {
	if e, ok := formats[name]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, name)
}

type Exporter interface {
	Export(secrets map[string]string, opts ExportOptions) error
	Preview(secrets map[string]string, opts ExportOptions) (*Preview, error)
	Format() string
}

type ExportOptions struct {
	TargetPath string
	Append     bool
	Force      bool
	Backup     bool
	BackupDir  string
	Include    []string
	Exclude    []string
	DryRun     bool
	Format     string
}

type Preview struct {
	NewKeys     []string
	UpdatedKeys []string
	SkippedKeys []string
	Content     string
}

func (p *Preview) Summary() string {
	return fmt.Sprintf("%d new, %d updates, %d skipped", len(p.NewKeys), len(p.UpdatedKeys), len(p.SkippedKeys))
}
