package api

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/prompter"
	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/rules"
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
		},
	}
}

// Action implements Handler.
func (i Init) Action(ctx *cli.Context) error {
	log := getLogger(ctx)

	rootPath := ctx.String(flagInitDirectoryPath.Name)
	dirFS := fs.NewFSWalker(rootPath, ".")

	cfg := &config.Config{}

	app, err := buildCore(ctx.Context, log, *cfg, dirFS)
	if err != nil {
		return fmt.Errorf("buildCore: %w", err)
	}

	opts := core.InitOptions{
		TemplateData: defaultTemplateData(),
		Prompter:     prompter.InteractivePrompter{},
	}

	err = app.Initialize(ctx.Context, dirFS, opts)
	if err != nil {
		return fmt.Errorf("app.Initialize: %w", err)
	}

	return nil
}

// defaultTemplateData builds InitTemplateData from all available rule groups.
func defaultTemplateData() core.InitTemplateData {
	groups := rules.AllGroups()
	lintGroups := make([]core.LintGroup, len(groups))
	for i, g := range groups {
		lintGroups[i] = core.LintGroup{
			Name:  g.Name,
			Rules: g.Rules,
		}
	}

	return core.InitTemplateData{
		LintGroups:          lintGroups,
		EnumZeroValueSuffix: "_NONE",
		ServiceSuffix:       "API",
		AgainstGitRef:       "master",
	}
}
