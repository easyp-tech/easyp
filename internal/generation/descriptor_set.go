package generation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

// generateV1DescriptorSet collects each module's descriptors before replacing
// the requested output. Shared imports are included once; conflicting files
// with the same protobuf import path cannot form a valid descriptor set.
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
	for _, target := range targets {
		moduleRequest := request
		moduleRequest.DescriptorSetOut = temporaryPath
		if err := generateSelectedV1Module(ctx, log, cache, moduleRequest, target.configPath, request.WorkDir, target.module, target.config); err != nil {
			return fmt.Errorf("generateSelectedV1Module: %w", err)
		}
		raw, err := os.ReadFile(temporaryPath)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		var selected descriptorpb.FileDescriptorSet
		if err := proto.Unmarshal(raw, &selected); err != nil {
			return fmt.Errorf("Unmarshal: %w", err)
		}
		for _, file := range selected.File {
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

	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(combined)
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
