package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
	q := Query{
		Imports: []string{},
		Plugins: c.plugins,
	}

	for _, dep := range c.deps {
		modulePaths, err := c.getModulePath(ctx, dep)
		if err != nil {
			return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
		}

		q.Imports = append(q.Imports, modulePaths)
	}

	for _, repo := range c.inputs.InputGitRepos {
		module := models.NewModule(repo.URL)

		isInstalled, err := c.storage.IsModuleInstalled(module)
		if err != nil {
			return fmt.Errorf("c.isModuleInstalled: %w", err)
		}

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

		if isInstalled {
			modulePaths, err := c.getModulePath(ctx, module.Name)
			if err != nil {
				return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
			}

			fsWalker := fs.NewFSWalker(modulePaths, repo.SubDirectory)

			err = fsWalker.WalkDir(gitGenerateCb(modulePaths))
			if err != nil {
				return fmt.Errorf("fsWalker.WalkDir1: %w", err)
			}

			continue
		}

		err = c.Get(ctx, module)
		if err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				slog.Error("Version not found", "dependency", module.Name, "version", module.Version)

				return fmt.Errorf("models.ErrVersionNotFound: %w", err)
			}

			return fmt.Errorf("c.Get: %w", err)
		}

		modulePaths, err := c.getModulePath(ctx, module.Name)
		if err != nil {
			return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
		}

		fsWalker := fs.NewFSWalker(modulePaths, repo.SubDirectory)
		err = fsWalker.WalkDir(gitGenerateCb(modulePaths))
		if err != nil {
			return fmt.Errorf("fsWalker.WalkDir: %w", err)
		}
	}

	for _, inputFilesDir := range c.inputs.InputFilesDir {
		fsWalker := fs.NewFSWalker(directory, inputFilesDir.Root)
		q.Imports = append(q.Imports, inputFilesDir.Root)

		err := fsWalker.WalkDir(func(walkPath string, err error) error {
			switch {
			case err != nil:
				return err
			case ctx.Err() != nil:
				return ctx.Err()
			case filepath.Ext(walkPath) != ".proto":
				return nil
			case shouldIgnore(walkPath, []string{path.Join(inputFilesDir.Root, inputFilesDir.Path)}):
				c.logger.DebugContext(ctx, "ignore", slog.String("walkPath", walkPath))

				return nil
			}

			addedFile := stripPrefix(walkPath, inputFilesDir.Root)
			q.Files = append(q.Files, addedFile)

			return nil
		})
		if err != nil {
			return fmt.Errorf("fsWalker.WalkDir: %w", err)
		}
	}

	//q.Imports = append(q.Imports, "/tmp/easyp/mod/github.com/svgrafov/test-proto/v0.0.0-20251031101008-f12c7ed5e68548f8f49883482b48a1f70d39f626/proto")

	c.logger.DebugContext(ctx, "data", "import", q.Imports, "files", q.Files)

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

	// Используем slice для сохранения правильного порядка
	var fileDescriptors []*descriptorpb.FileDescriptorProto
	processedFiles := make(map[string]bool)
	dependencyFiles := make([]string, 0)

	// Рекурсивная функция для добавления файла и его зависимостей в правильном порядке
	var addFileWithDeps func(string) error
	addFileWithDeps = func(fileName string) error {
		// Если уже обработали - пропускаем
		if processedFiles[fileName] {
			return nil
		}

		// Компилируем файл
		depRes, err := compiler.Compile(ctx, fileName)
		if err != nil {
			return fmt.Errorf("compile %s: %w", fileName, err)
		}

		if len(depRes) == 0 {
			return fmt.Errorf("no results for %s", fileName)
		}

		descriptor := protoutil.ProtoFromFileDescriptor(depRes[0])

		// ВАЖНО: сначала рекурсивно добавляем все зависимости
		for _, dep := range descriptor.Dependency {
			if err := addFileWithDeps(dep); err != nil {
				// Игнорируем ошибки для опциональных зависимостей
				c.logger.DebugContext(ctx, "Warning: could not compile dependency",
					slog.String("dependency", dep),
					slog.String("error", err.Error()))
			}
		}

		// Только после зависимостей добавляем сам файл (если еще не добавлен)
		if !processedFiles[fileName] {
			fileDescriptors = append(fileDescriptors, descriptor)
			processedFiles[fileName] = true
			dependencyFiles = append(dependencyFiles, fileName)
		}

		return nil
	}

	// Обрабатываем все файлы и их зависимости
	for _, file := range res {
		descriptor := protoutil.ProtoFromFileDescriptor(file)

		// Сначала добавляем все зависимости этого файла
		for _, dep := range descriptor.Dependency {
			if err := addFileWithDeps(dep); err != nil {
				c.logger.DebugContext(ctx, "Warning: could not compile dependency",
					slog.String("dependency", dep),
					slog.String("error", err.Error()))
			}
		}

		// Потом добавляем сам файл (если еще не добавлен)
		fileName := descriptor.GetName()
		if !processedFiles[fileName] {
			fileDescriptors = append(fileDescriptors, descriptor)
			processedFiles[fileName] = true
		}
	}

	// Логируем порядок файлов для отладки
	c.logger.DebugContext(ctx, "File order in request:")
	for i, fd := range fileDescriptors {
		c.logger.DebugContext(ctx, fmt.Sprintf("%d: %s", i, fd.GetName()))
	}

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
			Options: plugin.Options,
		}, req)
		if err != nil {
			return fmt.Errorf("execute plugin %s: %w", source, err)
		}

		// Проверяем на ошибки от плагина
		if resp.Error != nil {
			return fmt.Errorf("plugin error: %s", *resp.Error)
		}

		// Выводим информацию о сгенерированных файлах (для отладки)
		for _, file := range resp.File {
			// Определяем базовую директорию для вывода файлов с учетом plugin.Out
			var baseDir string
			if plugin.Out != "" {
				baseDir = filepath.Join(directory, plugin.Out)
			} else {
				baseDir = directory
			}

			p := filepath.Join(baseDir, file.GetName())

			c.logger.DebugContext(ctx, "generated file",
				slog.String("plugin", source),
				slog.String("file", file.GetName()),
				slog.String("plugin_out", plugin.Out),
				slog.String("full_path", p),
			)

			// Проверяем и создаем директорию, если она не существует (кроссплатформенно)
			dir := filepath.Dir(p)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("os.MkdirAll %s: %w", dir, err)
			}

			f, err := os.Create(p)
			if err != nil {
				return fmt.Errorf("os.Create: %w", err)
			}

			if file.Content == nil {
				continue
			}

			_, err = f.WriteString(*file.Content)
			if err != nil {
				return fmt.Errorf("f.WriteString: %w", err)
			}
		}
	}

	return nil
}

