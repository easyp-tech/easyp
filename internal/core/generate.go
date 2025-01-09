package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

const defaultCompiler = "protoc"

// Generate generates files.
func (c *Core) Generate(ctx context.Context, root, directory string) error {
	q := Query{
		Compiler: defaultCompiler,
		Imports: []string{
			root,
		},
		Plugins: c.plugins,
	}

	if c.protoRoot != "" {
		q.Imports[0] = c.protoRoot
	}

	for _, dep := range c.deps {
		modulePaths, err := c.getModulePath(ctx, dep)
		if err != nil {
			return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
		}

		q.Imports = append(q.Imports, modulePaths)
	}

	if c.generateOutDirs {
		for _, plug := range q.Plugins {
			if filepath.IsAbs(plug.Out) {
				continue
			}

			err := os.MkdirAll(plug.Out, 0777)
			if err != nil {
				return fmt.Errorf("os.MkdirAll: %w", err)
			}
		}
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

	fsWalker := fs.NewFSWalker(directory, "")
	err := fsWalker.WalkDir(func(path string, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case filepath.Ext(path) != ".proto":
			return nil
		case shouldIgnore(path, c.inputs.Dirs):
			c.logger.DebugContext(ctx, "ignore", slog.String("path", path))

			return nil
		}

		q.Files = append(q.Files, path)

		return nil
	})
	if err != nil {
		return fmt.Errorf("fsWalker.WalkDir: %w", err)
	}

	cmd := q.build()

	slog.DebugContext(ctx, "Run command", "cmd", cmd)

	_, err = c.console.RunCmd(ctx, root, cmd)
	if err != nil {
		return fmt.Errorf("adapters.RunCmd: %w", err)
	}

	return nil
}

func shouldIgnore(path string, dirs []string) bool {
	if len(dirs) == 0 {
		return false
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
