package gitmodules

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	moduleconfig "github.com/easyp-tech/easyp/internal/adapters/module_config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// trackedV1Files selects regular Git files, including those outside import roots.
// Symlinks and submodules are omitted, matching Go module archive behavior.
func trackedV1Files(ctx context.Context, checkout string) ([]string, error) {
	raw, err := gitV1(ctx, checkout, "ls-files", "--stage", "-z")
	if err != nil {
		return nil, fmt.Errorf("gitV1: %w", err)
	}
	var files []string
	for _, entry := range strings.Split(raw, "\x00") {
		if entry == "" {
			continue
		}
		metadata, name, ok := strings.Cut(entry, "\t")
		fields := strings.Fields(metadata)
		if !ok || len(fields) != 3 || fields[2] != "0" {
			return nil, fmt.Errorf("invalid Git index entry %q", entry)
		}
		if !filepath.IsLocal(filepath.FromSlash(name)) {
			return nil, fmt.Errorf("invalid tracked file path %q", name)
		}
		switch fields[0] {
		case "120000", "160000":
			if fields[0] == "120000" && moduleconfig.IsGitDependencyConfigFile(filepath.Base(name)) {
				return nil, fmt.Errorf("non-regular dependency config %q in Git index", name)
			}
			continue
		case "100644", "100755":
		default:
			return nil, fmt.Errorf("unsupported Git file mode %q for %q", fields[0], name)
		}
		if _, err := regularV1File(filepath.Join(checkout, filepath.FromSlash(name))); err != nil {
			return nil, fmt.Errorf("regularV1File: %w", err)
		}
		files = append(files, name)
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
