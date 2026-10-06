package api

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// selectedPolicyScopes retains explicit empty modules in frozen mode, so
// deleting the last source cannot bypass validation of the current graph.
func selectedPolicyScopes(replacements *policyReplacementSources, relative string, frozen bool) (map[string]breakingScope, error) {
	root := replacements.projectRoot
	scopes, err := discoverBreakingScopes(replacements, relative)
	if err != nil || !frozen {
		return scopes, err
	}
	scan := filepath.Join(root, relative)
	info, err := os.Stat(scan)
	if os.IsNotExist(err) {
		return scopes, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Stat: %w", err)
	}
	directory := scan
	if !info.IsDir() {
		directory = filepath.Dir(scan)
	}
	nearest, err := findV1PolicyModuleDir(root, directory)
	if err != nil {
		return nil, fmt.Errorf("findV1PolicyModuleDir: %w", err)
	}
	add := func(directory string) error {
		key, err := filepath.Rel(root, directory)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		if _, exists := scopes[key]; !exists {
			scopes[key] = breakingScope{}
		}
		return nil
	}
	if nearest != "" {
		if err := add(nearest); err != nil {
			return nil, err
		}
	}
	if info.IsDir() {
		err := workspace.WalkAt(root, scan, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if policySourceExcluded(root, path) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !entry.IsDir() && entry.Name() == v1.ModuleFile {
				return add(filepath.Dir(path))
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkDir: %w", err)
		}
	}
	if nearest == "" && len(scopes) == 0 {
		return nil, fmt.Errorf("frozen policy check requires protobuf.mod for %s", scan)
	}
	return scopes, nil
}
