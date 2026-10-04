package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/prompter"
	"github.com/easyp-tech/easyp/internal/flags"
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
		Usage:       "initialize native v1 project configuration",
		UsageText:   "easyp init [--dir DIR] [--module MODULE]",
		Description: "Create protobuf.mod, easyp.yaml, and easyp.gen.yaml without overwriting files unless confirmed.",
		Action:      i.Action,
		Flags: []cli.Flag{
			flags.Frozen(),
			flagInitDirectoryPath,
			flagInitModule,
		},
	}
}

// Action implements Handler.
func (i Init) Action(ctx *cli.Context) error {
	if flags.IsFrozen(ctx) {
		return fmt.Errorf("init is not allowed in frozen mode")
	}
	rootAbs, err := filepath.Abs(ctx.String(flagInitDirectoryPath.Name))
	if err != nil {
		return fmt.Errorf("Abs: %w", err)
	}
	if err := os.MkdirAll(rootAbs, 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	reader, writer := migrationIO(ctx)
	prompt := terminalInitializationPrompter{terminal: migrationHasTerminal(reader, writer), prompt: prompter.InteractivePrompter{}}
	if err := initializeV1(ctx.Context, rootAbs, ctx.String(flagInitModule.Name), prompt); err != nil {
		return fmt.Errorf("initializeV1: %w", err)
	}
	return nil
}
