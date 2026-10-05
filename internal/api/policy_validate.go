package api

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/modules"
	policyresolver "github.com/easyp-tech/easyp/internal/policy"
	"github.com/easyp-tech/easyp/internal/workspace"
)

type policyReadCache struct{ *gitmodules.Cache }

// Expose only cached verification during discovery, without the repository's
// version resolver promoted through policyReadCache's embedded Git cache.
type policyDiscoveryCache struct{ modules.Cache }

func (c policyReadCache) Install(ctx context.Context, lock v1.Lock) error {
	return c.VerifyCached(ctx, lock)
}

func policyReferenceValidator(ctx *cli.Context) func(string) ([]config.ValidationIssue, error) {
	return func(path string) ([]config.ValidationIssue, error) {
		switch filepath.Base(path) {
		case v1.GenerateFile, v1.ModuleFile, v1.LockFile:
			return nil, nil
		}
		file, found, err := readV1PolicyFile(path)
		if err != nil {
			return nil, err
		}
		if !found || (file.policy.Linters.Extends == "" && file.policy.Breaking.Extends == "") {
			return nil, nil
		}
		boundary, err := workspace.Boundary(filepath.Dir(path))
		if err != nil {
			return nil, err
		}
		moduleDir, err := findV1PolicyModuleDir(boundary, filepath.Dir(path))
		if err != nil {
			return nil, err
		}
		resolver := policyresolver.NewResolver(boundary, func(graphCtx context.Context, dir string) (modules.PolicyGraph, error) {
			cache, err := moduleCache(ctx)
			if err != nil {
				return nil, err
			}
			return modules.ResolvePolicyGraph(graphCtx, dir, policyReadCache{cache}, flags.IsFrozen(ctx))
		})
		var issues []config.ValidationIssue
		add := func(section string, err error) {
			if err == nil {
				return
			}
			issue := config.ValidationIssue{Code: "policy_extends", Message: fmt.Sprintf("%s.extends: %v", section, err), Severity: config.SeverityError}
			if node, ok := file.sections[section]; ok {
				issue.Line, issue.Column = node.Line, node.Column
				for i := 0; i+1 < len(node.Content); i += 2 {
					if node.Content[i].Value == "extends" {
						issue.Line, issue.Column = node.Content[i+1].Line, node.Content[i+1].Column
						break
					}
				}
			}
			issues = append(issues, issue)
		}
		contexts := []string{moduleDir}
		if moduleDir == "" {
			relative, err := filepath.Rel(boundary, filepath.Dir(path))
			if err != nil {
				return nil, err
			}
			replacements := newPolicyReplacementSources(ctx.Context, boundary, boundary, func() (modules.Cache, error) {
				cache, err := moduleCache(ctx)
				if err != nil {
					return nil, fmt.Errorf("moduleCache: %w", err)
				}
				return policyDiscoveryCache{policyReadCache{cache}}, nil
			}, flags.IsFrozen(ctx))
			scopes, scanErr := selectedPolicyScopes(replacements, relative, true)
			if scanErr == nil {
				var modules []string
				for _, name := range slices.Sorted(maps.Keys(scopes)) {
					if directory := policyDirectoryForReference(boundary, name); directory != "" {
						modules = append(modules, directory)
					}
				}
				if len(modules) > 0 {
					contexts = modules
				}
			}
		}
		for _, consumer := range contexts {
			lintApplies, breakingApplies := true, true
			if moduleDir == "" && consumer != "" {
				_, sources, _, _, err := resolveV1LintPolicyDetailed(consumer, filepath.Dir(path), path)
				if err != nil {
					return nil, err
				}
				lintApplies = sources.linters == path
				_, owner, err := resolveV1BreakingPolicy(consumer, filepath.Dir(path), path)
				if err != nil {
					return nil, err
				}
				breakingApplies = owner == path
			}
			if lintApplies && file.policy.Linters.Extends != "" {
				result, err := resolver.ResolveLint(ctx.Context, policyresolver.LintInput{PolicyPath: path, Policy: file.policy, Presence: file.presence, SettingsPath: path, Settings: file.policy.LinterSettings, SettingsSource: file.presence, ModuleDir: consumer})
				if err == nil {
					_, err = result.Policy.LintConfig()
				}
				if err != nil && consumer != "" {
					err = fmt.Errorf("consumer %s: %w", consumer, err)
				}
				add("linters", err)
			}
			if breakingApplies && file.policy.Breaking.Extends != "" {
				result, err := resolver.ResolveBreaking(ctx.Context, policyresolver.BreakingInput{PolicyPath: path, Policy: file.policy, Presence: file.presence, ModuleDir: consumer})
				if err == nil {
					_, err = result.Policy.BreakingConfig("")
				}
				if err != nil && consumer != "" {
					err = fmt.Errorf("consumer %s: %w", consumer, err)
				}
				add("breaking", err)
			}
		}
		return issues, nil
	}
}

// policyDirectoryForReference keeps discovery of absent/deleted source scopes
// separate from dependency resolution. Only real manifests supply a graph.
func policyDirectoryForReference(root, relative string) string {
	directory := filepath.Join(root, relative)
	if info, err := os.Stat(filepath.Join(directory, v1.ModuleFile)); err == nil && info.Mode().IsRegular() {
		return directory
	}
	return ""
}
