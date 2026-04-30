package api

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

var _ Handler = (*BreakingCheck)(nil)

// BreakingCheck is a handler for breaking command
type BreakingCheck struct{}

var (
	flagAgainstBranchName = &cli.StringFlag{
		Name:       "against",
		Usage:      "set branch to compare with",
		Required:   true,
		HasBeenSet: true,
		Value:      "master",
	}

	flagBreakingCheckRoot = &cli.StringFlag{
		Name:       "root",
		Usage:      "set root directory for file search (default: current working directory)",
		Required:   false,
		HasBeenSet: false,
		Value:      "",
		Aliases:    []string{"r"},
	}

	ErrBreakingCheckIssue = errors.New("has breaking check issue")
)

func (b BreakingCheck) Command() *cli.Command {
	return &cli.Command{
		Name:         "breaking",
		Usage:        "api breaking check",
		UsageText:    "api breaking check",
		Description:  "api breaking check",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       b.Action,
		OnUsageError: nil,
		Subcommands:  nil,
		Flags: []cli.Flag{
			flagLintDirectoryPath,
			flagAgainstBranchName,
			flagBreakingCheckRoot,
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

func (b BreakingCheck) Action(ctx *cli.Context) error {
	log := getLogger(ctx)

	err := b.action(ctx, log)
	if err != nil {
		var e *core.OpenImportFileError
		var g *core.GitRefNotFoundError

		switch {
		case errors.Is(err, ErrBreakingCheckIssue):
			os.Exit(1)
		case errors.As(err, &e):
			errExit(log, 2, "Cannot import file", slog.String("file name", e.FileName))
		case errors.As(err, &g):
			errExit(log, 2, "Cannot find git ref", slog.String("ref", g.GitRef))
		case errors.Is(err, core.ErrRepositoryDoesNotExist):
			errExit(log, 2, "Repository does not exist in current directory")
		default:
			return err
		}
	}

	return nil
}

func (b BreakingCheck) action(ctx *cli.Context, log logger.Logger) error {
	configPath, projectRoot, breakingCheckRoot, err := resolveRoots(ctx, flagBreakingCheckRoot.Name)
	if err != nil {
		return fmt.Errorf("resolveRoots: %w", err)
	}

	cfg, err := config.New(ctx.Context, configPath)
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}

	// Walker for Core (lockfile etc) - strictly based on project root
	projectWalker := fs.NewFSWalker(projectRoot, ".")
	app, err := buildCore(ctx.Context, log, *cfg, projectWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	path := ctx.String(flagLintDirectoryPath.Name)
	against := ctx.String(flagAgainstBranchName.Name)
	if cfg.BreakingCheck.AgainstGitRef == "" {
		cfg.BreakingCheck.AgainstGitRef = against
	}

	issues, err := app.BreakingCheck(ctx.Context, projectRoot, breakingCheckRoot, path)
	if err != nil {
		return fmt.Errorf("app.BreakingCheck: %w", err)
	}

	if len(issues) == 0 {
		return nil
	}

	format := flags.GetFormat(ctx, flags.TextFormat)
	if err := printIssues(
		format,
		os.Stdout,
		issues,
	); err != nil {
		return fmt.Errorf("printLintErrors: %w", err)
	}

	return ErrBreakingCheckIssue
}
