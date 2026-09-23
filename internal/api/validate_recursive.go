package api

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/config"
)

// validateConfigPath accepts a single file or discovers all EasyP config files
// below a directory. WalkDir provides a stable, lexical reporting order.
func validateConfigPath(path string) ([]config.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("config path %s: %w", path, err)
	}
	if !info.IsDir() {
		issues, err := validateConfigFile(path)
		if err != nil {
			return nil, fmt.Errorf("validate %s: %w", path, err)
		}
		for i := range issues {
			issues[i].File = filepath.Base(path)
		}
		return issues, nil
	}

	var issues []config.ValidationIssue
	var found int
	err = filepath.WalkDir(path, func(file string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isEasyPConfigFile(entry.Name()) {
			return nil
		}
		fileIssues, err := validateConfigFile(file)
		if err != nil {
			return fmt.Errorf("validate %s: %w", file, err)
		}
		relative, err := filepath.Rel(path, file)
		if err != nil {
			return err
		}
		for i := range fileIssues {
			fileIssues[i].File = filepath.ToSlash(relative)
		}
		issues = append(issues, fileIssues...)
		found++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk config directory %s: %w", path, err)
	}
	if found == 0 {
		return nil, fmt.Errorf("no configuration files found under %s", path)
	}
	return issues, nil
}

func isEasyPConfigFile(name string) bool {
	switch name {
	case "easyp.yaml", "easyp.gen.yaml", "protobuf.mod", "protobuf.lock":
		return true
	default:
		return false
	}
}
