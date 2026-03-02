package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/schemagen"
)

var _ Handler = (*SchemaGen)(nil)

type SchemaGen struct{}

var (
	flagSchemaGenOutVersioned = &cli.StringFlag{
		Name:       "out-versioned",
		Usage:      "path to versioned schema file",
		Required:   false,
		HasBeenSet: true,
		Value:      schemagen.DefaultVersionedOut,
	}

	flagSchemaGenOutLatest = &cli.StringFlag{
		Name:       "out-latest",
		Usage:      "path to latest schema alias file",
		Required:   false,
		HasBeenSet: true,
		Value:      schemagen.DefaultLatestOut,
	}
)

func (s SchemaGen) Command() *cli.Command {
	return &cli.Command{
		Name:        "schema-gen",
		Usage:       "generate easyp config JSON Schema artifacts",
		UsageText:   "schema-gen [--out-versioned path] [--out-latest path]",
		Description: "generate versioned and latest easyp config JSON Schema artifacts",
		Action:      s.Action,
		Flags: []cli.Flag{
			flagSchemaGenOutVersioned,
			flagSchemaGenOutLatest,
		},
	}
}

func (s SchemaGen) Action(ctx *cli.Context) error {
	if err := schemagen.Run(schemagen.Options{
		VersionedOut: ctx.String(flagSchemaGenOutVersioned.Name),
		LatestOut:    ctx.String(flagSchemaGenOutLatest.Name),
	}); err != nil {
		return fmt.Errorf("schemagen.Run: %w", err)
	}

	return nil
}
