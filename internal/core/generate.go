package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
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

const (
	defaultCompiler          = "protoc"
	averageGeneratedFileSize = 15 * 1024
)

// Generate generates files.
func (c *Core) Generate(ctx context.Context, root, directory string) error {
	q := Query{
		Compiler: defaultCompiler,
		Imports:  []string{},
		Plugins:  c.plugins,
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

				q.Files = append(q.Files, path)
				q.Imports = append(q.Imports, modulePaths)

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

	compiler := protocompile.Compiler{
		Resolver: protocompile.CompositeResolver{
			wellknownimports.WithStandardImports(
				&protocompile.SourceResolver{
					ImportPaths: append(q.Imports),
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

	filesToWrite := NewGenerateBucket()

	for _, plugin := range c.plugins {
		filesToGenerate := q.Files

		if plugin.WithImports {
			filesToGenerate = append(filesToGenerate, dependencyFiles...)
		}
		// Создаем запрос для плагина
		req := &pluginpb.CodeGeneratorRequest{
			FileToGenerate: filesToGenerate,
			ProtoFile:      fileDescriptors,
		}

		// Получаем подходящий executor для плагина
		executor := c.getExecutor(plugin)

		// Выполняем плагин
		resp, err := executor.Execute(ctx, pluginexecutor.Info{
			Name:    plugin.Name,
			Options: plugin.Options,
			URL:     plugin.URL,
		}, req)
		if err != nil {
			return fmt.Errorf("execute plugin %s: %w", plugin.Name, err)
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
				slog.String("plugin", plugin.Name),
				slog.String("file", file.GetName()),
				slog.String("plugin_out", plugin.Out),
				slog.String("full_path", p),
			)

			// Пишем файл в bucket с поддержкой insertion point
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

// addFileWithInsertionPoint добавляет файл в bucket с поддержкой insertion point
// inspired by https://github.com/bufbuild/buf/blob/v1.60.0/private/bufpkg/bufprotoplugin/response_writer.go#L75
func addFileWithInsertionPoint(
	ctx context.Context,
	filePath string,
	file *pluginpb.CodeGeneratorResponse_File,
	bucket *GenerateBucket,
) error {
	// Пишем файл даже если пустой
	fileContent := make([]byte, 0)
	if file.Content != nil {
		fileContent = []byte(*file.Content)
	}
	if insertionPoint := file.GetInsertionPoint(); insertionPoint != "" {
		// Если есть insertion point есть в file нужно найти уже имеющийся файл с таким же path в bucket
		// Этот механизм может быть сломан изменённым порядком выполнения плагинов
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
	if plugin.URL != "" {
		return c.remoteExecutor
	}

	return c.localExecutor
}
