package generation

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// Request contains the inputs of one generation run.
type Request struct {
	WorkDir          string
	Project          string
	DescriptorSetOut string
	IncludeImports   bool
}

// Run discovers generator files, prepares selected modules, and executes each generation.
// Missing locked dependencies are installed in the supplied cache.
func Run(ctx context.Context, log logger.Logger, cache modules.Cache, request Request) error {
	workDir, err := filepath.Abs(request.WorkDir)
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	request.WorkDir = workDir
	if request.DescriptorSetOut != "" && !filepath.IsAbs(request.DescriptorSetOut) {
		request.DescriptorSetOut = filepath.Join(workDir, request.DescriptorSetOut)
	}
	configs, err := discoverV1GenerateConfigs(workDir, request.Project)
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
			if err := generateSelectedV1Module(ctx, log, cache, request, configPath, workDir, module, gen); err != nil {
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
