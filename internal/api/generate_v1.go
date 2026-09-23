package api

import (
	"bytes"
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
		gen, err := readV1GenerateConfig(configPath)
		if err != nil {
			return fmt.Errorf("readV1GenerateConfig: %w", err)
		}
		if err := inheritV1GenerateOptions(workDir, configPath, &gen); err != nil {
			return fmt.Errorf("inheritV1GenerateOptions: %w", err)
		}
		if len(gen.Plugins) == 0 {
			continue
		}
		modules, err := selectV1Modules(workDir, filepath.Dir(configPath), gen.Generate.Modules)
		if err != nil {
			return fmt.Errorf("%s: %w", configPath, err)
		}
		if len(gen.Generate.Packages) > 0 {
			return fmt.Errorf("%s: generate.packages matching is not specified precisely enough for v1", configPath)
		}
		for _, module := range modules {
			if err := generateSelectedV1Module(ctx, log, configPath, workDir, module, gen); err != nil {
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
		path := filepath.Join(project, v1.GenerateFile)
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
		if entry.Name() == v1.GenerateFile {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func readV1GenerateConfig(path string) (v1.Generate, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return v1.Generate{}, fmt.Errorf("ReadFile: %w", err)
	}
	gen, err := v1.ParseGenerate(bytes.NewReader(raw))
	if err != nil {
		return v1.Generate{}, fmt.Errorf("%s: %w", path, err)
	}
	return gen, nil
}
