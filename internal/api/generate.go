package api

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/generation"
)

var _ Handler = (*Generate)(nil)

// Generate is a handler for generate command.
type Generate struct{}

var (
	flagGenerateDescriptorSetOut = &cli.StringFlag{
		Name:     "descriptor_set_out",
		Usage:    "output path for the binary FileDescriptorSet",
		Required: false,
	}

	flagGenerateDescriptorSetOutDir = &cli.StringFlag{
		Name:  "descriptor_set_out_dir",
		Usage: "output directory for one binary FileDescriptorSet per project and module (exclusive with descriptor_set_out)",
	}

	flagGenerateIncludeImports = &cli.BoolFlag{
		Name:     "include_imports",
		Usage:    "include all transitive dependencies in the FileDescriptorSet",
		Required: false,
	}
	flagGenerateProject = &cli.StringFlag{
		Name:  "project",
		Usage: "generate only the consumer project in this directory",
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
			flagGenerateDescriptorSetOut,
			flagGenerateDescriptorSetOutDir,
			flagGenerateIncludeImports,
			flagGenerateProject,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
	if ctx.String(flagGenerateDescriptorSetOut.Name) != "" && ctx.String(flagGenerateDescriptorSetOutDir.Name) != "" {
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
		WorkDir: root, Project: ctx.String(flagGenerateProject.Name),
		DescriptorSetOut:    ctx.String(flagGenerateDescriptorSetOut.Name),
		DescriptorSetOutDir: ctx.String(flagGenerateDescriptorSetOutDir.Name), IncludeImports: ctx.Bool(flagGenerateIncludeImports.Name),
	})
}
