package api

import (
	"bytes"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

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
		if !entry.IsDir() && filepath.Ext(path) == ".proto" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("WalkDir: %w", err)
	}
	apps := map[v1LintAppKey]*core.Core{}
	moduleRoots := map[string][]string{}
	var issues []core.IssueInfo
	for _, file := range files {
		policy, policyKey, err := resolveV1LintPolicy(filepath.Dir(file), projectRoot, configPath)
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
		appKey := v1LintAppKey{policy: policyKey, moduleDir: moduleDir}
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

type v1LintPolicySources struct {
	linters  string
	settings string
	issues   string
}

type v1LintAppKey struct {
	policy    v1LintPolicySources
	moduleDir string
}

func resolveV1LintPolicy(directory, projectRoot, configPath string) (v1.Policy, v1LintPolicySources, error) {
	var result v1.Policy
	var sources v1LintPolicySources
	var foundFile bool
	for _, dir := range ancestorDirs(directory, projectRoot) {
		path := v1PolicyPath(dir, projectRoot, configPath)
		file, found, err := readV1PolicyFile(path)
		if err != nil {
			return v1.Policy{}, v1LintPolicySources{}, fmt.Errorf("readV1PolicyFile: %w", err)
		}
		if !found {
			continue
		}
		foundFile = true
		if file.has("linters") && sources.linters == "" {
			result.Linters = file.policy.Linters
			sources.linters = path
		}
		if file.has("linters-settings") && sources.settings == "" {
			result.LinterSettings = file.policy.LinterSettings
			sources.settings = path
		}
		if file.has("issues") && sources.issues == "" {
			result.Issues = file.policy.Issues
			sources.issues = path
		}
	}
	if !foundFile {
		return v1.Policy{}, v1LintPolicySources{}, fmt.Errorf("no easyp.yaml policy for %s", directory)
	}
	if result.Linters.Default == "" {
		result.Linters.Default = "STANDARD"
	}
	return result, sources, nil
}
