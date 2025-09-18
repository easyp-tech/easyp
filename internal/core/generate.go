package core

import (
	"bytes"
	"context"
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
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

const defaultCompiler = "protoc"

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

		options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
			if v == "" {
				return k
			}

			return k + "=" + v
		})

		req := &pluginpb.CodeGeneratorRequest{
			FileToGenerate: q.Files,
			ProtoFile:      fileDescriptors,
			Parameter:      proto.String(strings.Join(options, ",")),
		}

		stdIn := &bytes.Buffer{}
		b, err := proto.Marshal(req)
		if err != nil {
			return fmt.Errorf("proto.Marshal: %w", err)
		}

		_, err = stdIn.Write(b)
		if err != nil {
			return fmt.Errorf("stdIn.Write: %w", err)
		}

		stdout, err := c.console.RunCmdWithStdin(ctx, root, stdIn, fmt.Sprintf("protoc-gen-%s", plugin.Name))
		if err != nil {
			return fmt.Errorf("runCmd: %w", err)
		}

		// Парсим ответ от плагина
		var resp pluginpb.CodeGeneratorResponse
		if err := proto.Unmarshal([]byte(stdout), &resp); err != nil {
			return fmt.Errorf("proto.Unmarshal response: %w", err)
		}

		logData, err := protojson.Marshal(&resp)
		if err != nil {
			slog.Error("ATTENTION: Could not marshal response to JSON")
		} else {
			fmt.Printf("\n\n\n%s\n\n\n", string(logData))
		}

		// Проверяем на ошибки от плагина
		if resp.Error != nil {
			return fmt.Errorf("plugin error: %s", *resp.Error)
		}

		// Выводим информацию о сгенерированных файлах (для отладки)
		for _, file := range resp.File {

			p := filepath.Join(directory, *file.Name)

			c.logger.DebugContext(ctx, "generated file",
				slog.String("plugin", plugin.Name),
				slog.String("file", *file.Name),
				slog.String("full_path", p),
			)

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
		return "", fmt.Errorf(stderr.String())
	}

	return stdout.String(), nil
}
