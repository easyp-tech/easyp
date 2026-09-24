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

type generationTarget struct {
	configPath string
	config     v1.Generate
	module     v1ModuleSelection
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

	var descriptorTargets []generationTarget
	for _, configPath := range configs {
		gen, err := readV1GenerateConfig(configPath)
		if err != nil {
			return fmt.Errorf("readV1GenerateConfig: %w", err)
		}
		if err := inheritV1GenerateOptions(workDir, configPath, &gen); err != nil {
			return fmt.Errorf("inheritV1GenerateOptions: %w", err)
		}
		if len(gen.Plugins) == 0 && request.DescriptorSetOut == "" {
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
			if request.DescriptorSetOut == "" {
				if err := generateSelectedV1Module(ctx, log, cache, request, configPath, workDir, module, gen); err != nil {
					return fmt.Errorf("generateSelectedV1Module: %w", err)
				}
				continue
			}
			descriptorTargets = append(descriptorTargets, generationTarget{configPath: configPath, config: gen, module: module})
		}
	}
	if request.DescriptorSetOut == "" {
		return nil
	}
	if request.Project == "" {
		selected := descriptorTargets[:0]
		for _, target := range descriptorTargets {
			if len(target.config.Plugins) == 0 && len(target.config.Generate.Modules) == 0 {
				inherited, err := isOptionsOnlyParent(target.configPath, configs)
				if err != nil {
					return fmt.Errorf("isOptionsOnlyParent: %w", err)
				}
				if inherited {
					continue
				}
			}
			selected = append(selected, target)
		}
		descriptorTargets = selected
	}
	if len(descriptorTargets) != 1 {
		return fmt.Errorf("descriptor set requires exactly one selected module; got %d", len(descriptorTargets))
	}
	target := descriptorTargets[0]
	if err := generateSelectedV1Module(ctx, log, cache, request, target.configPath, workDir, target.module, target.config); err != nil {
		return fmt.Errorf("generateSelectedV1Module: %w", err)
	}
	return nil
}

// isOptionsOnlyParent reports whether a config has child generators but no
// module or proto files of its own. Such a config only supplies inherited options.
func isOptionsOnlyParent(configPath string, configs []string) (bool, error) {
	dir := filepath.Dir(configPath)
	childDirs := make(map[string]bool)
	for _, other := range configs {
		if other == configPath {
			continue
		}
		childDir := filepath.Dir(other)
		rel, err := filepath.Rel(dir, childDir)
		if err != nil {
			return false, err
		}
		if filepath.IsLocal(rel) {
			childDirs[childDir] = true
		}
	}
	if len(childDirs) == 0 {
		return false, nil
	}
	if _, err := os.Stat(filepath.Join(dir, v1.ModuleFile)); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	hasOwnProto := false
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if childDirs[path] {
				return filepath.SkipDir
			}
			if path != dir && (entry.Name() == ".git" || entry.Name() == "easyp_vendor" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".proto" {
			hasOwnProto = true
			return filepath.SkipAll
		}
		return nil
	})
	return !hasOwnProto, err
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
