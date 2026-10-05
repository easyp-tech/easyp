package gitmodules

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"
)

// Released v0 installers ran git archive with the *.proto pathspec, then
// stripped the first matching legacy root. Git reproduces the pathspec and
// export attributes; filtering checkout files by suffix cannot reproduce them.
func hashMigrationProtoArchive(ctx context.Context, checkout string, files []string) (hash string, resultErr error) {
	roots, err := readMigrationLegacyRoots(checkout, files)
	if err != nil {
		return "", fmt.Errorf("readMigrationLegacyRoots: %w", err)
	}
	directory, err := os.MkdirTemp("", "easyp-migration-archive-")
	if err != nil {
		return "", fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(directory); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("RemoveAll: %w", err))
		}
	}()
	archivePath := filepath.Join(directory, "protos.zip")
	_, err = gitV1(ctx, checkout, "archive", "--format=zip", "--output="+archivePath, "HEAD", "--", "*.proto")
	if err != nil {
		return "", fmt.Errorf("gitV1: %w", err)
	}
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("OpenReader: %w", err)
	}
	defer func() {
		if err := archive.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", err))
		}
	}()
	tracked := make(map[string]bool, len(files))
	for _, name := range files {
		tracked[name] = true
	}
	installed, original := make(map[string]*zip.File), make(map[string]*zip.File)
	directories := make(map[string]bool)
	for _, file := range archive.File {
		name := strings.TrimSuffix(file.Name, "/")
		if !filepath.IsLocal(filepath.FromSlash(name)) || path.Clean(name) != name || strings.Contains(name, "\\") {
			return "", fmt.Errorf("unsupported archive path %q", file.Name)
		}
		// Directory entries keep their trailing slash while roots are stripped,
		// matching the published extract.Archive renamer.
		destination := path.Clean(renameMigrationLegacyFile(file.Name, roots))
		if file.FileInfo().IsDir() {
			for dir := destination; dir != "."; dir = path.Dir(dir) {
				directories[dir] = true
			}
			continue
		}
		if !file.Mode().IsRegular() || !tracked[name] {
			return "", fmt.Errorf("unsupported non-regular or untracked archive file %q", name)
		}
		if previous, exists := installed[destination]; exists {
			return "", fmt.Errorf("legacy file collision at %q between %q and %q", destination, previous.Name, name)
		}
		installed[destination], original[name] = file, file
		for dir := path.Dir(destination); dir != "."; dir = path.Dir(dir) {
			directories[dir] = true
		}
	}
	names := make([]string, 0, len(installed))
	for name := range installed {
		if directories[name] {
			return "", fmt.Errorf("legacy directory/file collision at %q", name)
		}
		names = append(names, name)
	}
	// Native checkouts must retain exactly the archived proto paths and bytes.
	// In particular, export-ignore and export-subst must not silently change
	// the contract after the historical installed-tree hash has been verified.
	for _, name := range files {
		if path.Ext(name) != ".proto" {
			continue
		}
		file, exists := original[name]
		if !exists {
			return "", fmt.Errorf("legacy archive source selection omits %q; manual migration is required", name)
		}
		archived, err := readMigrationArchiveFile(file)
		if err != nil {
			return "", fmt.Errorf("readMigrationArchiveFile: %w", err)
		}
		current, err := os.ReadFile(filepath.Join(checkout, filepath.FromSlash(name)))
		if err != nil {
			return "", fmt.Errorf("ReadFile: %w", err)
		}
		if !bytes.Equal(archived, current) {
			return "", fmt.Errorf("legacy archive source selection changes %q contents; manual migration is required", name)
		}
	}
	hash, err = dirhash.Hash1(names, func(name string) (io.ReadCloser, error) {
		file, err := installed[name].Open()
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

func readMigrationArchiveFile(file *zip.File) (content []byte, resultErr error) {
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("Close: %w", err))
		}
	}()
	content, err = io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	return content, nil
}
