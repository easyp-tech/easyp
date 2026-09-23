package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

// trackedV1Files validates the complete checkout, including files outside the
// selected module. Its content hash intentionally covers the entire repository.
func trackedV1Files(ctx context.Context, checkout string) ([]string, error) {
	raw, err := gitV1(ctx, checkout, "ls-files", "-z")
	if err != nil {
		return nil, fmt.Errorf("gitV1: %w", err)
	}
	files := strings.FieldsFunc(raw, func(r rune) bool { return r == 0 })
	for _, name := range files {
		if !filepath.IsLocal(filepath.FromSlash(name)) {
			return nil, fmt.Errorf("invalid tracked file path %q", name)
		}
		if _, err := regularV1File(filepath.Join(checkout, filepath.FromSlash(name))); err != nil {
			return nil, fmt.Errorf("regularV1File: %w", err)
		}
	}
	return files, nil
}

func hashV1Files(root string, files []string) (string, error) {
	hash, err := dirhash.Hash1(files, func(name string) (io.ReadCloser, error) {
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, fmt.Errorf("Open: %w", err)
		}
		return file, nil
	})
	if err != nil {
		return "", fmt.Errorf("Hash1: %w", err)
	}
	return hash, nil
}

func regularV1File(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("Lstat: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("unsupported non-regular file %q", path)
	}
	return info, nil
}

func copyV1RegularFile(source, destination string) error {
	info, err := regularV1File(source)
	if err != nil {
		return fmt.Errorf("regularV1File: %w", err)
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
