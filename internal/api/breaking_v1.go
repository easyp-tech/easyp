package api

import (
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
)

// checkV1Policies compares each module in an independent dependency context.
// Policies are selected from the current project; each baseline has its own
// sources, manifests and locks, without checking out the caller's worktree.
func (b BreakingCheck) checkV1Policies(ctx *cli.Context, log logger.Logger, configPath, projectRoot, scanRoot string) ([]core.IssueInfo, error) {
	scanPath := filepath.Join(scanRoot, ctx.String(flagLintDirectoryPath.Name))
	sources, err := discoverV1BreakingPolicySources(scanPath, projectRoot, configPath)
	if err != nil {
		return nil, fmt.Errorf("discoverV1BreakingPolicySources: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("moduleCache: %w", err)
	}
	snapshots := make(map[string]*gitadapter.Snapshot)
	defer func() {
		for _, snapshot := range snapshots {
			_ = snapshot.Close()
		}
	}()
	var issues []core.IssueInfo
	seen := make(map[breakingIssueKey]bool)
	for _, source := range sources {
		policy, _, err := resolveV1BreakingPolicy(filepath.Dir(source), projectRoot, configPath)
		if err != nil {
			return nil, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
		}
		cfg, err := resolveV1BreakingConfig(ctx, policy)
		if err != nil {
			return nil, fmt.Errorf("resolveV1BreakingConfig: %w", err)
		}
		snapshot := snapshots[cfg.AgainstGitRef]
		if snapshot == nil {
			snapshot, err = gitadapter.SnapshotRevision(ctx.Context, scanRoot, cfg.AgainstGitRef)
			if err != nil {
				return nil, fmt.Errorf("SnapshotRevision: %w", err)
			}
			snapshots[cfg.AgainstGitRef] = snapshot
		}
		repositoryRoot := snapshot.RepositoryRoot
		ignorePaths, err := breakingIgnorePaths(repositoryRoot, filepath.Dir(source), cfg.Ignore)
		if err != nil {
			return nil, fmt.Errorf("breakingIgnorePaths: %w", err)
		}
		relative, err := filepath.Rel(repositoryRoot, scanPath)
		if err != nil || !filepath.IsLocal(relative) {
			return nil, fmt.Errorf("%w: %s", core.ErrRootOutsideProject, scanPath)
		}
		current, err := discoverBreakingScopes(repositoryRoot, relative)
		if err != nil {
			return nil, fmt.Errorf("discoverBreakingScopes current: %w", err)
		}
		baseline, err := discoverBreakingScopes(snapshot.Root, relative)
		if err != nil {
			return nil, fmt.Errorf("discoverBreakingScopes baseline: %w", err)
		}
		keys := make(map[string]bool)
		for key := range current {
			keys[key] = true
		}
		for key := range baseline {
			keys[key] = true
		}
		for _, module := range slices.Sorted(maps.Keys(keys)) {
			currentFiles, baselineFiles := current[module].files, baseline[module].files
			// Keep the whole selected module in each comparison so FILE moves and
			// package members spanning policy subtrees remain comparable.
			ownFiles := append(append([]string(nil), currentFiles...), baselineFiles...)
			relevant, err := breakingScopeUsesPolicy(repositoryRoot, ownFiles, projectRoot, configPath, source)
			if err != nil {
				return nil, err
			}
			if !relevant {
				continue
			}
			currentImports, err := breakingImportRoots(ctx.Context, cache, repositoryRoot, module, currentFiles, nil)
			if err != nil {
				return nil, fmt.Errorf("current module %s: %w", module, err)
			}
			baselineImports, err := breakingImportRoots(ctx.Context, cache, snapshot.Root, module, baselineFiles, snapshot)
			if err != nil {
				return nil, fmt.Errorf("baseline %s module %s: %w", cfg.AgainstGitRef, module, err)
			}
			app := core.New(core.Options{Logger: log, ImportRoots: currentImports, BreakingCheckConfig: core.BreakingCheckConfig{
				IgnoreDirs: append(append([]string(nil), ignorePaths...), defaultVendorDir), AgainstGitRef: cfg.AgainstGitRef,
				FilesCheck:     slices.Contains(cfg.Use, core.BreakingCheckFilesCheck),
				IgnoreUnstable: cfg.IgnoreUnstable,
			}})
			found, err := app.CompareBreaking(ctx.Context, newBreakingWalker(repositoryRoot, currentFiles), newBreakingWalker(snapshot.Root, baselineFiles), baselineImports)
			if err != nil {
				return nil, fmt.Errorf("CompareBreaking module %s against %s: %w", module, cfg.AgainstGitRef, err)
			}
			for _, issue := range found {
				owned := slices.Contains(ownFiles, filepath.ToSlash(issue.Path))
				if owned {
					directory := filepath.Dir(filepath.Join(repositoryRoot, issue.Path))
					_, owner, err := resolveV1BreakingPolicy(directory, projectRoot, configPath)
					if err != nil {
						return nil, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
					}
					if owner != source {
						continue
					}
				} else {
					// Dependency findings belong to the effective policy of the checked
					// files, not an unrelated ancestor policy at the module directory.
					issue.Path = filepath.ToSlash(filepath.Join(module, issue.Path))
				}
				key := breakingIssueKey{path: issue.Path, rule: issue.RuleName, message: issue.Message, line: issue.Position.Line, column: issue.Position.Column}
				if !seen[key] {
					issues = append(issues, issue)
					seen[key] = true
				}
			}
		}
	}
	return issues, nil
}

func breakingScopeUsesPolicy(repositoryRoot string, files []string, projectRoot, configPath, source string) (bool, error) {
	directories := make([]string, 0, len(files))
	for _, file := range files {
		directories = append(directories, filepath.Dir(filepath.Join(repositoryRoot, file)))
	}
	for _, directory := range directories {
		_, owner, err := resolveV1BreakingPolicy(directory, projectRoot, configPath)
		if err != nil {
			return false, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
		}
		if owner == source {
			return true, nil
		}
	}
	return false, nil
}

func discoverV1BreakingPolicySources(scanPath, projectRoot, configPath string) ([]string, error) {
	sources := map[string]struct{}{}
	policyDirectory := scanPath
	info, statErr := os.Stat(scanPath)
	missing := os.IsNotExist(statErr)
	if statErr != nil && !missing {
		return nil, fmt.Errorf("Stat: %w", statErr)
	}
	if missing || !info.IsDir() {
		policyDirectory = filepath.Dir(scanPath)
	}
	_, owner, err := resolveV1BreakingPolicy(policyDirectory, projectRoot, configPath)
	if err != nil {
		return nil, err
	}
	sources[owner] = struct{}{}
	if missing {
		return []string{owner}, nil
	}
	err = filepath.WalkDir(scanPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if policySourceExcluded(projectRoot, path) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
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

// resolveV1BreakingConfig distinguishes an explicit flag from its default value.
func resolveV1BreakingConfig(ctx *cli.Context, policy v1.Policy) (config.BreakingCheck, error) {
	against := ctx.String(flagAgainstBranchName.Name)
	explicit := ctx.IsSet(flagAgainstBranchName.Name)
	if explicit && strings.TrimSpace(against) == "" {
		return config.BreakingCheck{}, fmt.Errorf("--against must not be empty")
	}
	cfg, err := policy.BreakingConfig(against)
	if err != nil {
		return config.BreakingCheck{}, fmt.Errorf("BreakingConfig: %w", err)
	}
	if explicit {
		cfg.AgainstGitRef = against
	}
	return cfg, nil
}

type breakingIssueKey struct {
	path, rule, message string
	line, column        int
}
