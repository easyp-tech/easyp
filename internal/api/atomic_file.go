package api

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeAtomicFile replaces path only after the complete temporary file closes.
func writeAtomicFile(path string, raw []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("Chmod: %w", err)
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("Write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("Rename: %w", err)
	}
	return nil
}
