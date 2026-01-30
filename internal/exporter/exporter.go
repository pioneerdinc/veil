package exporter

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

var formats = map[string]Exporter{
	"env": &EnvExporter{},
}

func Register(name string, e Exporter) {
	formats[name] = e
}

func Get(name string) Exporter {
	if e, ok := formats[name]; ok {
		return e
	}
	return formats["env"]
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

func safeWriteFile(path string, content []byte, perm os.FileMode, backup bool, backupDir string) error {
	if backup && fileExists(path) {
		backupPath, err := generateBackupPath(path, backupDir)
		if err != nil {
			return fmt.Errorf("failed to generate backup path: %w", err)
		}
		if err := copyFile(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, content, perm); err != nil {
		return err
	}

	if err := os.Chmod(tmpFile, perm); err != nil {
		os.Remove(tmpFile)
		return err
	}

	return os.Rename(tmpFile, path)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func generateBackupPath(originalPath string, backupDir string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Base(originalPath)
	backupName := fmt.Sprintf("%s.backup.%s", filename, timestamp)

	if backupDir != "" {
		if err := os.MkdirAll(backupDir, 0700); err != nil {
			return "", err
		}
		return filepath.Join(backupDir, backupName), nil
	}

	dir := filepath.Dir(originalPath)
	return filepath.Join(dir, backupName), nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, info.Mode())
}

func FilterSecrets(secrets map[string]string, include []string, exclude []string) map[string]string {
	result := make(map[string]string)

	for k, v := range secrets {
		if matchesFilters(k, include, exclude) {
			result[k] = v
		}
	}

	return result
}

func matchesFilters(key string, include []string, exclude []string) bool {
	if len(include) > 0 {
		included := false
		for _, pattern := range include {
			if matchPattern(key, pattern) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	if len(exclude) > 0 {
		for _, pattern := range exclude {
			if matchPattern(key, pattern) {
				return false
			}
		}
	}

	return true
}

func matchPattern(key, pattern string) bool {
	if pattern == key {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}
	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix
	}
	return false
}

func sortKeys(m map[string]string) []string {
	return slices.Sorted(maps.Keys(m))
}
