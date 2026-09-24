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
			flagGenerateIncludeImports,
			flagGenerateProject,
		},
		HelpName: "help",
	}
}

// Action implements Handler.
func (g Generate) Action(ctx *cli.Context) error {
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
		DescriptorSetOut: ctx.String(flagGenerateDescriptorSetOut.Name), IncludeImports: ctx.Bool(flagGenerateIncludeImports.Name),
	})
}
