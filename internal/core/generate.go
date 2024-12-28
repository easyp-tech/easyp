package core

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/models"
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

	for _, dep := range c.deps {
		modulePaths, err := c.getModulePath(ctx, dep)
		if err != nil {
			return fmt.Errorf("g.moduleReflect.GetModulePath: %w", err)
		}

		q.Imports = append(q.Imports, modulePaths)
	}

	for _, repo := range c.inputs.InputGitRepos {
		module := models.NewModule(fmt.Sprintf("%s@%s", repo.URL, repo.Tag))

		isInstalled, err := c.storage.IsModuleInstalled(module)
		if err != nil {
			return fmt.Errorf("c.isModuleInstalled: %w", err)
		}

		if isInstalled {
			slog.Info("Module is installed", "name", module.Name, "version", module.Version)
			continue
		}

		err = c.Get(ctx, module)
		if err != nil {
			if errors.Is(err, models.ErrVersionNotFound) {
				slog.Error("Version not found", "dependency", module.Name)

				return models.ErrVersionNotFound
			}

			return fmt.Errorf("c.Get: %w", err)
		}
	}

	//module := models.NewModule(dependency)
	//
	//isInstalled, err := c.storage.IsModuleInstalled(module)
	//if err != nil {
	//	return fmt.Errorf("c.isModuleInstalled: %w", err)
	//}
	//
	//if isInstalled {
	//	slog.Info("Module is installed", "name", module.Name, "version", module.Version)
	//	continue
	//}
	//
	//if err := c.Get(ctx, module); err != nil {
	//	if errors.Is(err, models.ErrVersionNotFound) {
	//		slog.Error("Version not found", "dependency", dependency)
	//		return models.ErrVersionNotFound
	//	}
	//
	//	return fmt.Errorf("c.Get: %w", err)
	//}

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			return nil
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
		return fmt.Errorf("filepath.WalkDir: %w", err)
	}

	_, err = c.console.RunCmd(ctx, root, q.build())
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
