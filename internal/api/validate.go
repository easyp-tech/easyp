package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
)

type Validate struct{}

func (v Validate) Command() *cli.Command {
	return &cli.Command{
		Name:        "validate-config",
		Aliases:     []string{"validate"},
		Usage:       "validate EasyP configuration files",
		Description: "recursively validate v1 EasyP files in the current directory or a selected path",
		UsageText:   "validate-config [--config file-or-directory] [--format json|text]",
		Flags: []cli.Flag{
			flags.Config,
			flags.Format,
		},
		Action: v.Action,
	}
}

type validateResult struct {
	Valid    bool                     `json:"valid"`
	Errors   []config.ValidationIssue `json:"errors,omitempty"`
	Warnings []config.ValidationIssue `json:"warnings,omitempty"`
}

func (v Validate) Action(ctx *cli.Context) error {
	configPath := validationTarget(ctx)
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Getwd: %w", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	issues, err := v1.ValidatePath(configPath)
	if err != nil {
		return fmt.Errorf("ValidatePath: %w", err)
	}

	// Separate errors from warnings - only errors cause validation failure
	var errors, warnings []config.ValidationIssue
	for _, issue := range issues {
		if issue.Severity == config.SeverityError {
			errors = append(errors, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	result := validateResult{
		Valid:    !config.HasErrors(issues),
		Errors:   errors,
		Warnings: warnings,
	}

	format := flags.GetFormat(ctx, flags.JSONFormat)
	switch format {
	case flags.JSONFormat:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return fmt.Errorf("Encode: %w", err)
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

func validationTarget(ctx *cli.Context) string {
	for _, scope := range ctx.Lineage() {
		if scope.IsSet(flags.Config.Name) {
			return scope.String(flags.Config.Name)
		}
	}
	return "."
}

func printValidateText(res validateResult) {
	if res.Valid {
		fmt.Println("VALID: true")
	} else {
		fmt.Println("VALID: false")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	if len(res.Errors) > 0 {
		fmt.Fprintln(w, "ERRORS:")
		fmt.Fprintln(w, "  #\tFILE\tLOCATION\tCODE\tMESSAGE")
		for i, e := range res.Errors {
			fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, e.File, validationLocation(e), e.Code, e.Message)
		}
	}

	if len(res.Warnings) > 0 {
		fmt.Fprintln(w, "WARNINGS:")
		fmt.Fprintln(w, "  #\tFILE\tLOCATION\tCODE\tMESSAGE")
		for i, e := range res.Warnings {
			fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, e.File, validationLocation(e), e.Code, e.Message)
		}
	}

	_ = w.Flush()
}

func validationLocation(issue config.ValidationIssue) string {
	if issue.Line == 0 {
		return "-"
	}
	return fmt.Sprintf("%d:%d", issue.Line, issue.Column)
}
