package generation

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

type preparedDescriptorTarget struct {
	target   generationTarget
	selected v1GenerationModule
	plan     *core.GenerationPlan
	full     *descriptorpb.FileDescriptorSet
	owners   map[string]string
	versions map[string]string
	label    string
}

// generateV1DescriptorSet validates every requested graph before executing any
// plugin. Export uses those same prepared graphs, never a second compilation.
func generateV1DescriptorSet(ctx context.Context, log logger.Logger, cache modules.Cache, request Request, targets []generationTarget) error {
	prepared, err := prepareDescriptorTargets(ctx, log, cache, request, targets)
	if err != nil {
		return fmt.Errorf("prepareDescriptorTargets: %w", err)
	}
	outputs, err := planDescriptorOutputs(request, prepared)
	if err != nil {
		return fmt.Errorf("planDescriptorOutputs: %w", err)
	}
	if err := executePreparedTargets(ctx, prepared); err != nil {
		return err
	}
	for _, output := range outputs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := output.write(); err != nil {
			return fmt.Errorf("write %s: %w", output.path, err)
		}
		log.Info(ctx, "descriptor set written", slog.String("path", output.path))
	}
	return nil
}

func prepareDescriptorTargets(ctx context.Context, log logger.Logger, cache modules.Cache, request Request, targets []generationTarget) ([]preparedDescriptorTarget, error) {
	prepared := make([]preparedDescriptorTarget, 0, len(targets))
	seen := make(map[string]v1ModuleSelection)
	requested := make(map[string]*sourceSelectorMatches)
	for _, target := range targets {
		if target.config.Generate.HasSourceSelectors() {
			requested[target.configPath] = newSourceSelectorMatches(target.config.Generate.Packages, target.config.Generate.Paths)
		}
	}
	for _, target := range targets {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		selected, err := resolveV1GenerationModule(ctx, cache, request.WorkspaceRoot, filepath.Dir(target.configPath), target.module, request.Frozen)
		if err != nil {
			return nil, fmt.Errorf("resolveV1GenerationModule for %s: %w", target.configPath, err)
		}
		key := filepath.Clean(target.configPath) + "\x00" + filepath.Clean(selected.directory)
		if previous, exists := seen[key]; exists {
			if !sameSourceSelectors(previous.packages, target.module.packages) || !sameSourceSelectors(previous.paths, target.module.paths) {
				return nil, fmt.Errorf("%s: generate.modules[%d] and generate.modules[%d] select module %q with conflicting filters", target.configPath, previous.index, target.module.index, selected.module.Name)
			}
			continue
		}
		seen[key] = target.module
		moduleSelectors := newSourceSelectorMatches(target.module.packages, target.module.paths)
		// Options-only selections still validate their graph in frozen mode, but
		// have no generation work unless descriptor export was requested.
		if len(target.config.Plugins) == 0 && !target.config.Generate.HasSourceSelectors() && request.DescriptorSetOut == "" && request.DescriptorSetOutDir == "" {
			continue
		}
		var files []string
		if selectors, filtered := requested[target.configPath]; filtered {
			files, err = selectedSourceFiles(ctx, selected, selectors, moduleSelectors)
			if err != nil {
				return nil, fmt.Errorf("selectedSourceFiles for %s: %w", target.configPath, err)
			}
			if err := moduleSelectors.validateMatches(fmt.Sprintf("generate.modules[%d]", target.module.index)); err != nil {
				return nil, fmt.Errorf("%s: %w", target.configPath, err)
			}
			if len(files) == 0 {
				continue
			}
		}
		label := fmt.Sprintf("project %q, module %q (%s)", relativeDescriptorPath(request.WorkDir, filepath.Dir(target.configPath)), selected.module.Name, relativeDescriptorPath(request.WorkDir, selected.directory))
		app, err := prepareV1ModuleCore(log, request, target.configPath, selected.directory, target.config, selected.module, selected.dependencies)
		if err != nil {
			return nil, fmt.Errorf("prepareV1ModuleCore for %s: %w", label, err)
		}
		var plan *core.GenerationPlan
		if files != nil {
			plan, err = app.PrepareGenerationFiles(ctx, selected.directory, files)
		} else {
			plan, err = app.PrepareGeneration(ctx, selected.directory)
		}
		if err != nil {
			return nil, fmt.Errorf("PrepareGeneration for %s: %w", label, err)
		}
		full := plan.DescriptorSet(true)
		if _, err := protodesc.NewFiles(full); err != nil {
			return nil, fmt.Errorf("validate descriptor set for %s: %w", label, err)
		}
		roots, err := modules.ModuleSources(selected.directory, selected.module)
		if err != nil {
			return nil, fmt.Errorf("ModuleSources: %w", err)
		}
		roots = append(roots, selected.dependencies...)
		owners, err := roots.FileModules()
		if err != nil {
			return nil, fmt.Errorf("FileModules: %w", err)
		}
		versions := descriptorVersions(request, selected)
		prepared = append(prepared, preparedDescriptorTarget{target: target, selected: selected, plan: plan, full: full, owners: owners, versions: versions, label: label})
	}
	for _, path := range slices.Sorted(maps.Keys(requested)) {
		selectors := requested[path]
		if err := selectors.validateMatches("generate"); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	return prepared, nil
}

func sameSourceSelectors(first, second []string) bool {
	a, b := slices.Clone(first), slices.Clone(second)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(slices.Compact(a), slices.Compact(b))
}

func unmatchedSourceSelectors(requested []string, matched map[string]bool) []string {
	var unknown []string
	for _, name := range requested {
		if !matched[name] && !slices.Contains(unknown, name) {
			unknown = append(unknown, name)
		}
	}
	return unknown
}

func combineDescriptorTargets(targets []preparedDescriptorTarget, includeImports bool) (*descriptorpb.FileDescriptorSet, error) {
	combined := &descriptorpb.FileDescriptorSet{}
	files := make(map[string]*descriptorpb.FileDescriptorProto)
	origins := make(map[string]int)
	selectedFiles := make(map[string]bool)
	var labels []string
	for index, target := range targets {
		labels = append(labels, target.label)
		for _, file := range target.plan.DescriptorSet(false).File {
			selectedFiles[file.GetName()] = true
		}
		for _, file := range target.full.File {
			name := file.GetName()
			if previous, ok := files[name]; ok {
				if !proto.Equal(previous, file) {
					return nil, fmt.Errorf("conflicting descriptor %q: %s; first: %s; second: %s; select a compatible --project or use --descriptor_set_out_dir for independent graphs", name, descriptorDifference(previous.ProtoReflect(), file.ProtoReflect(), ""), targets[origins[name]].origin(name), target.origin(name))
				}
				continue
			}
			files[name], origins[name] = file, index
			combined.File = append(combined.File, file)
		}
	}
	if _, err := protodesc.NewFiles(combined); err != nil {
		return nil, fmt.Errorf("validate combined descriptor set for [%s]: %w", strings.Join(labels, "; "), err)
	}
	if includeImports {
		return combined, nil
	}
	output := &descriptorpb.FileDescriptorSet{}
	for _, file := range combined.File {
		if selectedFiles[file.GetName()] {
			output.File = append(output.File, file)
		}
	}
	return output, nil
}

func relativeDescriptorPath(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func executePreparedTargets(ctx context.Context, prepared []preparedDescriptorTarget) error {
	bucket := core.NewGenerateBucket()
	for _, target := range prepared {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := target.plan.ExecuteInto(ctx, bucket); err != nil {
			return fmt.Errorf("ExecuteInto for %s: %w", target.label, err)
		}
	}
	if err := bucket.DumpToFs(ctx); err != nil {
		return fmt.Errorf("DumpToFs: %w", err)
	}
	return nil
}
