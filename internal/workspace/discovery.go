// Package workspace finds command context without crossing repository boundaries.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// Boundary finds the nearest Git repository, or the outermost EasyP ancestor
// below the home directory when working in a source tree without Git metadata.
func Boundary(start string) (string, error) {
	start, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("Abs: %w", err)
	}
	root := start
	home, _ := os.UserHomeDir()
	for dir := start; ; dir = filepath.Dir(dir) {
		if dir != start && dir == home {
			break
		}
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("Lstat: %w", err)
		}
		for _, name := range []string{"protobuf.mod", "easyp.yaml", "easyp.gen.yaml"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				root = dir
			} else if !errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("Stat: %w", err)
			}
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}
	return root, nil
}

// FindUp returns the nearest named regular file within boundary, or an empty path.
func FindUp(start, boundary, name string) (string, error) {
	for dir := filepath.Clean(start); ; dir = filepath.Dir(dir) {
		rel, err := filepath.Rel(boundary, dir)
		if err != nil || !filepath.IsLocal(rel) {
			return "", fmt.Errorf("directory %q is outside workspace %q", dir, boundary)
		}
		path := filepath.Join(dir, name)
		relative, _ := filepath.Rel(boundary, path)
		resolved, err := sourceview.ResolveLocal(context.Background(), boundary, relative)
		info := resolved.Info
		if err == nil {
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("configuration %q is not a regular file", path)
			}
			return path, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("Stat: %w", err)
		}
		if dir == boundary || dir == filepath.Dir(dir) {
			return "", nil
		}
	}
}

// SkipDirectory excludes storage, vendored sources, and nested repositories.
func SkipDirectory(root, dir string) bool {
	if dir == root {
		return false
	}
	name := filepath.Base(dir)
	if strings.HasPrefix(name, ".") || name == "easyp_vendor" || name == "node_modules" {
		return true
	}
	_, err := os.Lstat(filepath.Join(dir, ".git"))
	return err == nil
}

// Policy finds an ancestor policy or one unambiguous outermost policy below start.
// Descendant policy discovery never executes generation plugins.
func Policy(start string) (string, error) {
	boundary, err := Boundary(start)
	if err != nil {
		return "", fmt.Errorf("Boundary: %w", err)
	}
	path, err := FindUp(start, boundary, "easyp.yaml")
	if err != nil || path != "" {
		return path, err
	}
	var found []string
	err = Walk(start, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if SkipDirectory(start, path) {
			return filepath.SkipDir
		}
		candidate := filepath.Join(path, "easyp.yaml")
		relative, _ := filepath.Rel(boundary, candidate)
		resolved, err := sourceview.ResolveLocal(context.Background(), boundary, relative)
		if info := resolved.Info; err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("configuration %q is not a regular file", candidate)
			}
			found = append(found, candidate)
			return filepath.SkipDir
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("WalkDir: %w", err)
	}
	if len(found) == 1 {
		return found[0], nil
	}
	if len(found) > 1 {
		return "", fmt.Errorf("multiple independent policies found: %v; select one with --cfg", found)
	}
	return "", fmt.Errorf("no easyp.yaml found within %s", boundary)
}

// Module finds the nearest protobuf.mod without leaving the workspace.
func Module(start string) (string, error) {
	boundary, err := Boundary(start)
	if err != nil {
		return "", fmt.Errorf("Boundary: %w", err)
	}
	path, err := FindUp(start, boundary, "protobuf.mod")
	if err != nil {
		return "", fmt.Errorf("FindUp: %w", err)
	}
	if path == "" {
		return "", fmt.Errorf("no protobuf.mod found within %s", boundary)
	}
	return filepath.Dir(path), nil
}
