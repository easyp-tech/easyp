package api

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
)

// checkV1Policies runs the existing checker for each selected breaking policy.
// The checker needs the whole scan to compare package members across files;
// issues are assigned to the nearest policy only after that comparison.
func (b BreakingCheck) checkV1Policies(ctx *cli.Context, log logger.Logger, configPath, projectRoot, scanRoot string) ([]core.IssueInfo, error) {
	path := ctx.String(flagLintDirectoryPath.Name)
	sources, err := discoverV1BreakingPolicySources(filepath.Join(scanRoot, path), projectRoot, configPath)
	if err != nil {
		return nil, fmt.Errorf("discoverV1BreakingPolicySources: %w", err)
	}
	moduleDir, err := findV1PolicyModuleDir(projectRoot, scanRoot)
	if err != nil {
		return nil, fmt.Errorf("findV1PolicyModuleDir: %w", err)
	}
	var importRoots []string
	if moduleDir != "" {
		importRoots, err = resolveV1PolicyImportRoots(ctx.Context, log, moduleDir)
		if err != nil {
			return nil, fmt.Errorf("resolveV1PolicyImportRoots: %w", err)
		}
	}
	var issues []core.IssueInfo
	for _, source := range sources {
		policy, _, err := resolveV1BreakingPolicy(filepath.Dir(source), projectRoot, configPath)
		if err != nil {
			return nil, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
		}
		breakingConfig, err := policy.BreakingConfig(ctx.String(flagAgainstBranchName.Name))
		if err != nil {
			return nil, fmt.Errorf("BreakingConfig: %w", err)
		}
		app, err := buildCore(log, config.Config{BreakingCheck: breakingConfig})
		if err != nil {
			return nil, fmt.Errorf("buildCore: %w", err)
		}
		app.SetImportRoots(importRoots)
		found, err := app.BreakingCheck(ctx.Context, projectRoot, scanRoot, path)
		if err != nil {
			return nil, fmt.Errorf("BreakingCheck for %s: %w", source, err)
		}
		for _, issue := range found {
			_, owner, err := resolveV1BreakingPolicy(filepath.Dir(filepath.Join(scanRoot, issue.Path)), projectRoot, configPath)
			if err != nil {
				return nil, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
			}
			if owner == source {
				issues = append(issues, issue)
			}
		}
	}
	return issues, nil
}

func discoverV1BreakingPolicySources(scanPath, projectRoot, configPath string) ([]string, error) {
	sources := map[string]struct{}{}
	_, owner, err := resolveV1BreakingPolicy(scanPath, projectRoot, configPath)
	if err != nil {
		return nil, err
	}
	sources[owner] = struct{}{}
	err = filepath.WalkDir(scanPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != v1.PolicyFile || path == configPath {
			return nil
		}
		_, source, err := resolveV1BreakingPolicy(filepath.Dir(path), projectRoot, configPath)
		if err != nil {
			return err
		}
		sources[source] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WalkDir: %w", err)
	}
	result := make([]string, 0, len(sources))
	for source := range sources {
		result = append(result, source)
	}
	slices.Sort(result)
	return result, nil
}

func resolveV1BreakingPolicy(directory, projectRoot, configPath string) (v1.Policy, string, error) {
	for _, dir := range ancestorDirs(directory, projectRoot) {
		path := v1PolicyPath(dir, projectRoot, configPath)
		file, found, err := readV1PolicyFile(path)
		if err != nil {
			return v1.Policy{}, "", fmt.Errorf("readV1PolicyFile: %w", err)
		}
		if found && (file.has("breaking") || dir == projectRoot) {
			return file.policy, path, nil
		}
	}
	return v1.Policy{}, "", fmt.Errorf("no easyp.yaml policy for %s", directory)
}
