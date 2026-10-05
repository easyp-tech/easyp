package fs

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteAtomicFile durably replaces path after syncing the complete temporary file.
// Existing regular-file permissions are preserved; mode is used only for new files.
func WriteAtomicFile(path string, raw []byte, mode os.FileMode) error {
	parent := filepath.Dir(path)
	finalMode := mode
	if info, err := os.Lstat(path); err == nil {
		if info.Mode().IsRegular() {
			finalMode = info.Mode().Perm()
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Lstat: %w", err)
	}

	tmp, err := os.CreateTemp(parent, "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(finalMode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("Chmod: %w", err)
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("Write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("Sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("Rename: %w", err)
	}
	if err := syncDirectory(parent); err != nil {
		return fmt.Errorf("sync parent directory: %w", err)
	}
	return nil
}
