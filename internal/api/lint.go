package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/urfave/cli/v2"

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
	cfg, err := readConfig(ctx)
	if err != nil {
		return fmt.Errorf("readConfig: %w", err)
	}

	buf, err := io.ReadAll(cfgFile)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	jsBuf, err := yaml.YAMLToJSON(buf)
	if err != nil {
		return fmt.Errorf("yaml.YAMLToJSON: %w", err)
	}

	cfg := Config{}
	err = json.Unmarshal(jsBuf, &cfg)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	lintRules, err := cfg.buildLinterRules()
	if err != nil {
		return fmt.Errorf("cfg.buildLinterRules: %w", err)
	}

	c := lint.New(lintRules)

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

func (cfg Config) buildLinterRules() ([]lint.Rule, error) {
	if len(cfg.Lint.Use) > 0 {
		return cfg.buildFromUse()
	}

	return cfg.buildStdRules()
}

func (cfg Config) buildFromUse() ([]lint.Rule, error) {
	var useRule []lint.Rule

	for _, ruleName := range cfg.Lint.Use {
		rule, ok := rules.Rules(rules.Config{
			PackageDirectoryMatchRoot: ".",           // TODO: Move to config
			EnumZeroValueSuffixPrefix: "UNSPECIFIED", // TODO: Move to config
			ServiceSuffixSuffix:       "Service",     // TODO: Move to config
		})[ruleName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", lint.ErrInvalidRule, ruleName)
		}

		useRule = append(useRule, rule)
	}

	return useRule, nil
}

// todo: reflect
func (cfg Config) buildStdRules() ([]lint.Rule, error) {
	var useRule []lint.Rule

	// Minimal
	if cfg.Lint.DirectorySamePackage.Activated {
		useRule = append(useRule, &cfg.Lint.DirectorySamePackage.Value)
	}

	if cfg.Lint.PackageDefined.Activated {
		useRule = append(useRule, &cfg.Lint.PackageDefined.Value)
	}

	if cfg.Lint.PackageDirectoryMatch.Activated {
		useRule = append(useRule, &cfg.Lint.PackageDirectoryMatch.Value)
	}

	if cfg.Lint.PackageSameDirectory.Activated {
		useRule = append(useRule, &cfg.Lint.PackageSameDirectory.Value)
	}

	// Basic

	if cfg.Lint.EnumFirstValueZero.Activated {
		useRule = append(useRule, &cfg.Lint.EnumFirstValueZero.Value)
	}

	if cfg.Lint.EnumNoAllowAlias.Activated {
		useRule = append(useRule, &cfg.Lint.EnumNoAllowAlias.Value)
	}

	if cfg.Lint.EnumPascalCase.Activated {
		useRule = append(useRule, &cfg.Lint.EnumPascalCase.Value)
	}

	if cfg.Lint.EnumValueUpperSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.EnumValueUpperSnakeCase.Value)
	}

	if cfg.Lint.FieldLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.FieldLowerSnakeCase.Value)
	}

	if cfg.Lint.ImportNoPublic.Activated {
		useRule = append(useRule, &cfg.Lint.ImportNoPublic.Value)
	}

	if cfg.Lint.ImportNoWeak.Activated {
		useRule = append(useRule, &cfg.Lint.ImportNoWeak.Value)
	}

	if cfg.Lint.ImportUsed.Activated {
		useRule = append(useRule, &cfg.Lint.ImportUsed.Value)
	}

	if cfg.Lint.MessagePascalCase.Activated {
		useRule = append(useRule, &cfg.Lint.MessagePascalCase.Value)
	}

	if cfg.Lint.OneofLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.OneofLowerSnakeCase.Value)
	}

	if cfg.Lint.PackageLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.PackageLowerSnakeCase.Value)
	}

	// Default
	if cfg.Lint.EnumValuePrefix.Activated {
		useRule = append(useRule, &cfg.Lint.EnumValuePrefix.Value)
	}

	if cfg.Lint.EnumZeroValueSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.EnumZeroValueSuffix.Value)
	}

	if cfg.Lint.FileLowerSnakeCase.Activated {
		useRule = append(useRule, &cfg.Lint.FileLowerSnakeCase.Value)
	}

	if cfg.Lint.RPCRequestResponseUnique.Activated {
		useRule = append(useRule, &cfg.Lint.RPCRequestResponseUnique.Value)
	}

	if cfg.Lint.RPCRequestStandardName.Activated {
		useRule = append(useRule, &cfg.Lint.RPCRequestStandardName.Value)
	}

	if cfg.Lint.RPCResponseStandardName.Activated {
		useRule = append(useRule, &cfg.Lint.RPCResponseStandardName.Value)
	}

	if cfg.Lint.PackageVersionSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.PackageVersionSuffix.Value)
	}

	if cfg.Lint.ServiceSuffix.Activated {
		useRule = append(useRule, &cfg.Lint.ServiceSuffix.Value)
	}

	// Comments
	if cfg.Lint.RPCNoClientStreaming.Activated {
		useRule = append(useRule, &cfg.Lint.RPCNoClientStreaming.Value)
	}

	if cfg.Lint.RPCNoServerStreaming.Activated {
		useRule = append(useRule, &cfg.Lint.RPCNoServerStreaming.Value)
	}

	return useRule, nil
}
