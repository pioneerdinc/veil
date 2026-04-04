package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func GenerateBackupPath(originalPath string, backupDir string) (string, error) {
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
