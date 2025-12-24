package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/flags"
)

type Validate struct{}

var (
	flagValidateFormat = &cli.GenericFlag{
		Name:       "format",
		Usage:      "output format: json or text",
		Required:   false,
		HasBeenSet: false,
		Value: &EnumValue{
			Enum:    []string{JSONFormat, TextFormat},
			Default: JSONFormat,
		},
		Aliases: []string{"f"},
	}
)

func (v Validate) Command() *cli.Command {
	return &cli.Command{
		Name:        "validate-config",
		Aliases:     []string{"validate"},
		Usage:       "validate easyp config file",
		Description: "validate easyp.yaml for syntax and required fields",
		UsageText:   "validate-config [--config path] [--format json|text]",
		Flags: []cli.Flag{
			flags.Config,
			flagValidateFormat,
		},
		Action: v.Action,
	}
}

type validateResult struct {
	Valid  bool                     `json:"valid"`
	Errors []config.ValidationIssue `json:"errors,omitempty"`
}

func (v Validate) Action(ctx *cli.Context) error {
	configPath := ctx.String(flags.Config.Name)
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("os.Getwd: %w", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("config not found: %w", err)
	}

	issues, err := config.ValidateFile(configPath)
	if err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	result := validateResult{
		Valid:  len(issues) == 0,
		Errors: issues,
	}

	format := ctx.String(flagValidateFormat.Name)
	switch format {
	case JSONFormat:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return fmt.Errorf("json.Encode: %w", err)
		}
	case TextFormat:
		printValidateText(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if !result.Valid {
		// Non-zero exit code on validation failure
		os.Exit(1)
	}

	return nil
}

func printValidateText(res validateResult) {
	if res.Valid {
		fmt.Println("VALID: true")
		return
	}

	fmt.Println("VALID: false")
	if len(res.Errors) == 0 {
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ERRORS:")
	fmt.Fprintln(w, "  #\tCODE\tMESSAGE")
	for i, e := range res.Errors {
		fmt.Fprintf(w, "  %d\t%s\t%s\n", i+1, e.Code, e.Message)
	}
	_ = w.Flush()
}
