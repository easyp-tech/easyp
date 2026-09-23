package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

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
	log := getLogger(ctx)

	rootPath := ctx.String(flagInitDirectoryPath.Name)
	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(rootAbs, 0o755); err != nil {
		return err
	}
	v1Project, err := shouldInitializeV1(rootAbs)
	if err != nil {
		return err
	}
	if ctx.String(flagInitModule.Name) != "" || v1Project {
		return initializeV1(ctx.Context, rootAbs, ctx.String(flagInitModule.Name), prompter.InteractivePrompter{})
	}
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

func shouldInitializeV1(root string) (bool, error) {
	raw, err := os.ReadFile(filepath.Join(root, "easyp.yaml"))
	if err == nil {
		var header struct {
			Version string `yaml:"version"`
		}
		if err := yaml.Unmarshal(raw, &header); err != nil {
			return false, err
		}
		return header.Version == "v1", nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
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
