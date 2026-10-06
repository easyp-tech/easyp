package gitmodules

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/easyp-tech/easyp/internal/adapters/gitindex"
	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// trackedV1Files selects regular Git files, including those outside import roots.
// Symlinks and submodules are omitted, matching Go module archive behavior.
func trackedV1Files(ctx context.Context, checkout string) ([]string, error) {
	entries, err := readV1GitIndex(ctx, checkout)
	if err != nil {
		return nil, fmt.Errorf("readV1GitIndex: %w", err)
	}
	files, err := regularTrackedV1Files(checkout, entries)
	if err != nil {
		return nil, fmt.Errorf("regularTrackedV1Files: %w", err)
	}
	return files, nil
}

func readV1GitIndex(ctx context.Context, checkout string) ([]gitindex.Entry, error) {
	raw, err := gitV1(ctx, checkout, "ls-files", "--stage", "-z")
	if err != nil {
		return nil, fmt.Errorf("gitV1: %w", err)
	}
	entries, err := gitindex.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}
	return entries, nil
}

// Snapshot hashes and installation use the same regular-file selection.
// Policy checks use Git modes even when symlinks are materialized as text.
func regularTrackedV1Files(checkout string, entries []gitindex.Entry) ([]string, error) {
	var files []string
	for _, entry := range entries {
		switch entry.Mode {
		case gitindex.Symlink, gitindex.Gitlink:
			if entry.Mode == gitindex.Symlink && moduleconfig.IsGitDependencyConfigFile(filepath.Base(entry.Path)) {
				return nil, fmt.Errorf("non-regular dependency config %q in Git index", entry.Path)
			}
			continue
		case gitindex.RegularFile, gitindex.ExecutableFile:
		default:
			return nil, fmt.Errorf("unsupported Git file mode %q for %q", entry.Mode, entry.Path)
		}
		if _, err := regularV1File(filepath.Join(checkout, filepath.FromSlash(entry.Path))); err != nil {
			return nil, fmt.Errorf("regularV1File: %w", err)
		}
		files = append(files, entry.Path)
	}
	return files, nil
}

func selectV1ProtoFiles(files []string, filters []v1.ProtoFileFilter) []string {
	if len(filters) == 0 {
		return files
	}
	selected := make([]string, 0, len(files))
	for _, name := range files {
		if filepath.Ext(name) != ".proto" {
			selected = append(selected, name)
			continue
		}
		insideRoot, allowed := false, false
		for _, filter := range filters {
			if !v1FileWithin(name, filter.Root) {
				continue
			}
			insideRoot = true
			included := len(filter.Includes) == 0
			for _, directory := range filter.Includes {
				included = included || v1FileWithin(name, directory)
			}
			for _, directory := range filter.Excludes {
				if v1FileWithin(name, directory) {
					included = false
					break
				}
			}
			if included {
				allowed = true
				break
			}
		}
		if !insideRoot || allowed {
			selected = append(selected, name)
		}
	}
	return selected
}

func v1FileWithin(name, directory string) bool {
	directory = filepath.ToSlash(directory)
	return directory == "." || name == directory || strings.HasPrefix(name, directory+"/")
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
