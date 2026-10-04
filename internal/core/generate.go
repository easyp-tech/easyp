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

// GenerationPlan holds a compiled graph with managed options already applied.
// Preparing a plan does not execute plugins or write generated output.
type GenerationPlan struct {
	core            *Core
	root            string
	files           []string
	descriptors     []*descriptorpb.FileDescriptorProto
	dependencyFiles []string
}

// Generate prepares a module, optionally exports its descriptors, and runs plugins.
func (c *Core) Generate(ctx context.Context, root, descriptorSetOut string, includeImports bool) error {
	plan, err := c.PrepareGeneration(ctx, root)
	if err != nil {
		return fmt.Errorf("PrepareGeneration: %w", err)
	}
	if descriptorSetOut != "" {
		data, err := proto.MarshalOptions{Deterministic: true}.Marshal(plan.DescriptorSet(includeImports))
		if err != nil {
			return fmt.Errorf("Marshal: %w", err)
		}
		if err := os.WriteFile(descriptorSetOut, data, 0o644); err != nil {
			return fmt.Errorf("WriteFile: %w", err)
		}
	}
	return plan.Execute(ctx)
}

// PrepareGeneration compiles and transforms a module without invoking plugins.
func (c *Core) PrepareGeneration(ctx context.Context, root string) (*GenerationPlan, error) {
	return c.prepareGeneration(ctx, root, nil)
}

// PrepareGenerationFiles compiles only explicitly selected module source files
// and their complete import closure. Names are relative to module import roots.
// An empty selection is not the same as an implicit all-files selection.
func (c *Core) PrepareGenerationFiles(ctx context.Context, root string, files []string) (*GenerationPlan, error) {
	if len(files) == 0 {
		return nil, ErrEmptyInputFiles
	}
	return c.prepareGeneration(ctx, root, files)
}

func (c *Core) prepareGeneration(ctx context.Context, root string, selected []string) (*GenerationPlan, error) {
	c.logger.Info(ctx, "preparing code generation", slog.String("root", root))
	imports := append([]string{}, c.importRoots...)
	var files []string
	selection := make(map[string]bool, len(selected))
	for _, name := range selected {
		selection[name] = true
	}
	found := make(map[string]bool)

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
			if selected != nil && !selection[addedFile] {
				return nil
			}
			if !found[addedFile] {
				files = append(files, addedFile)
				found[addedFile] = true
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("WalkDir: %w", err)
		}
	}

	for _, name := range selected {
		if !found[name] {
			return nil, fmt.Errorf("selected source %q is outside the module inputs", name)
		}
	}
	c.logger.Debug(ctx, "resolved imports and files", slog.Any("imports", imports), slog.Any("files", files))

	if len(files) == 0 {
		return nil, ErrEmptyInputFiles
	}

	// Search local roots before dependency roots.
	slices.Reverse(imports)

	compiler := protocompile.Compiler{
		Resolver:       wellknownimports.WithStandardImports(&protocompile.SourceResolver{ImportPaths: imports}),
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	compiled, err := compiler.Compile(ctx, files...)
	if err != nil {
		return nil, fmt.Errorf("Compile: %w", err)
	}

	fileDescriptors, dependencyFiles := collectFileDescriptors(compiled)

	fileNames := make([]string, len(fileDescriptors))
	for i, fd := range fileDescriptors {
		fileNames[i] = fd.GetName()
	}
	c.logger.Debug(ctx, "resolved file descriptor order", slog.Int("file_count", len(fileDescriptors)), slog.Any("files", fileNames))

	if c.managedMode.Enabled || c.managedMode.GoPackageOnly {
		c.logger.Debug(ctx, "applying managed mode to file descriptors")
		fileToModule := c.buildFileToModuleMap(files)
		if err := ApplyManagedMode(fileDescriptors, c.managedMode, fileToModule); err != nil {
			return nil, fmt.Errorf("ApplyManagedMode: %w", err)
		}
	}

	return &GenerationPlan{core: c, root: root, files: files,
		descriptors: fileDescriptors, dependencyFiles: dependencyFiles}, nil
}

