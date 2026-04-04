package fsutil

import (
	"fmt"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func SafeWriteFile(path string, content []byte, perm os.FileMode, backup bool, backupDir string) error {
	if backup && FileExists(path) {
		backupPath, err := GenerateBackupPath(path, backupDir)
		if err != nil {
			return fmt.Errorf("failed to generate backup path: %w", err)
		}
		if err := CopyFile(path, backupPath); err != nil {
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

func CopyFile(src, dst string) error {
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
