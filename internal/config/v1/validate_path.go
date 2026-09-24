package v1

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/config"
)

// ValidatePath accepts a single file or discovers all EasyP config files
// below a directory. WalkDir provides a stable, lexical reporting order.
func ValidatePath(path string) ([]config.ValidationIssue, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	if !info.IsDir() {
		return validateNamedFile(path, filepath.Base(path))
	}

	files, err := findConfigFiles(path)
	if err != nil {
		return nil, fmt.Errorf("findConfigFiles: %w", err)
	}
	var issues []config.ValidationIssue
	for _, file := range files {
		relative, err := filepath.Rel(path, file)
		if err != nil {
			return nil, fmt.Errorf("Rel: %w", err)
		}
		fileIssues, err := validateNamedFile(file, filepath.ToSlash(relative))
		if err != nil {
			return nil, fmt.Errorf("validateNamedFile: %w", err)
		}
		issues = append(issues, fileIssues...)
	}
	return issues, nil
}

func findConfigFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(file string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isEasyPConfigFile(entry.Name()) {
			return nil
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no configuration files found under %s", root)
	}
	return files, nil
}

func isEasyPConfigFile(name string) bool {
	switch name {
	case PolicyFile, GenerateFile, ModuleFile, LockFile:
		return true
	default:
		return false
	}
}

func validateNamedFile(path, name string) ([]config.ValidationIssue, error) {
	issues, err := ValidateFile(path)
	if err != nil {
		return nil, fmt.Errorf("ValidateFile: %w", err)
	}
	for i := range issues {
		issues[i].File = name
	}
	return issues, nil
}
