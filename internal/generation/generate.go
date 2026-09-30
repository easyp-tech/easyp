package generation

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
	"github.com/easyp-tech/easyp/internal/workspace"
)

// Request contains the inputs of one generation run.
type Request struct {
	WorkDir             string
	Project             string
	Projects            []string
	AllProjects         bool
	WorkspaceRoot       string
	DescriptorSetOut    string
	DescriptorSetOutDir string
	IncludeImports      bool
}

type generationTarget struct {
	configPath string
	config     v1.Generate
	module     v1ModuleSelection
}

// Run discovers generator files, prepares selected modules, and executes each generation.
// Missing locked dependencies are installed in the supplied cache.
func Run(ctx context.Context, log logger.Logger, cache modules.Cache, request Request) error {
	if request.DescriptorSetOut != "" && request.DescriptorSetOutDir != "" {
		return fmt.Errorf("--descriptor_set_out and --descriptor_set_out_dir are mutually exclusive")
	}
	exportDescriptors := request.DescriptorSetOut != "" || request.DescriptorSetOutDir != ""
	workDir, err := filepath.Abs(request.WorkDir)
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	request.WorkDir = workDir
	if request.WorkspaceRoot == "" {
		request.WorkspaceRoot, err = workspace.Boundary(workDir)
	} else {
		if !filepath.IsAbs(request.WorkspaceRoot) {
			request.WorkspaceRoot = filepath.Join(workDir, request.WorkspaceRoot)
		}
		request.WorkspaceRoot, err = filepath.Abs(request.WorkspaceRoot)
	}
	if err != nil {
		return fmt.Errorf("Boundary: %w", err)
	}
	workspaceRel, err := filepath.Rel(request.WorkspaceRoot, workDir)
	if err != nil || !filepath.IsLocal(workspaceRel) {
		return fmt.Errorf("working directory must be inside --workspace")
	}
	if request.DescriptorSetOut != "" && !filepath.IsAbs(request.DescriptorSetOut) {
		request.DescriptorSetOut = filepath.Join(workDir, request.DescriptorSetOut)
	}
	if request.DescriptorSetOutDir != "" && !filepath.IsAbs(request.DescriptorSetOutDir) {
		request.DescriptorSetOutDir = filepath.Join(workDir, request.DescriptorSetOutDir)
	}
	configs, err := selectGenerateConfigs(request)
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
		if err := inheritV1GenerateOptions(request.WorkspaceRoot, configPath, &gen); err != nil {
			return fmt.Errorf("inheritV1GenerateOptions: %w", err)
		}
		if len(gen.Plugins) == 0 && !exportDescriptors {
			continue
		}
		modules, err := selectV1Modules(request.WorkspaceRoot, filepath.Dir(configPath), gen.Generate.Modules)
		if err != nil {
			return fmt.Errorf("%s: %w", configPath, err)
		}
		if len(gen.Generate.Packages) > 0 {
			return fmt.Errorf("%s: generate.packages matching is not specified precisely enough for v1", configPath)
		}
		for _, module := range modules {
			descriptorTargets = append(descriptorTargets, generationTarget{configPath: configPath, config: gen, module: module})
		}
	}
	if request.AllProjects {
		selected := descriptorTargets[:0]
		for _, target := range descriptorTargets {
			if len(target.config.Plugins) == 0 && len(target.config.Generate.Modules) == 0 {
				inherited, err := isOptionsOnlyParent(target.configPath, target.config, configs)
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
	if len(descriptorTargets) == 0 && exportDescriptors {
		return fmt.Errorf("descriptor set has no selected modules")
	}
	if !exportDescriptors {
		if len(descriptorTargets) == 0 {
			return nil
		}
		prepared, err := prepareDescriptorTargets(ctx, log, cache, request, descriptorTargets)
		if err != nil {
			return fmt.Errorf("prepareDescriptorTargets: %w", err)
		}
		return executePreparedTargets(ctx, prepared)
	}
	if err := generateV1DescriptorSet(ctx, log, cache, request, descriptorTargets); err != nil {
		return fmt.Errorf("generateV1DescriptorSet: %w", err)
	}
	return nil
}

// isOptionsOnlyParent reports whether a config has child generators but no
// declared module of its own. Explicit options-only ancestors do not acquire
// ownership of incidental proto files elsewhere in the repository.
func isOptionsOnlyParent(configPath string, gen v1.Generate, configs []string) (bool, error) {
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

	// A thin options-only ancestor is not an implicit module. Unrelated proto
	// files (for example a legacy tree) must not make it a generation target.
	if gen.Options.Go.PackagePrefix != nil && !gen.InheritedGoPackagePrefix &&
		!gen.Generate.Managed.Enabled && len(gen.Generate.Managed.Override) == 0 && len(gen.Generate.Managed.Disable) == 0 {
		return true, nil
	}

	hasOwnProto := false
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if childDirs[path] || path_helpers.ShouldSkipV1SourceDir(dir, path) {
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
			if workspace.SkipDirectory(root, path) {
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
