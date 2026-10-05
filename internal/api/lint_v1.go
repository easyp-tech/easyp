package api

import (
	"bytes"
	"context"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
	policyresolver "github.com/easyp-tech/easyp/internal/policy"
	"github.com/easyp-tech/easyp/internal/rules"
	"github.com/easyp-tech/easyp/internal/workspace"
)

func (l Lint) actionV1(ctx *cli.Context, log logger.Logger, configPath, projectRoot, lintRoot string) error {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("ReadFile: %w", err)
	}
	_, err = v1.ParsePolicy(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("ParsePolicy: %w", err)
	}
	searchDir := filepath.Join(lintRoot, ctx.String(flagLintDirectoryPath.Name))
	var files []string
	err = filepath.WalkDir(searchDir, func(path string, entry iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if policySourceExcluded(projectRoot, path) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("WalkDir: %w", err)
	}
	var cache modules.Cache
	replacements := newPolicyReplacementSources(ctx.Context, projectRoot, projectRoot, func() (modules.Cache, error) {
		if cache == nil {
			var err error
			cache, err = moduleCache(ctx)
			if err != nil {
				return nil, fmt.Errorf("moduleCache: %w", err)
			}
		}
		return cache, nil
	}, flags.IsFrozen(ctx))
	if flags.IsFrozen(ctx) {
		relative, err := filepath.Rel(projectRoot, searchDir)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		selected, err := selectedPolicyScopes(replacements, relative, true)
		if err != nil {
			return fmt.Errorf("selectedPolicyScopes: %w", err)
		}
		cache, err = moduleCache(ctx)
		if err != nil {
			return fmt.Errorf("moduleCache: %w", err)
		}
		for directory := range selected {
			if _, err := modules.EnsureFrozenSources(ctx.Context, filepath.Join(projectRoot, directory), cache); err != nil {
				return err
			}
		}
	}
	apps := map[v1LintAppKey]*core.Core{}
	batchFiles := make(map[v1LintAppKey][]string)
	var batchOrder []v1LintAppKey
	excludedByPath := make(map[string][]string)
	moduleRoots := map[string]modules.SourceRoots{}
	policyBoundary, err := workspace.Boundary(projectRoot)
	if err != nil {
		return fmt.Errorf("Boundary: %w", err)
	}
	resolver := policyresolver.NewResolver(policyBoundary, func(graphCtx context.Context, moduleDir string) (modules.PolicyGraph, error) {
		if cache == nil {
			var cacheErr error
			cache, cacheErr = moduleCache(ctx)
			if cacheErr != nil {
				return nil, fmt.Errorf("moduleCache: %w", cacheErr)
			}
		}
		return modules.ResolvePolicyGraph(graphCtx, moduleDir, cache, flags.IsFrozen(ctx))
	})
	var issues []core.IssueInfo
	for _, file := range files {
		unselectedReplacement, err := replacements.isUnselectedReplacementSource(searchDir, file)
		if err != nil {
			return fmt.Errorf("isUnselectedReplacementSource: %w", err)
		}
		if unselectedReplacement {
			continue
		}
		moduleDir, err := findV1PolicyModuleDir(projectRoot, filepath.Dir(file))
		if err != nil {
			return fmt.Errorf("findV1PolicyModuleDir: %w", err)
		}
		policy, policyKey, lintersPresence, settingsPresence, err := resolveV1LintPolicyDetailed(filepath.Dir(file), projectRoot, configPath)
		if err != nil {
			return fmt.Errorf("resolveV1LintPolicy: %w", err)
		}
		var issuePath string
		if policyKey.issues != "" {
			issuePath, err = filepath.Rel(filepath.Dir(policyKey.issues), file)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
		}
		excludedLinters, excludeFile, err := policy.Issues.ExclusionsForPath(filepath.ToSlash(issuePath))
		if err != nil {
			return fmt.Errorf("ExclusionsForPath: %w", err)
		}
		if excludeFile {
			continue
		}
		if flags.IsFrozen(ctx) && moduleDir == "" {
			return fmt.Errorf("frozen lint requires protobuf.mod for %s", file)
		}
		appKey := v1LintAppKey{policy: policyKey, moduleDir: moduleDir}
		if _, ok := apps[appKey]; !ok {
			resolved, err := resolver.ResolveLint(ctx.Context, policyresolver.LintInput{
				PolicyPath:     policyKey.linters,
				Policy:         policy,
				Presence:       lintersPresence,
				SettingsPath:   policyKey.settings,
				Settings:       policy.LinterSettings,
				SettingsSource: settingsPresence,
				ModuleDir:      moduleDir,
			})
			if err != nil {
				return fmt.Errorf("resolve policy for consuming file %s: %w", file, err)
			}
			lintConfig, err := resolved.Policy.LintConfig()
			if err != nil {
				return fmt.Errorf("LintConfig: %w", err)
			}
			if moduleDir != "" {
				_, module, moduleErr := modules.ReadManifest(moduleDir)
				if moduleErr != nil {
					return fmt.Errorf("ReadManifest: %w", moduleErr)
				}
				lintConfig.PackageDirectoryPrefix = v1PackageDirectoryPrefix(module.Name)
			}
			var importRoots modules.SourceRoots
			if moduleDir != "" {
				roots, known := moduleRoots[moduleDir]
				if !known {
					if cache == nil {
						cache, err = moduleCache(ctx)
						if err != nil {
							return fmt.Errorf("moduleCache: %w", err)
						}
					}
					roots, err = policyImportRoots(ctx.Context, cache, moduleDir, flags.IsFrozen(ctx))
					if err != nil {
						return fmt.Errorf("ensureV1PolicyImportRoots: %w", err)
					}
					moduleRoots[moduleDir] = roots
				}
				importRoots = roots
			}
			app, err := buildCore(log, config.Config{Lint: lintConfig}, importRoots)
			if err != nil {
				return fmt.Errorf("buildCore: %w", err)
			}
			apps[appKey] = app
			batchOrder = append(batchOrder, appKey)
		}
		rel, err := filepath.Rel(lintRoot, file)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		rel = filepath.ToSlash(rel)
		batchFiles[appKey] = append(batchFiles[appKey], rel)
		excludedByPath[rel] = excludedLinters
	}
	// Keep stateful rules within one effective policy/module, but avoid starting
	// a separate lint run for every file. Named path exclusions remain per file.
	for _, key := range batchOrder {
		fileIssues, err := apps[key].Lint(ctx.Context, newSelectedLintWalker(lintRoot, batchFiles[key]))
		if err != nil {
			return fmt.Errorf("Lint: %w", err)
		}
		for _, issue := range fileIssues {
			if v1IssueRuleExcluded(issue.RuleName, excludedByPath[filepath.ToSlash(issue.Path)]) {
				continue
			}
			if issue.RuleName == "PACKAGE_DIRECTORY_MATCH" && issue.SourceName != "" {
				issue.Path = filepath.ToSlash(issue.SourceName)
			}
			issues = append(issues, issue)
		}
	}
	// Grouping must not reorder user-facing diagnostics between policy subtrees.
	slices.SortStableFunc(issues, func(a, b core.IssueInfo) int {
		return strings.Compare(filepath.ToSlash(a.Path), filepath.ToSlash(b.Path))
	})
	if len(issues) == 0 {
		return nil
	}
	if err := printIssues(flags.GetFormat(ctx, flags.TextFormat), os.Stdout, issues); err != nil {
		return fmt.Errorf("printIssues: %w", err)
	}
	return ErrHasLintIssue
}

