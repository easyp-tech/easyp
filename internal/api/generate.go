package api

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
	"github.com/easyp-tech/easyp/internal/generation"
)

var _ Handler = (*Generate)(nil)

// Generate is a handler for generate command.
type Generate struct{}

const (
	flagGenerateDescriptorSetOut    = "descriptor_set_out"
	flagGenerateDescriptorSetOutDir = "descriptor_set_out_dir"
	flagGenerateAll                 = "all"
	flagGenerateWorkspace           = "workspace"
	flagGenerateIncludeImports      = "include_imports"
	flagGenerateProject             = "project"
	flagGenerateConfig              = "gen-config"
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
			flags.Frozen(),
			&cli.StringFlag{Name: flagGenerateConfig, TakesFile: true, Usage: "select a native generation config file (exclusive with --project and --all)"},
			&cli.StringFlag{Name: flagGenerateDescriptorSetOut, Usage: "output path for the binary FileDescriptorSet"},
			&cli.StringFlag{Name: flagGenerateDescriptorSetOutDir, Usage: "output directory for one binary FileDescriptorSet per project and module (exclusive with descriptor_set_out)"},
			&cli.BoolFlag{Name: flagGenerateIncludeImports, Usage: "include all transitive dependencies in the FileDescriptorSet"},
			&cli.StringSliceFlag{Name: flagGenerateProject, Usage: "select a consumer project directory (repeatable)"},
			&cli.BoolFlag{Name: flagGenerateAll, Usage: "explicitly select all generation projects below the working directory"},
			&cli.StringFlag{Name: flagGenerateWorkspace, Usage: "explicit workspace root for repository-relative module paths"},
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	if ctx.IsSet(flagGenerateConfig) && ctx.String(flagGenerateConfig) == "" {
		return fmt.Errorf("--gen-config must not be empty")
	}
	if ctx.String(flagGenerateDescriptorSetOut) != "" && ctx.String(flagGenerateDescriptorSetOutDir) != "" {
		return fmt.Errorf("--descriptor_set_out and --descriptor_set_out_dir are mutually exclusive")
	}
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getwd: %w", err)
	}
	cache, err := moduleCache(ctx)
	if err != nil {
		return fmt.Errorf("moduleCache: %w", err)
	}
	return generation.Run(ctx.Context, getLogger(ctx), cache, generation.Request{
		Frozen: flags.IsFrozen(ctx), WorkDir: root, GenConfig: ctx.String(flagGenerateConfig),
		Projects: ctx.StringSlice(flagGenerateProject), AllProjects: ctx.Bool(flagGenerateAll), WorkspaceRoot: ctx.String(flagGenerateWorkspace),
		DescriptorSetOut:    ctx.String(flagGenerateDescriptorSetOut),
		DescriptorSetOutDir: ctx.String(flagGenerateDescriptorSetOutDir), IncludeImports: ctx.Bool(flagGenerateIncludeImports),
	})
}
