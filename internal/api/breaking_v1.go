package api

import (
	"context"
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
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
	policyresolver "github.com/easyp-tech/easyp/internal/policy"
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
		if policy.Breaking.Extends != "" {
			resolved, err := b.checkV1ExtendedBreakingSource(ctx, log, configPath, projectRoot, scanRoot, scanPath, source, policy, cache, snapshots, seen)
			if err != nil {
				return nil, fmt.Errorf("checkV1ExtendedBreakingSource: %w", err)
			}
			issues = append(issues, resolved...)
			continue
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
		current, err := selectedPolicyScopes(repositoryRoot, relative, flags.IsFrozen(ctx))
		if err != nil {
			return nil, fmt.Errorf("discoverBreakingScopes current: %w", err)
		}
		baseline, err := selectedPolicyScopes(snapshot.Root, relative, flags.IsFrozen(ctx))
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
			if flags.IsFrozen(ctx) && len(ownFiles) == 0 {
				_, owner, err := resolveV1BreakingPolicy(filepath.Join(repositoryRoot, module), projectRoot, configPath)
				if err != nil {
					return nil, fmt.Errorf("resolveV1BreakingPolicy: %w", err)
				}
				relevant = owner == source
			}
			if !relevant {
				continue
			}
			currentImports, err := breakingImportRootsMode(ctx.Context, cache, repositoryRoot, module, currentFiles, nil, flags.IsFrozen(ctx))
			if err != nil {
				return nil, fmt.Errorf("current module %s: %w", module, err)
			}
			baselineImports, err := breakingImportRootsMode(ctx.Context, cache, snapshot.Root, module, baselineFiles, snapshot, flags.IsFrozen(ctx))
			if err != nil {
				return nil, fmt.Errorf("baseline %s module %s: %w", cfg.AgainstGitRef, module, err)
			}
			app := core.New(core.Options{Logger: log, ImportRoots: currentImports, BreakingCheckConfig: core.BreakingCheckConfig{
				IgnoreDirs: append(append([]string(nil), ignorePaths...), defaultVendorDir), AgainstGitRef: cfg.AgainstGitRef,
				FilesCheck:     slices.Contains(cfg.Use, core.BreakingCheckFilesCheck),
				Categories:     cfg.Categories,
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
		moduleDir, err := findV1PolicyModuleDir(projectRoot, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("findV1PolicyModuleDir: %w", err)
		}
		replacementTarget, err := isReplacementTargetModule(projectRoot, moduleDir)
		if err != nil {
			return fmt.Errorf("isReplacementTargetModule: %w", err)
		}
		if replacementTarget && !path_helpers.IsTargetPath(moduleDir, scanPath) {
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

// checkV1ExtendedBreakingSource resolves a shared breaking policy separately
// for every checked module. This is required because nested consumers can have
// different verified graphs and local replacements for the same module name.
func (b BreakingCheck) checkV1ExtendedBreakingSource(
	ctx *cli.Context, log logger.Logger,
	configPath, projectRoot, scanRoot, scanPath, source string,
	policy v1.Policy, cache modules.Cache,
	snapshots map[string]*gitadapter.Snapshot, seen map[breakingIssueKey]bool,
) ([]core.IssueInfo, error) {
	repositoryRoot, err := gitadapter.RepositoryRoot(scanRoot)
	if err != nil {
		return nil, fmt.Errorf("RepositoryRoot: %w", err)
	}
	relative, err := filepath.Rel(repositoryRoot, scanPath)
	if err != nil || !filepath.IsLocal(relative) {
		return nil, fmt.Errorf("%w: %s", core.ErrRootOutsideProject, scanPath)
	}
	current, err := selectedPolicyScopes(repositoryRoot, relative, flags.IsFrozen(ctx))
	if err != nil {
		return nil, fmt.Errorf("selectedPolicyScopes: %w", err)
	}
	// Retain a native module after deletion of its last proto file, even in
	// normal mode. Frozen's additional empty-scope checks remain mandatory.
	if !flags.IsFrozen(ctx) {
		expanded, expandErr := selectedPolicyScopes(repositoryRoot, relative, true)
		if expandErr == nil {
			current = expanded
		}
	}
	policyFile, found, err := readV1PolicyFile(source)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("policy source %s disappeared", source)
	}

	type candidate struct {
		module     string
		historical *gitadapter.Snapshot
	}
	var queue []candidate
	queued := make(map[string]bool)
	enqueue := func(scopes map[string]breakingScope, historical *gitadapter.Snapshot) error {
		for _, module := range slices.Sorted(maps.Keys(scopes)) {
			if queued[module] {
				continue
			}
			relevant, err := breakingScopeUsesPolicy(repositoryRoot, scopes[module].files, projectRoot, configPath, source)
			if err != nil {
				return err
			}
			if len(scopes[module].files) == 0 {
				_, owner, err := resolveV1BreakingPolicy(filepath.Join(repositoryRoot, module), projectRoot, configPath)
				if err != nil {
					return err
				}
				relevant = owner == source
			}
			if relevant {
				queued[module] = true
				queue = append(queue, candidate{module: module, historical: historical})
			}
		}
		return nil
	}
	if err := enqueue(current, nil); err != nil {
		return nil, err
	}
	baselineScopes := make(map[string]map[string]breakingScope)
	loadBaseline := func(ref string) (*gitadapter.Snapshot, map[string]breakingScope, error) {
		snapshot := snapshots[ref]
		if snapshot == nil {
			var err error
			snapshot, err = gitadapter.SnapshotRevision(ctx.Context, scanRoot, ref)
			if err != nil {
				return nil, nil, fmt.Errorf("SnapshotRevision: %w", err)
			}
			snapshots[ref] = snapshot
		}
		scopes, known := baselineScopes[ref]
		if !known {
			var err error
			scopes, err = selectedPolicyScopes(snapshot.Root, relative, flags.IsFrozen(ctx))
			if err != nil {
				return nil, nil, fmt.Errorf("selectedPolicyScopes: %w", err)
			}
			baselineScopes[ref] = scopes
		}
		return snapshot, scopes, nil
	}
	resolver := policyresolver.NewResolver(repositoryRoot, func(graphCtx context.Context, dir string) (modules.PolicyGraph, error) {
		for _, snapshot := range snapshots {
			relative, err := filepath.Rel(snapshot.Root, dir)
			if err == nil && filepath.IsLocal(relative) {
				return modules.ResolvePolicyGraphAt(graphCtx, dir, cache, flags.IsFrozen(ctx), func(target string) (string, error) { return snapshotReplacementPath(snapshot, dir, target) })
			}
		}
		return modules.ResolvePolicyGraph(graphCtx, dir, cache, flags.IsFrozen(ctx))
	})
	if len(queue) == 0 {
		// A wholly deleted scope has no current module from which to load a remote
		// policy. An explicit baseline, or a resolvable local policy, supplies the
		// historical graph. Never silently succeed without comparing the deletion.
		var cfg config.BreakingCheck
		if ctx.IsSet(flagAgainstBranchName.Name) {
			unresolved := policy
			unresolved.Breaking.Extends = ""
			cfg, err = resolveV1BreakingConfig(ctx, unresolved)
		} else {
			moduleDir, findErr := findV1PolicyModuleDir(projectRoot, filepath.Dir(source))
			if findErr != nil {
				return nil, findErr
			}
			resolved, resolveErr := resolver.ResolveBreaking(ctx.Context, policyresolver.BreakingInput{PolicyPath: source, Policy: policy, Presence: policyFile.presence, ModuleDir: moduleDir})
			if resolveErr != nil {
				return nil, fmt.Errorf("deleted or empty scope needs --against to locate its historical policy graph: %w", resolveErr)
			}
			cfg, err = resolveV1BreakingConfig(ctx, resolved.Policy)
		}
		if err != nil {
			return nil, err
		}
		snapshot, baseline, err := loadBaseline(cfg.AgainstGitRef)
		if err != nil {
			return nil, err
		}
		if err := enqueue(baseline, snapshot); err != nil {
			return nil, err
		}
	}
	var result []core.IssueInfo
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		module := item.module
		moduleDir := policyDirectoryForReference(repositoryRoot, module)
		if moduleDir == "" && item.historical != nil {
			moduleDir = policyDirectoryForReference(item.historical.Root, module)
		}
		resolved, err := resolver.ResolveBreaking(ctx.Context, policyresolver.BreakingInput{PolicyPath: source, Policy: policy, Presence: policyFile.presence, ModuleDir: moduleDir})
		if err != nil {
			return nil, fmt.Errorf("resolve policy for consuming module %s: %w", module, err)
		}
		cfg, err := resolveV1BreakingConfig(ctx, resolved.Policy)
		if err != nil {
			return nil, fmt.Errorf("resolveV1BreakingConfig: %w", err)
		}
		snapshot, baseline, err := loadBaseline(cfg.AgainstGitRef)
		if err != nil {
			return nil, err
		}
		if err := enqueue(baseline, snapshot); err != nil {
			return nil, err
		}
		currentFiles, baselineFiles := current[module].files, baseline[module].files
		ownFiles := append(append([]string(nil), currentFiles...), baselineFiles...)
		ignorePaths, err := breakingIgnorePaths(repositoryRoot, filepath.Dir(source), cfg.Ignore)
		if err != nil {
			return nil, fmt.Errorf("breakingIgnorePaths: %w", err)
		}
		currentImports, err := breakingImportRootsMode(ctx.Context, cache, repositoryRoot, module, currentFiles, nil, flags.IsFrozen(ctx))
		if err != nil {
			return nil, fmt.Errorf("current module %s: %w", module, err)
		}
		baselineImports, err := breakingImportRootsMode(ctx.Context, cache, snapshot.Root, module, baselineFiles, snapshot, flags.IsFrozen(ctx))
		if err != nil {
			return nil, fmt.Errorf("baseline %s module %s: %w", cfg.AgainstGitRef, module, err)
		}
		app := core.New(core.Options{Logger: log, ImportRoots: currentImports, BreakingCheckConfig: core.BreakingCheckConfig{
			IgnoreDirs: append(append([]string(nil), ignorePaths...), defaultVendorDir), AgainstGitRef: cfg.AgainstGitRef,
			FilesCheck: slices.Contains(cfg.Use, core.BreakingCheckFilesCheck), Categories: cfg.Categories, IgnoreUnstable: cfg.IgnoreUnstable,
		}})
		findings, err := app.CompareBreaking(ctx.Context, newBreakingWalker(repositoryRoot, currentFiles), newBreakingWalker(snapshot.Root, baselineFiles), baselineImports)
		if err != nil {
			return nil, fmt.Errorf("CompareBreaking module %s against %s: %w", module, cfg.AgainstGitRef, err)
		}
		for _, issue := range findings {
			if slices.Contains(ownFiles, filepath.ToSlash(issue.Path)) {
				_, owner, err := resolveV1BreakingPolicy(filepath.Dir(filepath.Join(repositoryRoot, issue.Path)), projectRoot, configPath)
				if err != nil {
					return nil, err
				}
				if owner != source {
					continue
				}
			} else {
				issue.Path = filepath.ToSlash(filepath.Join(module, issue.Path))
			}
			key := breakingIssueKey{path: issue.Path, rule: issue.RuleName, message: issue.Message, line: issue.Position.Line, column: issue.Position.Column}
			if !seen[key] {
				result = append(result, issue)
				seen[key] = true
			}
		}
	}
	return result, nil
}

type breakingIssueKey struct {
	path, rule, message string
	line, column        int
}
