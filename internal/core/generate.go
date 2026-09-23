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

// SetImportRoots supplies import paths resolved from module dependencies.
func (c *Core) SetImportRoots(roots []string) {
	c.importRoots = append([]string(nil), roots...)
}

// SetFileModules supplies module identities for managed-mode selectors.
func (c *Core) SetFileModules(modules map[string]string) {
	c.fileModules = maps.Clone(modules)
}

// Generate generates code using source and import roots resolved by the module layer.
func (c *Core) Generate(ctx context.Context, root, descriptorSetOut string, includeImports bool) error {
	c.logger.Info(ctx, "starting code generation", slog.String("root", root))

	q := Query{
		Imports: append([]string{}, c.importRoots...),
		Plugins: c.plugins,
	}

	for _, inputFilesDir := range c.inputs.InputFilesDir {
		searchPath := filepath.Join(inputFilesDir.Root, inputFilesDir.Path)
		fsWalker := fs.NewFSWalker(root, searchPath)
		importRoot := filepath.Join(root, inputFilesDir.Root)
		if !slices.Contains(q.Imports, importRoot) {
			q.Imports = append(q.Imports, importRoot)
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
			q.Files = append(q.Files, addedFile)

			return nil
		})
		if err != nil {
			return fmt.Errorf("fsWalker.WalkDir: %w", err)
		}
	}

	c.logger.Debug(ctx, "resolved imports and files", slog.Any("imports", q.Imports), slog.Any("files", q.Files))

	if len(q.Files) == 0 {
		return ErrEmptyInputFiles
	}

	slices.Reverse(q.Imports) // local first, dependencies last

	compiler := protocompile.Compiler{
		Resolver: protocompile.CompositeResolver{
			wellknownimports.WithStandardImports(
				&protocompile.SourceResolver{
					ImportPaths: q.Imports,
				},
			),
		},
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	res, err := compiler.Compile(ctx, q.Files...)
	if err != nil {
		return fmt.Errorf("compiler.Compile: %w", err)
	}

	// Use slice to preserve correct order
	var fileDescriptors []*descriptorpb.FileDescriptorProto
	processedFiles := make(map[string]bool)
	dependencyFiles := make([]string, 0)

	// Recursive function to add file and its dependencies in correct order
	var addFileWithDeps func(protoreflect.FileDescriptor) error
	addFileWithDeps = func(file protoreflect.FileDescriptor) error {
		fileName := file.Path()
		// If already processed - skip
		if processedFiles[fileName] {
			return nil
		}

		// IMPORTANT: first recursively add all dependencies
		for i := range file.Imports().Len() {
			dep := file.Imports().Get(i)

			if err := addFileWithDeps(dep); err != nil {
				// Ignore errors for optional dependencies
				c.logger.Warn(ctx, "could not compile dependency",
					slog.String("dependency", dep.Path()),
					slog.Any("error", err))
			}
		}

		descriptor := protoutil.ProtoFromFileDescriptor(file)

		// Only after dependencies add the file itself (if not already added)
		if !processedFiles[fileName] {
			fileDescriptors = append(fileDescriptors, descriptor)
			processedFiles[fileName] = true
			dependencyFiles = append(dependencyFiles, fileName)
		}

		return nil
	}

	// Process all files and their dependencies
	for _, file := range res {
		reflectFd := file.(protoreflect.FileDescriptor)
		descriptor := protoutil.ProtoFromFileDescriptor(file)

		// First add all dependencies of this file
		for i := range reflectFd.Imports().Len() {
			dep := reflectFd.Imports().Get(i)

			if err := addFileWithDeps(dep); err != nil {
				c.logger.Warn(ctx, "could not compile dependency",
					slog.String("dependency", dep.Path()),
					slog.Any("error", err))
			}
		}

		// Then add the file itself (if not already added)
		fileName := descriptor.GetName()
		if !processedFiles[fileName] {
			fileDescriptors = append(fileDescriptors, descriptor)
			processedFiles[fileName] = true
		}
	}

	// Log file order for debugging
	fileNames := make([]string, len(fileDescriptors))
	for i, fd := range fileDescriptors {
		fileNames[i] = fd.GetName()
	}
	c.logger.Debug(ctx, "resolved file descriptor order", slog.Int("file_count", len(fileDescriptors)), slog.Any("files", fileNames))

	// Build file to module mapping for managed mode
	fileToModule := c.buildFileToModuleMap(q.Files)

	// Apply managed mode to file descriptors
	if c.managedMode.Enabled {
		c.logger.Debug(ctx, "applying managed mode to file descriptors")
		if err := ApplyManagedMode(fileDescriptors, c.managedMode, fileToModule); err != nil {
			return fmt.Errorf("ApplyManagedMode: %w", err)
		}
	}

	if descriptorSetOut != "" {
		var descriptorsToSave []*descriptorpb.FileDescriptorProto
		if includeImports {
			descriptorsToSave = fileDescriptors
		} else {
			// Filter out imports, keep only target files
			targetFiles := make(map[string]bool)
			for _, f := range q.Files {
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
			return fmt.Errorf("proto.Marshal: %w", err)
		}

		if err := os.WriteFile(descriptorSetOut, data, 0644); err != nil {
			return fmt.Errorf("os.WriteFile: %w", err)
		}
	}

	filesToWrite := NewGenerateBucket()

	for _, plugin := range c.plugins {
		filesToGenerate := q.Files

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
			Command: plugin.Source.Command,
			Options: plugin.Options,
		}, req)
		if err != nil {
			return fmt.Errorf("execute plugin %s: %w, executor: %s", source, err, executor.GetName())
		}

		// Check for plugin errors
		if resp.Error != nil {
			return fmt.Errorf("plugin %s error: %s, executor: %s", plugin.Source, *resp.Error, executor.GetName())
		}

		// Output information about generated files (for debugging)
		for _, file := range resp.File {
			// Determine base directory for output files considering plugin.Out
			var baseDir string
			if plugin.Out != "" {
				baseDir = filepath.Join(root, plugin.Out)
			} else {
				baseDir = root
			}

			p := filepath.Join(baseDir, file.GetName())

			c.logger.Debug(ctx, "generated file",
				slog.String("plugin", source),
				slog.String("file", file.GetName()),
				slog.String("plugin_out", plugin.Out),
				slog.String("full_path", p),
			)

			// Write file to bucket with insertion point support
			if err := addFileWithInsertionPoint(ctx, p, file, filesToWrite); err != nil {
				return fmt.Errorf("addFileWithInsertionPoint: %w", err)
			}
		}
	}

	err = filesToWrite.DumpToFs(ctx)
	if err != nil {
		return fmt.Errorf("filesToWrite.DumpToFs: %w", err)
	}

	c.logger.Info(ctx, "code generation completed")

	return nil
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
