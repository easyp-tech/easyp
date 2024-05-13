package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	wfs "github.com/easyp-tech/easyp/internal/fs"
	"github.com/easyp-tech/easyp/internal/initialization"
)

var _ Handler = (*Init)(nil)

// Init is a handler for initialization EasyP configuration.
type Init struct{}

var (
	flagInitDirectoryPath = &cli.StringFlag{
		Name:       "dir",
		Usage:      "directory path to initialize",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"d"},
		EnvVars:    []string{"EASYP_INIT_DIR"},
	}
)

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
		},
	}
}

// Action implements Handler.
func (i Init) Action(ctx *cli.Context) error {
	rootPath := ctx.String(flagInitDirectoryPath.Name)
	dirFS := wfs.Disk(rootPath)

	initer := initialization.New()
	err := initer.Initialize(ctx.Context, dirFS, []string{"DEFAULT"})
	if err != nil {
		return fmt.Errorf("initer.Initialize: %w", err)
	}

	return nil
}
