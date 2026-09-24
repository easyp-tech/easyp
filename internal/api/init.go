package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/prompter"
)

var _ Handler = (*Init)(nil)

// Init is a handler for initialization EasyP configuration.
type Init struct{}

var (
	flagInitDirectoryPath = &cli.StringFlag{
		Name:    "dir",
		Usage:   "directory path to initialize",
		Value:   ".",
		Aliases: []string{"d"},
		EnvVars: []string{"EASYP_INIT_DIR"},
	}
	flagInitModule = &cli.StringFlag{
		Name:  "module",
		Usage: "canonical protobuf module identity for a new v1 project",
	}
)

// Command implements Handler.
func (i Init) Command() *cli.Command {
	return &cli.Command{
		Name:        "init",
		Aliases:     []string{"i"},
		Usage:       "initialize configuration",
		UsageText:   "initialize configuration",
		Description: "initialize configuration",
		Action:      i.Action,
		Flags: []cli.Flag{
			flagInitDirectoryPath,
			flagInitModule,
		},
	}
}

// Action implements Handler.
func (i Init) Action(ctx *cli.Context) error {
	rootAbs, err := filepath.Abs(ctx.String(flagInitDirectoryPath.Name))
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	if err := os.MkdirAll(rootAbs, 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	if err := initializeV1(ctx.Context, rootAbs, ctx.String(flagInitModule.Name), prompter.InteractivePrompter{}); err != nil {
		return fmt.Errorf("initializeV1: %w", err)
	}
	return nil
}
