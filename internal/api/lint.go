package api

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

var _ Handler = (*Lint)(nil)

// Lint is a handler for lint command.
type Lint struct{}

var (
	flagLintDirectoryPath = &cli.StringFlag{
		Name:       "path",
		Usage:      "set path to directory with proto files",
		Required:   true,
		HasBeenSet: true,
		Value:      ".",
		Aliases:    []string{"p"},
		EnvVars:    []string{"EASYP_PATH"},
	}
)

// Command implements Handler.
func (l Lint) Command() *cli.Command {
	return &cli.Command{
		Name:         "lint",
		Aliases:      []string{"l"},
		Usage:        "linting proto files",
		UsageText:    "linting proto files",
		Description:  "linting proto files",
		ArgsUsage:    "",
		Category:     "",
		BashComplete: nil,
		Before:       nil,
		After:        nil,
		Action:       l.Action,
		OnUsageError: nil,
		Subcommands:  nil,
		Flags: []cli.Flag{
			flagCfg,
			flagLintDirectoryPath,
		},
		SkipFlagParsing:        false,
		HideHelp:               false,
		HideHelpCommand:        false,
		Hidden:                 false,
		UseShortOptionHandling: false,
		HelpName:               "help",
		CustomHelpTemplate:     "",
	}
}

// Action implements Handler.
func (l Lint) Action(ctx *cli.Context) error {
	cfgFile, err := os.Open(ctx.String(flagCfg.Name))
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}

	cfg := Config{}
	err = yaml.NewDecoder(cfgFile).Decode(&cfg)
	if err != nil {
		return fmt.Errorf("yaml.NewDecoder.Decode: %w", err)
	}

	var useRule []lint.Rule
	for _, ruleName := range cfg.Lint.Use {
		rule, ok := rules.Rules(rules.Config{
			PackageDirectoryMatchRoot: ".",           // TODO: Move to config
			EnumZeroValueSuffixPrefix: "UNSPECIFIED", // TODO: Move to config
			ServiceSuffixSuffix:       "Service",     // TODO: Move to config
		})[ruleName]
		if !ok {
			return fmt.Errorf("%w: %s", lint.ErrInvalidRule, ruleName)
		}

		useRule = append(useRule, rule)
	}

	c := lint.New(useRule)

	dirFS := os.DirFS(ctx.String(flagLintDirectoryPath.Name))

	res := c.Lint(ctx.Context, dirFS)
	if splitErr, ok := res.(interface{ Unwrap() []error }); ok {

		for _, err := range splitErr.Unwrap() {
			slog.Info(err.Error())
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("c.Lint: %w", err)
	}

	return nil
}
