package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/schemagen"
)

// SchemaGen writes JSON Schemas for the v1 YAML documents.
type SchemaGen struct{}

var _ Handler = (*SchemaGen)(nil)

var flagSchemaGenOutDir = &cli.StringFlag{
	Name:  "out-dir",
	Usage: "directory for generated v1 JSON Schemas",
	Value: schemagen.DefaultOutDir,
}

// Command implements Handler.
func (s SchemaGen) Command() *cli.Command {
	return &cli.Command{
		Name:   "schema-gen",
		Usage:  "generate JSON Schemas for v1 YAML files",
		Action: s.Action,
		Flags:  []cli.Flag{flagSchemaGenOutDir},
	}
}

// Action implements Handler.
func (s SchemaGen) Action(ctx *cli.Context) error {
	if err := schemagen.Run(schemagen.Options{OutDir: ctx.String(flagSchemaGenOutDir.Name)}); err != nil {
		return fmt.Errorf("Run: %w", err)
	}
	return nil
}