func shouldIgnore(path string, dirs []string) bool {
	if len(dirs) == 0 {
		return true
	}

	for _, dir := range dirs {
		if strings.HasPrefix(path, dir) {
			return false
		}
	}

	return true
}

func (c *Core) getModulePath(ctx context.Context, requestedDependency string) (string, error) {
	module := models.NewModule(requestedDependency)

	isInstalled, err := c.storage.IsModuleInstalled(module)
	if err != nil {
		return "", fmt.Errorf("h.storage.IsModuleInstalled: %w", err)
	}

	if !isInstalled {
		if err := c.Get(ctx, module); err != nil {
			return "", fmt.Errorf("h.mod.Get: %w", err)
		}
	}

	lockFileInfo, err := c.lockFile.Read(module.Name)
	if err != nil {
		return "", fmt.Errorf("lockFile.Read: %w", err)
	}

	installedPath := c.storage.GetInstallDir(module.Name, lockFileInfo.Version)

	return installedPath, nil
}

func stripPrefix(path, prefix string) string {
	normalizedPath := filepath.ToSlash(path)
	normalizedPrefix := filepath.ToSlash(prefix)

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

func (c *Core) getExecutor(plugin Plugin) pluginexecutor.Executor {
	if plugin.Source.Remote != "" {
		return c.remoteExecutor
	}

	return c.localExecutor
}
