package api

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/api/config"
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ Handler = (*Lint)(nil)

// Lint is a handler for lint command.
type Lint struct{}

var (
	flagLintDirectoryPath = &cli.StringFlag{
		Name:       "path",
		Usage:      "set path to directory with proto files",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"p"},
		EnvVars:    []string{"EASYP_PATH"},
	}
)

// Command implements Handler.
func (l Lint) Command() *cli.Command {
	return &cli.Command{
		Name:         "lint",
		Aliases:      []string{"l"},
		Usage:        "linting proto files",
		UsageText:    "linting proto files",
		Description:  "linting proto files",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       l.Action,
		OnUsageError: nil,
		Subcommands:  nil,
		Flags: []cli.Flag{
			flagLintDirectoryPath,
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

// Action implements Handler.
func (l Lint) Action(ctx *cli.Context) error {
	cfg, err := config.ReadConfig(ctx)
	if err != nil {
		return fmt.Errorf("ReadConfig: %w", err)
	}

	lintRules, err := cfg.BuildLinterRules()
	if err != nil {
		return fmt.Errorf("cfg.buildLinterRules: %w", err)
	}

	rootPath := ctx.String(flagLintDirectoryPath.Name)
	dirFS := os.DirFS(rootPath)

	c := lint.New(lintRules, cfg.Lint.Ignore, cfg.Deps)

	res := c.Lint(ctx.Context, dirFS)
	if splitErr, ok := res.(interface{ Unwrap() []error }); ok {

		if err := printLintErrors(os.Stderr, splitErr.Unwrap()); err != nil {
			return fmt.Errorf("printLintErrors: %w", err)
		}

		os.Exit(1)

		return nil
	}

	if err != nil {
		return fmt.Errorf("c.Lint: %w", err)
	}

	return nil
}

func printLintErrors(w io.Writer, errs []error) error {
	buffer := bytes.NewBuffer(nil)
	for _, err := range errs {
		buffer.Reset()

		_, _ = buffer.WriteString(err.Error())
		_, _ = buffer.WriteString("\n")
		if _, err := w.Write(buffer.Bytes()); err != nil {
			return err
		}
	}

	return nil
}
