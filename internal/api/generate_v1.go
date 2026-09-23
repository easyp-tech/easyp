package api

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func (g Generate) generate(ctx *cli.Context, log logger.Logger) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	configs, err := discoverV1GenerateConfigs(workDir, ctx.String(flagGenerateProject.Name))
	if err != nil {
		return fmt.Errorf("discoverV1GenerateConfigs: %w", err)
	}
	if len(configs) == 0 {
		return fmt.Errorf("no easyp.gen.yaml found in %s", workDir)
	}

	for _, configPath := range configs {
		fp, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("Open: %w", err)
		}
		gen, parseErr := v1.ParseGenerate(fp)
		_ = fp.Close()
		if parseErr != nil {
			return fmt.Errorf("%s: %w", configPath, parseErr)
		}
		if err := inheritV1GenerateOptions(workDir, configPath, &gen); err != nil {
			return fmt.Errorf("inheritV1GenerateOptions: %w", err)
		}
		if len(gen.Plugins) == 0 {
			continue // a root defaults file may contain only options
		}
		moduleDirs, err := selectedV1ModuleDirs(workDir, filepath.Dir(configPath), gen.Generate.Modules)
		if err != nil {
			return fmt.Errorf("%s: %w", configPath, err)
		}
		if len(gen.Generate.Packages) > 0 {
			return fmt.Errorf("%s: generate.packages matching is not specified precisely enough for v1", configPath)
		}
		for _, moduleDir := range moduleDirs {
			if err := generateSelectedV1Module(ctx, log, configPath, workDir, moduleDir, gen); err != nil {
				return fmt.Errorf("generateSelectedV1Module: %w", err)
			}
		}
	}
	return nil
}

func discoverV1GenerateConfigs(root, project string) ([]string, error) {
	if project != "" {
		if !filepath.IsAbs(project) {
			project = filepath.Join(root, project)
		}
		path := filepath.Join(project, "easyp.gen.yaml")
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("find project generator %s: %w", path, err)
		}
		return []string{path}, nil
	}
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "easyp_vendor" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "easyp.gen.yaml" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func selectedV1ModuleDirs(repoRoot, configDir string, names []string) ([]string, error) {
	if len(names) == 0 {
		for dir := configDir; ; dir = filepath.Dir(dir) {
			_, err := os.Stat(filepath.Join(dir, "protobuf.mod"))
			switch {
			case errors.Is(err, os.ErrNotExist):
			case err != nil:
				return nil, fmt.Errorf("module in %s: %w", dir, err)
			default:
				return []string{dir}, nil
			}
			if dir == repoRoot || dir == filepath.Dir(dir) {
				break
			}
		}
		return nil, fmt.Errorf("no protobuf.mod for generator in %s", configDir)
	}
	result := make([]string, 0, len(names))
	for _, name := range names {
		if filepath.IsAbs(name) {
			result = append(result, name) // canonical local Git source identity
			continue
		}
		path := filepath.Clean(filepath.Join(repoRoot, name))
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("module path %q leaves repository", name)
		}
		if _, err := os.Stat(filepath.Join(path, "protobuf.mod")); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("module %q: %w", name, err)
			}
			// A module identity may name a requirement in protobuf.lock.
			result = append(result, name)
			continue
		}
		result = append(result, path)
	}
	return result, nil
}
