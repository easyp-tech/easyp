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
func ReadLocalDependency(directory, source string) (_ v1.Module, resultErr error) {
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
			err := os.MkdirAll(filepath.Join(stage, filepath.FromSlash(logical)), 0o755)
			if err != nil {
				return fmt.Errorf("MkdirAll: %w", err)
			}
			return nil
		}
		if !metadata {
			return nil
		}
		raw, err := sourceview.ReadLocal(context.Background(), directory, logical)
		if err != nil {
			return fmt.Errorf("ReadLocal: %w", err)
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