// DescriptorSet returns an independent copy of the prepared descriptors.
func (p *GenerationPlan) DescriptorSet(includeImports bool) *descriptorpb.FileDescriptorSet {
	set := &descriptorpb.FileDescriptorSet{}
	targets := make(map[string]bool, len(p.files))
	if !includeImports {
		for _, file := range p.files {
			targets[file] = true
		}
	}
	for _, file := range p.descriptors {
		if includeImports || targets[file.GetName()] {
			set.File = append(set.File, proto.Clone(file).(*descriptorpb.FileDescriptorProto))
		}
	}
	return set
}

// Execute invokes plugins using the prepared graph, without compiling it again.
func (p *GenerationPlan) Execute(ctx context.Context) error {
	bucket := NewGenerateBucket()
	if err := p.ExecuteInto(ctx, bucket); err != nil {
		return err
	}
	return bucket.DumpToFs(ctx)
}

// ExecuteInto stages plugin results in a shared bucket, rejecting conflicting outputs.
func (p *GenerationPlan) ExecuteInto(ctx context.Context, filesToWrite *GenerateBucket) error {
	c, root := p.core, p.root
	files, fileDescriptors, dependencyFiles := p.files, p.descriptors, p.dependencyFiles

	for _, plugin := range c.plugins {
		filesToGenerate := slices.Clone(files)

		if plugin.WithImports {
			for _, dependency := range dependencyFiles {
				if !slices.Contains(filesToGenerate, dependency) {
					filesToGenerate = append(filesToGenerate, dependency)
				}
			}
		}

		req := &pluginpb.CodeGeneratorRequest{
			FileToGenerate:  filesToGenerate,
			ProtoFile:       fileDescriptors,
			CompilerVersion: version.CompilerVersion(),
		}

		// Executors receive independent requests; prepared descriptors stay immutable.
		req = proto.Clone(req).(*pluginpb.CodeGeneratorRequest)

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
			outputName := file.GetName()
			if c.goPackageOutputByPackage {
				outputName = goPackageOutputPath(outputName, filesToGenerate, fileDescriptors)
			}
			if !filepath.IsLocal(filepath.FromSlash(outputName)) {
				return fmt.Errorf("plugin returned invalid output path %q", outputName)
			}
			path := filepath.Join(outputDir, filepath.FromSlash(outputName))

			c.logger.Debug(ctx, "generated file",
				slog.String("plugin", source),
				slog.String("file", outputName),
				slog.String("plugin_out", plugin.Out),
				slog.String("full_path", path),
			)

			if err := addFileWithInsertionPoint(ctx, path, file, filesToWrite); err != nil {
				return fmt.Errorf("addFileWithInsertionPoint: %w", err)
			}
		}
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

	if previous, ok := bucket.GetFile(ctx, filePath); ok && !bytes.Equal(previous.Data(), fileContent) {
		return fmt.Errorf("conflicting generated output %q; choose distinct plugin out paths and matching go_package values", filePath)
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

func goPackageOutputPath(name string, filesToGenerate []string, descriptors []*descriptorpb.FileDescriptorProto) string {
	if !strings.HasSuffix(name, ".go") {
		return name
	}
	descriptorByName := make(map[string]*descriptorpb.FileDescriptorProto, len(descriptors))
	for _, descriptor := range descriptors {
		descriptorByName[descriptor.GetName()] = descriptor
	}
	output := filepath.ToSlash(name)
	outputDir, outputBase := filepath.ToSlash(filepath.Dir(output)), filepath.Base(output)
	for _, source := range filesToGenerate {
		descriptor := descriptorByName[source]
		if descriptor == nil || descriptor.GetPackage() == "" {
			continue
		}
		source = filepath.ToSlash(source)
		sourceDir := filepath.ToSlash(filepath.Dir(source))
		sourceBase := strings.TrimSuffix(filepath.Base(source), ".proto")
		if outputDir != sourceDir || (!strings.HasPrefix(outputBase, sourceBase+".") && !strings.HasPrefix(outputBase, sourceBase+"_")) {
			continue
		}
		packageDir := strings.ReplaceAll(descriptor.GetPackage(), ".", "/")
		if packageDir == sourceDir {
			return name
		}
		return filepath.ToSlash(filepath.Join(packageDir, outputBase))
	}
	return name
}
