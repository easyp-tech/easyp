package generation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// generateV1DescriptorSet collects each module's full descriptor graph before
// replacing the requested output. The full graph is validated as one protobuf
// registry; dependency descriptors are filtered only after validation.
func generateV1DescriptorSet(ctx context.Context, log logger.Logger, cache modules.Cache, request Request, targets []generationTarget) error {
	temporary, err := os.CreateTemp(filepath.Dir(request.DescriptorSetOut), ".easyp-descriptor-*.pb")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}

	combined := &descriptorpb.FileDescriptorSet{}
	files := make(map[string]*descriptorpb.FileDescriptorProto)
	selectedFiles := make(map[string]bool)
	for _, target := range targets {
		selected, err := resolveV1GenerationModule(ctx, cache, request.WorkDir, filepath.Dir(target.configPath), target.module)
		if err != nil {
			return fmt.Errorf("resolveV1GenerationModule: %w", err)
		}
		owned, err := v1ModuleDescriptorNames(selected)
		if err != nil {
			return fmt.Errorf("v1ModuleDescriptorNames: %w", err)
		}
		for name := range owned {
			selectedFiles[name] = true
		}

		moduleRequest := request
		moduleRequest.DescriptorSetOut = temporaryPath
		moduleRequest.IncludeImports = true
		if err := generateV1ModuleWithRoots(ctx, log, moduleRequest, target.configPath, selected.directory, target.config, selected.module, selected.dependencies); err != nil {
			return fmt.Errorf("generateV1ModuleWithRoots: %w", err)
		}
		raw, err := os.ReadFile(temporaryPath)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		var moduleSet descriptorpb.FileDescriptorSet
		if err := proto.Unmarshal(raw, &moduleSet); err != nil {
			return fmt.Errorf("Unmarshal: %w", err)
		}
		for _, file := range moduleSet.File {
			name := file.GetName()
			if previous, ok := files[name]; ok {
				if !proto.Equal(previous, file) {
					return fmt.Errorf("conflicting descriptor %q", name)
				}
				continue
			}
			files[name] = file
			combined.File = append(combined.File, file)
		}
	}

	if _, err := protodesc.NewFiles(combined); err != nil {
		return fmt.Errorf("validate combined descriptor set: %w", err)
	}

	output := combined
	if !request.IncludeImports {
		output = &descriptorpb.FileDescriptorSet{}
		for _, file := range combined.File {
			if selectedFiles[file.GetName()] {
				output.File = append(output.File, file)
			}
		}
	}
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(output)
	if err != nil {
		return fmt.Errorf("Marshal: %w", err)
	}
	mode := os.FileMode(0o644)
	info, err := os.Stat(request.DescriptorSetOut)
	if err == nil {
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Stat: %w", err)
	}
	if err := fs.WriteAtomicFile(request.DescriptorSetOut, raw, mode); err != nil {
		return fmt.Errorf("WriteAtomicFile: %w", err)
	}
	return nil
}

func v1ModuleDescriptorNames(selected v1GenerationModule) (map[string]bool, error) {
	roots, err := modules.ModuleSources(selected.directory, selected.module)
	if err != nil {
		return nil, fmt.Errorf("ModuleSources: %w", err)
	}
	files := make(map[string]bool)
	for _, root := range roots {
		if err := modules.WalkProtoFiles(root.Path, func(path string) error {
			relative, err := filepath.Rel(root.Path, path)
			if err != nil {
				return fmt.Errorf("Rel: %w", err)
			}
			files[filepath.ToSlash(relative)] = true
			return nil
		}); err != nil {
			return nil, fmt.Errorf("WalkProtoFiles: %w", err)
		}
	}
	return files, nil
}
