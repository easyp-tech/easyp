package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
)

var _ Handler = (*Lint)(nil)

// Lint is a handler for lint command.
type Lint struct{}

// Format is the format of output.
const (
	TextFormat = "text"
	JSONFormat = "json"
)

var (
	flagLintDirectoryPath = &cli.StringFlag{
		Name:       "path",
		Usage:      "set relative path to directory with proto files",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"p"},
	}

	flagFormat = &cli.GenericFlag{
		Name:       "format",
		Usage:      "set format of output",
		Required:   false,
		HasBeenSet: false,
		Value: &EnumValue{
			Enum:    []string{TextFormat, JSONFormat},
			Default: TextFormat,
		},
		Aliases: []string{"f"},
		EnvVars: []string{"EASYP_FORMAT"},
	}

	ErrHasLintIssue = errors.New("has lint issue")
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
			flagFormat,
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
	err := l.action(ctx)
	if err != nil {
		var e *core.OpenImportFileError

		switch {
		case errors.Is(err, ErrHasLintIssue):
			os.Exit(1)
		case errors.As(err, &e):
			errExit(2, "Cannot import file", "file name", e.FileName)
		default:
			return err
		}
	}

	return nil
}

func (l Lint) action(ctx *cli.Context) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	path := ctx.String(flagLintDirectoryPath.Name)

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	core.SetAllowCommentIgnores(cfg.Lint.AllowCommentIgnores)

	fsWalker := fs.NewFSWalker(workingDir, path)

	app, err := buildCore(ctx.Context, *cfg, fsWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}
	issues, err := app.Lint(ctx.Context, fsWalker)
	if err != nil {
		return fmt.Errorf("c.Lint: %w", err)
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

	return ErrHasLintIssue
}

func printIssues(format string, w io.Writer, issues []core.IssueInfo) error {
	switch format {
	case TextFormat:
		return textPrinter(w, issues)
	case JSONFormat:
		return jsonPrinter(w, issues)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// textPrinter prints the error in text format.
func textPrinter(w io.Writer, issues []core.IssueInfo) error {
	buffer := bytes.NewBuffer(nil)
	for _, issue := range issues {
		buffer.Reset()

		str := fmt.Sprintf("%s:%d:%d:%s %s (%s)",
			issue.Path,
			issue.Position.Line,
			issue.Position.Column,
			issue.SourceName,
			issue.Message,
			issue.RuleName,
		)

		_, _ = buffer.WriteString(str)
		_, _ = buffer.WriteString("\n")
		if _, err := w.Write(buffer.Bytes()); err != nil {
			return fmt.Errorf("w.Write: %w", err)
		}
	}

	return nil
}

// jsonPrinter prints the error in json format.
func jsonPrinter(w io.Writer, issues []core.IssueInfo) error {
	for _, issue := range issues {
		marshalErr := json.NewEncoder(w).Encode(issue)
		if marshalErr != nil {
			return fmt.Errorf("json.NewEncoder.Encode: %w", marshalErr)
		}
	}

	return nil
}
