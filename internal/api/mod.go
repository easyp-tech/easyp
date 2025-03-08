package api

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"go.redsock.ru/protopack/internal/core/models"
	"go.redsock.ru/protopack/internal/flags"
	"go.redsock.ru/protopack/internal/fs/fs"

	"go.redsock.ru/protopack/internal/config"
	"go.redsock.ru/protopack/internal/core"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	downloadCmd := &cli.Command{
		Name:        "download",
		Usage:       "download modules to local cache",
		UsageText:   "download modules to local cache",
		Description: "download modules to local cache",
		Action:      m.Download,
	}
	updateCmd := &cli.Command{
		Name:        "update",
		Usage:       "update modules version using version from config",
		UsageText:   "update modules version using version from config",
		Description: "update modules version using version from config",
		Action:      m.Update,
	}
	vendorCmd := &cli.Command{
		Name:        "vendor",
		Usage:       "copy proto files from deps to vendor dir",
		UsageText:   "copy proto files from deps to vendor dir",
		Description: "copy proto files from deps to vendor dir",
		Action:      m.Vendor,
	}

	return &cli.Command{
		Name:                   "mod",
		Aliases:                []string{"m"},
		Usage:                  "package manager",
		UsageText:              "package manager",
		Description:            "package manager",
		ArgsUsage:              "",
		Category:               "",
		BashComplete:           nil,
		Before:                 nil,
		After:                  nil,
		Action:                 nil,
		OnUsageError:           nil,
		Subcommands:            []*cli.Command{downloadCmd, updateCmd, vendorCmd},
		Flags:                  []cli.Flag{},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

func (m Mod) Download(ctx *cli.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}
	dirWalker := fs.NewFSWalker(workingDir, ".")

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	if err := app.Download(ctx.Context, cfg.Deps); err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			os.Exit(1)
		}

		return fmt.Errorf("cmd.Download: %w", err)
	}
	return nil
}

func (m Mod) Update(ctx *cli.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}
	dirWalker := fs.NewFSWalker(workingDir, ".")

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	if err := app.Update(ctx.Context, cfg.Deps); err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			os.Exit(1)
		}

		return fmt.Errorf("cmd.Download: %w", err)
	}
	return nil
}

func (m Mod) Vendor(ctx *cli.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}
	dirWalker := fs.NewFSWalker(workingDir, ".")

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	if err := app.Vendor(ctx.Context); err != nil {
		if errors.Is(err, models.ErrVersionNotFound) {
			os.Exit(1)
		}

		return fmt.Errorf("cmd.Download: %w", err)
	}
	return nil
}