type v1LintPolicySources struct {
	linters  string
	settings string
	issues   string
}

type v1LintAppKey struct {
	policy    v1LintPolicySources
	moduleDir string
}

func v1PackageDirectoryPrefix(identity string) string {
	logical := filepath.ToSlash(strings.TrimSuffix(identity, "/"))
	if major, err := v1.ModulePathMajor(identity); err == nil && major != "" {
		logical = strings.TrimSuffix(logical, major)
	}
	logical = strings.TrimSuffix(logical, "/")
	if logical == "" {
		return ""
	}
	base := filepath.Base(logical)
	return strings.TrimSuffix(base, ".git")
}

func v1IssueRuleExcluded(name string, selections []string) bool {
	if len(selections) == 0 {
		return false
	}
	if slices.Contains(selections, name) {
		return true
	}
	for _, group := range rules.AllGroups() {
		if slices.Contains(selections, group.Key) && slices.Contains(group.Rules, name) {
			return true
		}
	}
	return false
}

func resolveV1LintPolicy(directory, projectRoot, configPath string) (v1.Policy, v1LintPolicySources, error) {
	policy, sources, _, _, err := resolveV1LintPolicyDetailed(directory, projectRoot, configPath)
	return policy, sources, err
}

func resolveV1LintPolicyDetailed(directory, projectRoot, configPath string) (v1.Policy, v1LintPolicySources, v1.PolicyPresence, v1.PolicyPresence, error) {
	var result v1.Policy
	var sources v1LintPolicySources
	lintersPresence := v1.NewPolicyPresence()
	settingsPresence := v1.NewPolicyPresence()
	var foundFile bool
	for _, dir := range ancestorDirs(directory, projectRoot) {
		path := v1PolicyPath(dir, projectRoot, configPath)
		file, found, err := readV1PolicyFile(path)
		if err != nil {
			return v1.Policy{}, v1LintPolicySources{}, v1.PolicyPresence{}, v1.PolicyPresence{}, fmt.Errorf("readV1PolicyFile: %w", err)
		}
		if !found {
			continue
		}
		foundFile = true
		if file.has("linters") && sources.linters == "" {
			result.Linters = file.policy.Linters
			sources.linters = path
			lintersPresence = file.presence
		}
		if file.has("linters-settings") && sources.settings == "" {
			result.LinterSettings = file.policy.LinterSettings
			sources.settings = path
			settingsPresence = file.presence
		}
		if file.has("issues") && sources.issues == "" {
			result.Issues = file.policy.Issues
			sources.issues = path
		}
	}
	if !foundFile {
		return v1.Policy{}, v1LintPolicySources{}, v1.PolicyPresence{}, v1.PolicyPresence{}, fmt.Errorf("no easyp.yaml policy for %s", directory)
	}
	if result.Linters.Default == "" {
		result.Linters.Default = "STANDARD"
	}
	return result, sources, lintersPresence, settingsPresence, nil
}
