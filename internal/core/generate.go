package core

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/wellknownimports"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

// Generate generates files.
func (c *Core) Generate(ctx context.Context, root, directory string) error {
	// TODO: call download before
	q := Query{
		Imports: []string{},
		Plugins: c.plugins,
	}

	for lockFileInfo := range c.lockFile.DepsIter() {
		modulePath := c.storage.GetInstallDir(lockFileInfo.Name, lockFileInfo.Version)

		q.Imports = append(q.Imports, modulePath)
	}

	for _, repo := range c.inputs.InputGitRepos {
		gitGenerateCb := func(modulePaths string) func(path string, err error) error {
			return func(path string, err error) error {
				switch {
				case err != nil:
					return err
				case ctx.Err() != nil:
					return ctx.Err()
				case filepath.Ext(path) != ".proto":
					return nil
				}

				addedFile := stripPrefix(path, repo.Root)

				q.Files = append(q.Files, addedFile)
				q.Imports = append(q.Imports, modulePaths)
				if repo.Root != "" {
					q.Imports = append(q.Imports, filepath.Join(modulePaths, repo.Root))
				}

				return nil
			}
		}

		module := models.NewModule(repo.URL)

		lockFileInfo, err := c.lockFile.Read(module.Name)
		if err != nil {
			return fmt.Errorf("c.lockFile.Read: %w", err)
		}

		modulePaths := c.storage.GetInstallDir(lockFileInfo.Name, lockFileInfo.Version)

		fsWalker := fs.NewFSWalker(modulePaths, repo.SubDirectory)
		err = fsWalker.WalkDir(gitGenerateCb(modulePaths))
		if err != nil {
			return fmt.Errorf("fsWalker.WalkDir: %w", err)
		}
	}

	for _, inputFilesDir := range c.inputs.InputFilesDir {
		searchPath := filepath.Join(inputFilesDir.Root, inputFilesDir.Path)
		// Skip if inputFilesDir.Root and directory don't overlap
		if directory != "." && !pathsOverlap(directory, searchPath) {
			c.logger.DebugContext(ctx, "skipping inputFilesDir",
				slog.String("directory", directory),
				slog.String("searchPath", searchPath),
				slog.String("reason", "paths don't overlap"),
			)
			continue
		}

		fsWalker := fs.NewFSWalker(root, searchPath)
		q.Imports = append(q.Imports, inputFilesDir.Root)

		err := fsWalker.WalkDir(func(walkPath string, err error) error {
			switch {
			case err != nil:
				return err
			case ctx.Err() != nil:
				return ctx.Err()
			case filepath.Ext(walkPath) != ".proto":
				return nil
			case shouldIgnore(walkPath, []string{directory}):
				c.logger.DebugContext(ctx, "ignore", slog.String("walkPath", walkPath), slog.String("directory", directory))
				return nil
			}

			// Get file path relative to inputFilesDir.Root
			// walkPath starts with inputFilesDir.Root
			addedFile := stripPrefix(walkPath, inputFilesDir.Root)
			q.Files = append(q.Files, addedFile)

			return nil
		})
		if err != nil {
			return fmt.Errorf("fsWalker.WalkDir: %w", err)
		}
	}

	c.logger.DebugContext(ctx, "data", "import", q.Imports, "files", q.Files)

	if len(q.Files) == 0 {
		return ErrEmptyInputFiles
	}

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
	var addFileWithDeps func(string) error
	addFileWithDeps = func(fileName string) error {
		// If already processed - skip
		if processedFiles[fileName] {
			return nil
		}

		// Compile file
		depRes, err := compiler.Compile(ctx, fileName)
		if err != nil {
			return fmt.Errorf("compile %s: %w", fileName, err)
		}

		if len(depRes) == 0 {
			return fmt.Errorf("no results for %s", fileName)
		}

		descriptor := protoutil.ProtoFromFileDescriptor(depRes[0])

		// IMPORTANT: first recursively add all dependencies
		for _, dep := range descriptor.Dependency {
			if err := addFileWithDeps(dep); err != nil {
				// Ignore errors for optional dependencies
				c.logger.DebugContext(ctx, "Warning: could not compile dependency",
					slog.String("dependency", dep),
					slog.String("error", err.Error()))
			}
		}

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
		descriptor := protoutil.ProtoFromFileDescriptor(file)

		// First add all dependencies of this file
		for _, dep := range descriptor.Dependency {
			if err := addFileWithDeps(dep); err != nil {
				c.logger.DebugContext(ctx, "Warning: could not compile dependency",
					slog.String("dependency", dep),
					slog.String("error", err.Error()))
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
	c.logger.DebugContext(ctx, "File order in request:")
	for i, fd := range fileDescriptors {
		c.logger.DebugContext(ctx, fmt.Sprintf("%d: %s", i, fd.GetName()))
	}

	// Build file to module mapping for managed mode
	fileToModule := c.buildFileToModuleMap(q.Files)

	// Apply managed mode to file descriptors
	if c.managedMode.Enabled {
		c.logger.DebugContext(ctx, "Applying managed mode to file descriptors")
		if err := ApplyManagedMode(fileDescriptors, c.managedMode, fileToModule); err != nil {
			return fmt.Errorf("ApplyManagedMode: %w", err)
		}
	}

	filesToWrite := NewGenerateBucket()

	for _, plugin := range c.plugins {
		filesToGenerate := q.Files

		if plugin.WithImports {
			filesToGenerate = append(filesToGenerate, dependencyFiles...)
		}

		req := &pluginpb.CodeGeneratorRequest{
			FileToGenerate: filesToGenerate,
			ProtoFile:      fileDescriptors,
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

			c.logger.DebugContext(ctx, "generated file",
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

// pathsOverlap checks if two paths overlap (one is within another or they are equal).
// It is recommended to pass absolute paths.
func pathsOverlap(a, b string) bool {
	na := filepath.Clean(a)
	nb := filepath.Clean(b)

	// Full match is always overlap
	if na == nb {
		return true
	}

	// Add separator at the end to distinguish "/foo/bar" from "/foo/bark"
	naSlash := na + string(filepath.Separator)
	nbSlash := nb + string(filepath.Separator)

	// na is parent of nb
	if strings.HasPrefix(nbSlash, naSlash) {
		return true
	}

	// nb is parent of na
	if strings.HasPrefix(naSlash, nbSlash) {
		return true
	}

	return false
}

func shouldIgnore(path string, dirs []string) bool {
	path = filepath.Clean(path)
	if len(dirs) == 0 {
		return true
	}

	for _, dir := range dirs {
		dir = filepath.Clean(dir)

		// Special case: if dir is ".", match everything
		if dir == "." {
			slog.Debug("shouldIgnore: dir is '.', matching all paths", "path", path)
			return false // Don't ignore - match everything
		}

		// Check if path starts with dir (prefix matching)
		if strings.HasPrefix(path, dir+"/") || path == dir {
			slog.Debug("shouldIgnore: path starts with dir", "path", path, "dir", dir)
			return false // Don't ignore - path is within directory
		}

		// Check regex pattern (for wildcard patterns)
		// QuoteMeta escapes all special chars (including *), then we convert \* back to .* for wildcard matching
		pattern := regexp.QuoteMeta(dir)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		regexPattern := "^" + pattern

		matched, err := regexp.MatchString(regexPattern, path)
		if err != nil {
			slog.Warn("shouldIgnore: regex match error", "path", path, "dir", dir, "regex", regexPattern, "error", err)
			continue
		}
		if matched {
			slog.Debug("shouldIgnore: path matches regex pattern", "path", path, "dir", dir, "regex", regexPattern)
			return false // Don't ignore - path matches pattern
		}
	}

	return true
}

func stripPrefix(path, prefix string) string {
	normalizedPath := filepath.ToSlash(path)
	normalizedPrefix := filepath.ToSlash(prefix)
	// Remove trailing slash from prefix if present
	normalizedPrefix = strings.TrimSuffix(normalizedPrefix, "/")

	return strings.TrimPrefix(normalizedPath, normalizedPrefix+"/")
}

// For debuging
func runCmd(ctx context.Context, dir string, command string, stdIn *bytes.Buffer, commandParams ...string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	fullCommand := append([]string{command}, commandParams...)
	cmd := exec.CommandContext(ctx, "bash", "-c", strings.Join(fullCommand, " "))
	cmd.Dir = dir
	cmd.Stdin = stdIn
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s", stderr.String())
	}

	return stdout.String(), nil
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

// buildFileToModuleMap creates a mapping from file paths to their module names.
// This is used by managed mode to apply module-specific rules.
//
// The mapping works by scanning installed dependency directories and mapping
// relative proto file paths to their source module. For example:
//   - Module "github.com/googleapis/googleapis" installed at ~/.easyp/mod/github.com/googleapis/googleapis/v1/
//   - Contains file: google/api/annotations.proto
//   - Mapping: "google/api/annotations.proto" → "github.com/googleapis/googleapis"
func (c *Core) buildFileToModuleMap(files []string) map[string]string {
	fileToModule := make(map[string]string)

	// Map main files - they belong to the local project (empty module)
	for _, file := range files {
		fileToModule[file] = ""
	}

	// Build mapping from dependency install directories
	// For each dependency, scan its install dir and map relative paths to module name
	for _, dep := range c.deps {
		module := models.NewModule(dep)
		c.mapModuleFiles(module.Name, fileToModule)
	}

	// Also map files from git repo inputs
	for _, repo := range c.inputs.InputGitRepos {
		module := models.NewModule(repo.URL)
		c.mapModuleFiles(module.Name, fileToModule)
	}

	return fileToModule
}

// mapModuleFiles scans a module's install directory and adds proto file mappings.
func (c *Core) mapModuleFiles(moduleName string, fileToModule map[string]string) {
	// Get module version from lock file
	lockInfo, err := c.lockFile.Read(moduleName)
	if err != nil {
		// Module not installed or not in lock file - skip
		return
	}

	// Get install directory
	installDir := c.storage.GetInstallDir(moduleName, lockInfo.Version)

	// Walk the install directory and map all .proto files
	err = filepath.WalkDir(installDir, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if d.IsDir() || filepath.Ext(filePath) != ".proto" {
			return nil
		}

		// Get relative path from install dir (this is the import path)
		relPath, err := filepath.Rel(installDir, filePath)
		if err != nil {
			return nil
		}

		// Normalize to forward slashes (proto import paths use forward slashes)
		relPath = filepath.ToSlash(relPath)

		// Map this file to its module
		fileToModule[relPath] = moduleName

		return nil
	})

	if err != nil {
		// Log error but don't fail - managed mode can work without module mapping
		c.logger.Debug("Failed to scan module directory",
			"module", moduleName,
			"installDir", installDir,
			"error", err)
	}
}
