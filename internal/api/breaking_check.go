package api

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
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
			flagFormat,
			flagAgainstBranchName,
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
	err := b.action(ctx)
	if err != nil {
		var e *core.OpenImportFileError
		var g *core.GitRefNotFoundError

		switch {
		case errors.Is(err, ErrBreakingCheckIssue):
			os.Exit(1)
		case errors.As(err, &e):
			errExit(2, "Cannot import file", "file name", e.FileName)
		case errors.As(err, &g):
			errExit(2, "Cannot find git ref", "ref", g.GitRef)
		case errors.Is(err, core.ErrRepositoryDoesNotExist):
			errExit(2, "Repository does not exist in current directory")
		default:
			return err
		}
	}

	return nil
}

func (b BreakingCheck) action(ctx *cli.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}

	path := ctx.String(flagLintDirectoryPath.Name)
	against := ctx.String(flagAgainstBranchName.Name)
	if against != "" {
		cfg.BreakingCheck.AgainstGitRef = against
	}

	dirWalker := fs.NewFSWalker(workingDir, ".")
	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	issues, err := app.BreakingCheck(ctx.Context, workingDir, path)
	if err != nil {
		return fmt.Errorf("app.BreakingCheck: %w", err)
	}

	if len(issues) == 0 {
		return nil
	}

	format := ctx.String(flagFormat.Name)
	if err := printIssues(
		format,
		os.Stdout,
		issues,
	); err != nil {
		return fmt.Errorf("printLintErrors: %w", err)
	}

	return ErrBreakingCheckIssue
}
