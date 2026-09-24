package core

import (
	"bytes"
	"context"
	"fmt"
	stdfs "io/fs"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/wellknownimports"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/version"
)

// Generate generates code using source and import roots resolved by the module layer.
func (c *Core) Generate(ctx context.Context, root, descriptorSetOut string, includeImports bool) error {
	c.logger.Info(ctx, "starting code generation", slog.String("root", root))

	imports := append([]string{}, c.importRoots...)
	var files []string

	for _, inputFilesDir := range c.inputs.InputFilesDir {
		searchPath := filepath.Join(inputFilesDir.Root, inputFilesDir.Path)
		fsWalker := fs.NewFSWalker(root, searchPath)
		importRoot := filepath.Join(root, inputFilesDir.Root)
		if !slices.Contains(imports, importRoot) {
			imports = append(imports, importRoot)
		}

		err := fsWalker.WalkDir(func(walkPath string, err error) error {
			switch {
			case err != nil:
				return err
			case ctx.Err() != nil:
				return ctx.Err()
			case path_helpers.ShouldSkipV1SourceDir(importRoot, filepath.Join(root, walkPath)):
				return stdfs.SkipDir
			case filepath.Ext(walkPath) != ".proto":
				return nil
			}

			// Convert to relative path matching proto import format
			addedFile := stripPrefix(walkPath, inputFilesDir.Root)
			files = append(files, addedFile)

			return nil
		})
		if err != nil {
			return fmt.Errorf("WalkDir: %w", err)
		}
	}

	c.logger.Debug(ctx, "resolved imports and files", slog.Any("imports", imports), slog.Any("files", files))

	if len(files) == 0 {
		return ErrEmptyInputFiles
	}

	// Search local roots before dependency roots.
	slices.Reverse(imports)

	compiler := protocompile.Compiler{
		Resolver:       wellknownimports.WithStandardImports(&protocompile.SourceResolver{ImportPaths: imports}),
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	compiled, err := compiler.Compile(ctx, files...)
	if err != nil {
		return fmt.Errorf("Compile: %w", err)
	}

	fileDescriptors, dependencyFiles := collectFileDescriptors(compiled)

	fileNames := make([]string, len(fileDescriptors))
	for i, fd := range fileDescriptors {
		fileNames[i] = fd.GetName()
	}
	c.logger.Debug(ctx, "resolved file descriptor order", slog.Int("file_count", len(fileDescriptors)), slog.Any("files", fileNames))

	if c.managedMode.Enabled {
		c.logger.Debug(ctx, "applying managed mode to file descriptors")
		fileToModule := c.buildFileToModuleMap(files)
		if err := ApplyManagedMode(fileDescriptors, c.managedMode, fileToModule); err != nil {
			return fmt.Errorf("ApplyManagedMode: %w", err)
		}
	}

	if descriptorSetOut != "" {
		descriptorsToSave := fileDescriptors
		if !includeImports {
			descriptorsToSave = nil
			targetFiles := make(map[string]bool)
			for _, f := range files {
				targetFiles[f] = true
			}
			for _, fd := range fileDescriptors {
				if targetFiles[fd.GetName()] {
					descriptorsToSave = append(descriptorsToSave, fd)
				}
			}
		}

		descriptorSet := &descriptorpb.FileDescriptorSet{
			File: descriptorsToSave,
		}

		data, err := proto.MarshalOptions{Deterministic: true}.Marshal(descriptorSet)
		if err != nil {
			return fmt.Errorf("Marshal: %w", err)
		}

		if err := os.WriteFile(descriptorSetOut, data, 0o644); err != nil {
			return fmt.Errorf("WriteFile: %w", err)
		}
	}

	filesToWrite := NewGenerateBucket()

	for _, plugin := range c.plugins {
		filesToGenerate := files

		if plugin.WithImports {
			filesToGenerate = append(filesToGenerate, dependencyFiles...)
		}

		req := &pluginpb.CodeGeneratorRequest{
			FileToGenerate:  filesToGenerate,
			ProtoFile:       fileDescriptors,
			CompilerVersion: version.CompilerVersion(),
		}

		executor := c.getExecutor(plugin)

		source := plugin.Source.Name
		if plugin.Source.Remote != "" {
			source = plugin.Source.Remote
		}

		if plugin.Source.Path != "" {
			source = plugin.Source.Path
		}

		resp, err := executor.Execute(ctx, pluginexecutor.Info{
			Source:  source,
			WorkDir: c.pluginWorkDir,
			Command: plugin.Source.Command,
			Options: plugin.Options,
		}, req)
		if err != nil {
			return fmt.Errorf("execute plugin %s: %w, executor: %s", source, err, executor.GetName())
		}

		if resp.Error != nil {
			return fmt.Errorf("plugin %s error: %s, executor: %s", plugin.Source, *resp.Error, executor.GetName())
		}

		outputDir := root
		if plugin.Out != "" {
			outputDir = filepath.Join(root, plugin.Out)
		}
		for _, file := range resp.File {
			path := filepath.Join(outputDir, file.GetName())

			c.logger.Debug(ctx, "generated file",
				slog.String("plugin", source),
				slog.String("file", file.GetName()),
				slog.String("plugin_out", plugin.Out),
				slog.String("full_path", path),
			)

			if err := addFileWithInsertionPoint(ctx, path, file, filesToWrite); err != nil {
				return fmt.Errorf("addFileWithInsertionPoint: %w", err)
			}
		}
	}

	err = filesToWrite.DumpToFs(ctx)
	if err != nil {
		return fmt.Errorf("DumpToFs: %w", err)
	}

	c.logger.Info(ctx, "code generation completed")

	return nil
}

// collectFileDescriptors orders each descriptor after its imports, preserving
// the compiler's target order and recording files first reached as dependencies.
func collectFileDescriptors(files linker.Files) ([]*descriptorpb.FileDescriptorProto, []string) {
	var descriptors []*descriptorpb.FileDescriptorProto
	var dependencies []string
	processed := make(map[string]bool)

	var visit func(protoreflect.FileDescriptor, bool)
	visit = func(file protoreflect.FileDescriptor, dependency bool) {
		name := file.Path()
		if processed[name] {
			return
		}
		for i := range file.Imports().Len() {
			visit(file.Imports().Get(i), true)
		}
		descriptors = append(descriptors, protoutil.ProtoFromFileDescriptor(file))
		processed[name] = true
		if dependency {
			dependencies = append(dependencies, name)
		}
	}

	for _, file := range files {
		visit(file, false)
	}
	return descriptors, dependencies
}

// addFileWithInsertionPoint add file to bucket with insertion point support
// inspired by https://github.com/bufbuild/buf/blob/v1.60.0/private/bufpkg/bufprotoplugin/response_writer.go#L75
func addFileWithInsertionPoint(
	ctx context.Context,
	filePath string,
	file *pluginpb.CodeGeneratorResponse_File,
	bucket *GenerateBucket,
) error {
	// Write file to bucket with insertion point support
	fileContent := make([]byte, 0)
	if file.Content != nil {
		fileContent = []byte(*file.Content)
	}
	if insertionPoint := file.GetInsertionPoint(); insertionPoint != "" {
		// If insertion point is present, find existing file in bucket
		// This mechanism may be broken if plugins are executed in a different order
		// inspired by https://github.com/bufbuild/buf/blob/v1.60.0/private/pkg/storage/storagemem/bucket.go#L144
		existsFile, ok := bucket.GetFile(ctx, filePath)
		if !ok || len(existsFile.Data()) == 0 {
			return fmt.Errorf("file not found (bucket): %s", filePath)
		}

		newFileContent, err := writeInsertionPoint(
			ctx,
			file,
			bytes.NewReader(existsFile.Data()),
		)
		if err != nil {
			return fmt.Errorf("writeInsertionPoint: %w", err)
		}

		bucket.PutFile(ctx, filePath, newFileContent)
		return nil
	}

	bucket.PutFile(ctx, filePath, fileContent)
	return nil
}

// stripPrefix removes prefix from path and normalizes to forward slashes.
func stripPrefix(path, prefix string) string {
	normalizedPath := filepath.ToSlash(path)
	normalizedPrefix := filepath.ToSlash(filepath.Clean(prefix))
	// Remove trailing slash from prefix if present
	normalizedPrefix = strings.TrimSuffix(normalizedPrefix, "/")

	return strings.TrimPrefix(normalizedPath, normalizedPrefix+"/")
}

// isPluginInPath checks if the plugin is available in PATH
func (c *Core) isPluginInPath(pluginName string) bool {
	pluginCmd := fmt.Sprintf("protoc-gen-%s", pluginName)
	_, err := exec.LookPath(pluginCmd)
	return err == nil
}

func (c *Core) getExecutor(plugin Plugin) pluginexecutor.Executor {
	// Priority 1: If command is specified, use command executor
	if len(plugin.Source.Command) > 0 {
		return c.commandExecutor
	}

	// Priority 2: If remote URL is specified, use remote executor
	if plugin.Source.Remote != "" {
		return c.remoteExecutor
	}

	// Priority 3: If plugin is builtin and not found in PATH, use builtin executor
	if pluginexecutor.IsBuiltinPlugin(plugin.Source.Name) && !c.isPluginInPath(plugin.Source.Name) {
		return c.builtinExecutor
	}

	// Priority 4: Otherwise use local executor (backward compatibility)
	return c.localExecutor
}

// buildFileToModuleMap maps generated files to their v1 module identities.
func (c *Core) buildFileToModuleMap(files []string) map[string]string {
	fileToModule := make(map[string]string, len(files)+len(c.fileModules))
	for _, file := range files {
		fileToModule[file] = ""
	}
	maps.Copy(fileToModule, c.fileModules)
	return fileToModule
}
