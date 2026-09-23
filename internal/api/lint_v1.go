package api

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func (l Lint) actionV1(ctx *cli.Context, log logger.Logger, configPath, projectRoot, lintRoot string) error {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("ReadFile: %w", err)
	}
	_, err = v1.ParsePolicy(strings.NewReader(string(raw)))
	if err != nil {
		return fmt.Errorf("ParsePolicy: %w", err)
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
		return fmt.Errorf("WalkDir: %w", err)
	}
	apps := map[string]*core.Core{}
	moduleRoots := map[string][]string{}
	var issues []core.IssueInfo
	for _, file := range files {
		policy, policyKey, err := resolveV1LintPolicy(file, projectRoot, configPath)
		if err != nil {
			return fmt.Errorf("resolveV1LintPolicy: %w", err)
		}
		if policy.ExcludesAllIssues() {
			continue
		}
		moduleDir, err := findV1PolicyModuleDir(projectRoot, filepath.Dir(file))
		if err != nil {
			return fmt.Errorf("findV1PolicyModuleDir: %w", err)
		}
		appKey := policyKey + "|" + moduleDir
		app, ok := apps[appKey]
		if !ok {
			lintConfig, err := policy.LintConfig()
			if err != nil {
				return fmt.Errorf("LintConfig: %w", err)
			}
			app, err = buildCore(log, config.Config{Lint: lintConfig})
			if err != nil {
				return fmt.Errorf("buildCore: %w", err)
			}
			if moduleDir != "" {
				roots, known := moduleRoots[moduleDir]
				if !known {
					roots, err = resolveV1PolicyImportRoots(ctx.Context, log, moduleDir)
					if err != nil {
						return fmt.Errorf("resolveV1PolicyImportRoots: %w", err)
					}
					moduleRoots[moduleDir] = roots
				}
				app.SetImportRoots(roots)
			}
			apps[appKey] = app
		}
		rel, err := filepath.Rel(lintRoot, file)
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		fileIssues, err := app.Lint(ctx.Context, fs.NewFSWalker(lintRoot, rel))
		if err != nil {
			return fmt.Errorf("Lint: %w", err)
		}
		issues = append(issues, fileIssues...)
	}
	if len(issues) == 0 {
		return nil
	}
	if err := printIssues(flags.GetFormat(ctx, flags.TextFormat), os.Stdout, issues); err != nil {
		return fmt.Errorf("printIssues: %w", err)
	}
	return ErrHasLintIssue
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
		raw, found, err := readOptionalFile(path)
		if err != nil {
			return v1.Policy{}, "", fmt.Errorf("%s: %w", path, err)
		}
		if found {
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
