package api

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/fs/fs"
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

	flagGenerateRoot = &cli.StringFlag{
		Name:       "root",
		Usage:      "set root directory for file search (default: current working directory)",
		Required:   false,
		HasBeenSet: false,
		Value:      "",
		Aliases:    []string{"r"},
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
			flagGenerateRoot,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	logger := slog.Default()

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd: %w", err)
	}

	root := ctx.String(flagGenerateRoot.Name)

	// Determine root directory for file search
	// If --root is not specified, use workingDir (default behavior)
	rootDir := workingDir
	if root != "" {
		if filepath.IsAbs(root) {
			// If root is absolute, use it as is
			rootDir = root
		} else {
			// If root is relative, concatenate with workingDir
			rootDir = filepath.Join(workingDir, root)
		}
	}

	// Normalize the root directory path
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("filepath.Abs(rootDir): %w", err)
	}

	cfg, err := config.New(ctx.Context, ctx.String(flags.Config.Name))
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}
	dirWalker := fs.NewFSWalker(rootDir, ".")
	app, err := buildCore(ctx.Context, *cfg, dirWalker)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	dir := ctx.String(flagGenerateDirectoryPath.Name)
	err = app.Generate(ctx.Context, rootDir, dir)
	if err != nil {
		if errors.Is(err, core.ErrEmptyInputFiles) {
			logger.Warn("empty input files!")
			return nil
		}

		return fmt.Errorf("generator.Generate: %w", err)
	}

	return nil
}
