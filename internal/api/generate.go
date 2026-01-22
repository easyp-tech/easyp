package api

import (
	"errors"
	"fmt"
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

	flagGenerateDescriptorSetOut = &cli.StringFlag{
		Name:     "descriptor_set_out",
		Usage:    "output path for the binary FileDescriptorSet",
		Required: false,
	}

	flagGenerateIncludeImports = &cli.BoolFlag{
		Name:     "include_imports",
		Usage:    "include all transitive dependencies in the FileDescriptorSet",
		Required: false,
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
			flagGenerateDescriptorSetOut,
			flagGenerateIncludeImports,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	log := getLogger(ctx)

	configPath, projectRoot, generateRoot, err := resolveRoots(ctx, flagGenerateRoot.Name)
	if err != nil {
		return err
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

	dir := ctx.String(flagGenerateDirectoryPath.Name)
	descriptorSetOut := ctx.String(flagGenerateDescriptorSetOut.Name)
	includeImports := ctx.Bool(flagGenerateIncludeImports.Name)

	if err := app.Generate(ctx.Context, generateRoot, dir, descriptorSetOut, includeImports); err != nil {
		if errors.Is(err, core.ErrEmptyInputFiles) {
			log.Warn(ctx.Context, "empty input files!")
			return nil
		}
		return fmt.Errorf("generator.Generate: %w", err)
	}

	return nil
}

// resolveRoots computes configPath (absolute), projectRoot (dir of config), and operation root based on provided root flag.
func resolveRoots(ctx *cli.Context, rootFlagName string) (string, string, string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", "", "", fmt.Errorf("os.Getwd: %w", err)
	}

	root := ctx.String(rootFlagName)
	configPath := ctx.String(flags.Config.Name)

	// 1. Determine Project Root (for config and lockfile)
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(workingDir, configPath)
	}
	projectRoot := filepath.Dir(configPath)

	// 2. Determine operation root (where to search for files)
	var opRoot string
	if root != "" {
		if filepath.IsAbs(root) {
			opRoot = root
		} else {
			opRoot = filepath.Join(workingDir, root)
		}
	} else {
		opRoot = projectRoot
	}

	// Normalize to absolute path
	opRoot, err = filepath.Abs(opRoot)
	if err != nil {
		return "", "", "", fmt.Errorf("filepath.Abs(opRoot): %w", err)
	}

	return configPath, projectRoot, opRoot, nil
}
