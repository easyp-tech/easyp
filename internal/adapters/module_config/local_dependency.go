package moduleconfig

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/sourceview"
)

// ReadLocalDependency adapts local metadata through a bounded logical view.
// The regular temporary metadata tree keeps the Git compatibility reader's
// strict historical checkout policy independent of local alias support.
func ReadLocalDependency(directory, source string) (v1.Module, error) {
	return readLocalDependency(directory, source, func(name string) ([]byte, error) {
		return sourceview.ReadLocal(context.Background(), directory, name)
	}, false)
}

// ReadLocalDependencyWithReader adapts captured metadata bytes while retaining
// the same bounded directory selection and regular temporary metadata tree.
func ReadLocalDependencyWithReader(directory, source string, readFile func(string) ([]byte, error)) (v1.Module, error) {
	return readLocalDependency(directory, source, readFile, true)
}

func readLocalDependency(directory, source string, readFile func(string) ([]byte, error), observeDirectories bool) (_ v1.Module, resultErr error) {
	stage, err := os.MkdirTemp("", "easyp-local-metadata-*")
	if err != nil {
		return v1.Module{}, fmt.Errorf("MkdirTemp: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, os.RemoveAll(stage)) }()
	err = sourceview.WalkLocal(context.Background(), directory, ".", func(logical string, resolved sourceview.Resolution, walkErr error) error {
		metadata := IsGitDependencyConfigFile(filepath.Base(logical))
		if walkErr != nil {
			if metadata || logical == "." {
				return walkErr
			}
			return nil
		}
		if resolved.Info.IsDir() {
			if logical != "." && (filepath.Base(logical) == ".git" || filepath.Base(logical) == "easyp_vendor" || filepath.Base(logical) == "node_modules") {
				return fs.SkipDir
			}
			if observeDirectories {
				// Mode selection also depends on absent metadata, including Buf
				// workspace members. Capture that absence before parsing the tree.
				for _, name := range []string{dependencyManifestFile, bufModuleConfigFile, bufWorkConfigFile, bufLockFile, legacyEasyPConfigFile} {
					_, err := readFile(filepath.Join(logical, name))
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("readFile: %w", err)
					}
				}
			}
			err := os.MkdirAll(filepath.Join(stage, filepath.FromSlash(logical)), 0o755)
			if err != nil {
				return fmt.Errorf("MkdirAll: %w", err)
			}
			return nil
		}
		if !metadata {
			return nil
		}
		raw, err := readFile(logical)
		if err != nil {
			return fmt.Errorf("readFile: %w", err)
		}
		err = os.WriteFile(filepath.Join(stage, filepath.FromSlash(logical)), raw, resolved.Info.Mode().Perm())
		if err != nil {
			return fmt.Errorf("WriteFile: %w", err)
		}
		return nil
	})
	if err != nil {
		return v1.Module{}, fmt.Errorf("WalkLocal: %w", err)
	}
	module, err := ReadGitDependency(stage, source)
	if err != nil {
		return v1.Module{}, fmt.Errorf("ReadGitDependency: %w", err)
	}
	return module, nil
}
