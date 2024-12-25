package api

import (
	"fmt"
	"path"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/api/factories"
	"github.com/easyp-tech/easyp/internal/generate"
)

var _ Handler = (*Generate)(nil)

// Generate is a handler for generate command.
type Generate struct{}

var (
	flagGenerateDirectoryPath = &cli.StringFlag{
		Name:       "path",
		Usage:      "set path to directory with proto files",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"p"},
		EnvVars:    []string{"EASYP_ROOT_GENERATE_PATH"},
	}
)

// Command implements Handler.
func (g Generate) Command() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"g"},
		Usage:       "generate code from proto files",
		UsageText:   "generate code from proto files",
		Description: "generate code from proto files",
		Action:      g.Action,
		Flags: []cli.Flag{
			flagGenerateDirectoryPath,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	moduleReflect, err := factories.NewModuleReflect()
	if err != nil {
		return fmt.Errorf("factories.NewModuleReflect: %w", err)
	}

	generator := generate.New(generate.Config{
		Deps: cfg.Deps,
		Plugins: lo.Map(cfg.Generate.Plugins, func(p config.Plugin, _ int) generate.Plugin {
			return generate.Plugin{
				Name:    p.Name,
				Out:     p.Out,
				Options: p.Opts,
			}
		}),
		ModuleReflect: moduleReflect,
		Inputs: generate.Inputs{
			Dirs: lo.Filter(lo.Map(cfg.Generate.Inputs, func(i config.Input, _ int) string {
				return i.Directory
			}), func(s string, _ int) bool {
				return s != ""
			}),
		},
		ProtoRoot:       cfg.Generate.ProtoRoot,
		GenerateOutDirs: cfg.Generate.GenerateOutDirs,
	})

	dir := ctx.String(flagGenerateDirectoryPath.Name)
	if cfg.Generate.DependencyEntryPoint != nil {
		modulePaths, err := moduleReflect.GetModulePath(ctx.Context, cfg.Generate.DependencyEntryPoint.Dep)
		if err != nil {
			return fmt.Errorf("moduleReflect.GetModulePath: %w", err)
		}
		dir = path.Join(modulePaths, cfg.Generate.DependencyEntryPoint.Path)
	}
	err = generator.Generate(ctx.Context, ".", dir)
	if err != nil {
		return fmt.Errorf("generator.Generate: %w", err)
	}

	return nil
}
