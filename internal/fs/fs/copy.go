package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func regularFile(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("Lstat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("unsupported non-regular file %q", path)
	}
	return info, nil
}

// CopyRegularFile exclusively creates destination from a regular source file, preserving its permissions.
func CopyRegularFile(source, destination string) error {
	info, err := regularFile(source)
	if err != nil {
		return fmt.Errorf("regularFile: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Open: %w", err)
	}
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		_ = in.Close()
		return fmt.Errorf("OpenFile: %w", err)
	}
	_, copyErr := io.Copy(out, in)
	closeOutErr := out.Close()
	closeInErr := in.Close()
	if copyErr != nil {
		return fmt.Errorf("Copy: %w", copyErr)
	}
	if closeOutErr != nil {
		return fmt.Errorf("Close: %w", closeOutErr)
	}
	if closeInErr != nil {
		return fmt.Errorf("Close: %w", closeInErr)
	}
	return nil
}
