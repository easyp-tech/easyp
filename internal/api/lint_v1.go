package api

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func (l Lint) actionV1(ctx *cli.Context, log logger.Logger, configPath, projectRoot, lintRoot string) (bool, error) {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return false, nil // legacy path reports its usual missing-config error
	}
	isV1, err := v1.IsPolicyConfig(raw)
	if err != nil {
		return true, fmt.Errorf("IsPolicyConfig: %w", err)
	}
	if !isV1 {
		return false, nil
	}
	searchDir := filepath.Join(lintRoot, ctx.String(flagLintDirectoryPath.Name))
	var files []string
	err = filepath.WalkDir(searchDir, func(path string, entry iofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return true, err
	}
	apps := map[string]*core.Core{}
	var issues []core.IssueInfo
	for _, file := range files {
		policy, policyKey, err := resolveV1LintPolicy(file, projectRoot, configPath)
		if err != nil {
			return true, err
		}
		if policy.ExcludesAllIssues() {
			continue
		}
		app, ok := apps[policyKey]
		if !ok {
			lintConfig, err := policy.LegacyLint()
			if err != nil {
				return true, err
			}
			app, err = buildCoreWithModFile(log, config.Config{Lint: lintConfig}, fs.NewFSWalker(projectRoot, "."), &modfile.File{})
			if err != nil {
				return true, fmt.Errorf("build v1 linter: %w", err)
			}
			apps[policyKey] = app
		}
		rel, err := filepath.Rel(lintRoot, file)
		if err != nil {
			return true, err
		}
		fileIssues, err := app.LintV1(ctx.Context, fs.NewFSWalker(lintRoot, rel))
		if err != nil {
			return true, err
		}
		issues = append(issues, fileIssues...)
	}
	if len(issues) == 0 {
		return true, nil
	}
	if err := printIssues(flags.GetFormat(ctx, flags.TextFormat), os.Stdout, issues); err != nil {
		return true, err
	}
	return true, ErrHasLintIssue
}

func resolveV1LintPolicy(file, projectRoot, configPath string) (v1.Policy, string, error) {
	var result v1.Policy
	sources := map[string]string{}
	var foundFile bool
	for dir := filepath.Dir(file); ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, "easyp.yaml")
		if dir == projectRoot {
			path = configPath
		}
		if raw, err := os.ReadFile(path); err == nil {
			foundFile = true
			var sections map[string]yaml.Node
			if err := yaml.Unmarshal(raw, &sections); err != nil {
				return v1.Policy{}, "", fmt.Errorf("%s: %w", path, err)
			}
			policy, err := v1.ParsePolicy(strings.NewReader(string(raw)))
			if err != nil {
				return v1.Policy{}, "", fmt.Errorf("%s: %w", path, err)
			}
			if _, ok := sections["linters"]; ok && sources["linters"] == "" {
				result.Linters = policy.Linters
				sources["linters"] = path
			}
			if _, ok := sections["linters-settings"]; ok && sources["linters-settings"] == "" {
				result.LinterSettings = policy.LinterSettings
				sources["linters-settings"] = path
			}
			if _, ok := sections["issues"]; ok && sources["issues"] == "" {
				result.Issues = policy.Issues
				sources["issues"] = path
			}
		} else if !os.IsNotExist(err) {
			return v1.Policy{}, "", err
		}
		if dir == projectRoot || dir == filepath.Dir(dir) {
			break
		}
	}
	if !foundFile {
		return v1.Policy{}, "", fmt.Errorf("no easyp.yaml policy for %s", file)
	}
	if result.Linters.Default == "" {
		result.Linters.Default = "STANDARD"
	}
	key := sources["linters"] + "|" + sources["linters-settings"] + "|" + sources["issues"]
	return result, key, nil
}
