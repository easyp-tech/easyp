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

func (v Validate) Command() *cli.Command {
	return &cli.Command{
		Name:        "validate-config",
		Aliases:     []string{"validate"},
		Usage:       "validate easyp config file",
		Description: "validate easyp.yaml for syntax and required fields",
		UsageText:   "validate-config [--config path] [--format json|text]",
		Flags: []cli.Flag{
			flags.Config,
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

	format := flags.GetFormat(ctx, flags.JSONFormat)
	switch format {
	case flags.JSONFormat:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return fmt.Errorf("json.Encode: %w", err)
		}
	case flags.TextFormat:
		printValidateText(result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if result.Valid {
		return nil
	}

	return ErrHasValidateIssue
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
