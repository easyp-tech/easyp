package generation

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
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
	for _, target := range prepared {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := target.plan.Execute(ctx); err != nil {
			return fmt.Errorf("Execute for %s: %w", target.label, err)
		}
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
	seen := make(map[string]bool)
	for _, target := range targets {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		selected, err := resolveV1GenerationModule(ctx, cache, request.WorkDir, filepath.Dir(target.configPath), target.module)
		if err != nil {
			return nil, fmt.Errorf("resolveV1GenerationModule for %s: %w", target.configPath, err)
		}
		key := filepath.Clean(target.configPath) + "\x00" + filepath.Clean(selected.directory)
		if seen[key] {
			continue
		}
		seen[key] = true
		label := fmt.Sprintf("project %q, module %q (%s)", relativeDescriptorPath(request.WorkDir, filepath.Dir(target.configPath)), selected.module.Name, relativeDescriptorPath(request.WorkDir, selected.directory))
		app, err := prepareV1ModuleCore(log, request, target.configPath, selected.directory, target.config, selected.module, selected.dependencies)
		if err != nil {
			return nil, fmt.Errorf("prepareV1ModuleCore for %s: %w", label, err)
		}
		plan, err := app.PrepareGeneration(ctx, selected.directory)
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
	return prepared, nil
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
