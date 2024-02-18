package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/package_manager/mod"
	"github.com/easyp-tech/easyp/internal/package_manager/services/storage"
)

var _ Handler = (*Mod)(nil)

// Mod is a handler for package manager
type Mod struct{}

func (m Mod) Command() *cli.Command {
	return &cli.Command{
		Name:         "mod",
		Aliases:      []string{"m"},
		Usage:        "package manager",
		UsageText:    "package manager",
		Description:  "package manager",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       m.Action,
		OnUsageError: nil,
		Subcommands:  nil,
		Flags: []cli.Flag{
			flagCfg,
			// FIXME: Use flags for package_manager
			// flagLintDirectoryPath,
		},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

const (
	envEasypPath     = "EASYPPATH"
	defaultEasypPath = ".easyp"
)

func (m Mod) Action(ctx *cli.Context) error {
	cfg, err := readConfig(ctx)
	if err != nil {
		return fmt.Errorf("readConfig: %w", err)
	}

	easypPath := os.Getenv(envEasypPath)
	if easypPath == "" {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("os.UserHomeDir: %w", err)
		}
		easypPath = filepath.Join(userHomeDir, defaultEasypPath)
	}

	fmt.Printf("Use storage: %s\n", easypPath)

	store := storage.New(easypPath)
	cmd := mod.New(store)

	for _, dependency := range cfg.Deps {
		if err := cmd.Get(ctx.Context, dependency); err != nil {
			return fmt.Errorf("cmd.Get: %w", err)
		}
	}

	return nil
}
