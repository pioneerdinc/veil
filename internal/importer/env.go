package importer

import (
	"github.com/ossydotpy/veil/internal/envfile"
	"github.com/ossydotpy/veil/internal/filter"
)

type EnvImporter struct{}

func (e *EnvImporter) Format() string {
	return "env"
}

func (e *EnvImporter) Import(opts ImportOptions) (map[string]string, error) {
	secrets, err := envfile.ParseEnvFile(opts.SourcePath)
	if err != nil {
		return nil, err
	}

	filtered := filter.FilterSecrets(secrets, opts.Include, opts.Exclude)
	return filtered, nil
}
