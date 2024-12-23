package api

import (
	"fmt"
	"log/slog"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	lockfile "github.com/easyp-tech/easyp/internal/core/adapters/lock_file"
	moduleconfig "github.com/easyp-tech/easyp/internal/core/adapters/module_config"
	"github.com/easyp-tech/easyp/internal/core/adapters/storage"
	"github.com/easyp-tech/easyp/internal/factories"
	wfs "github.com/easyp-tech/easyp/internal/fs"
)

var _ Handler = (*Init)(nil)

// Init is a handler for initialization EasyP configuration.
type Init struct{}

var (
	flagInitDirectoryPath = &cli.StringFlag{
		Name:       "dir",
		Usage:      "directory path to initialize",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"d"},
		EnvVars:    []string{"EASYP_INIT_DIR"},
	}
)

// Command implements Handler.
func (i Init) Command() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Aliases:     []string{"i"},
		Usage:       "initialize configuration",
		UsageText:   "initialize configuration",
		Description: "initialize configuration",
		Action:      i.Action,
		Flags: []cli.Flag{
			flagInitDirectoryPath,
		},
	}
}

// Action implements Handler.
func (i Init) Action(ctx *cli.Context) error {
	rootPath := ctx.String(flagInitDirectoryPath.Name)
	dirFS := wfs.Disk(rootPath)

	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("config.ReadConfig: %w", err)
	}
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	lintRules, err := cfg.BuildLinterRules()
	if err != nil {
		return fmt.Errorf("cfg.BuildLinterRules: %w", err)
	}

	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		return fmt.Errorf("factories.NewModuleReflect: %w", err)
	}

	lockFile := lockfile.New()
	easypPath, err := getEasypPath()
	if err != nil {
		return fmt.Errorf("getEasypPath: %w", err)
	}

	store := storage.New(easypPath, lockFile)

	moduleCfg := moduleconfig.New()

	app := core.New(
		lintRules,
		cfg.Lint.Ignore,
		cfg.Deps,
		moduleReflect,
		cfg.Lint.IgnoreOnly,
		slog.Default(), // TODO: remove global state
		lo.Map(cfg.Generate.Plugins, func(p config.Plugin, _ int) core.Plugin {
			return core.Plugin{
				Name:    p.Name,
				Out:     p.Out,
				Options: p.Opts,
			}
		}),
		core.Inputs{
			Dirs: lo.Filter(lo.Map(cfg.Generate.Inputs, func(i config.Input, _ int) string {
				return i.Directory
			}), func(s string, _ int) bool {
				return s != ""
			}),
		},
		console.New(),
		store,
		moduleCfg,
		lockFile,
	)

	err = app.Initialize(ctx.Context, dirFS, []string{"DEFAULT"})
	if err != nil {
		return fmt.Errorf("initer.Initialize: %w", err)
	}

	return nil
}
