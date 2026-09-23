package api

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
)

// Validate handles recursive configuration validation.
type Validate struct{}

// Command implements Handler.
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

// Action implements Handler.
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

	var report validateResult
	for _, issue := range issues {
		if issue.Severity == config.SeverityError {
			report.Errors = append(report.Errors, issue)
			continue
		}
		report.Warnings = append(report.Warnings, issue)
	}
	report.Valid = len(report.Errors) == 0

	output := ctx.App.Writer
	if output == nil {
		output = os.Stdout
	}
	format := flags.GetFormat(ctx, flags.JSONFormat)
	switch format {
	case flags.JSONFormat:
		enc := json.NewEncoder(output)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("Encode: %w", err)
		}
	case flags.TextFormat:
		if err := writeValidationText(output, report); err != nil {
			return fmt.Errorf("writeValidationText: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if !report.Valid {
		return ErrHasValidateIssue
	}
	return nil
}

func validationTarget(ctx *cli.Context) string {
	for _, scope := range ctx.Lineage() {
		if scope.IsSet(flags.Config.Name) {
			return scope.String(flags.Config.Name)
		}
	}
	return "."
}

func writeValidationText(output io.Writer, report validateResult) error {
	if _, err := fmt.Fprintf(output, "VALID: %t\n", report.Valid); err != nil {
		return fmt.Errorf("Fprintf: %w", err)
	}

	w := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)

	if len(report.Errors) > 0 {
		if _, err := fmt.Fprintln(w, "ERRORS:\n  #\tFILE\tLOCATION\tCODE\tMESSAGE"); err != nil {
			return fmt.Errorf("Fprintln: %w", err)
		}
		for i, issue := range report.Errors {
			if _, err := fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, issue.File, validationLocation(issue), issue.Code, issue.Message); err != nil {
				return fmt.Errorf("Fprintf: %w", err)
			}
		}
	}

	if len(report.Warnings) > 0 {
		if _, err := fmt.Fprintln(w, "WARNINGS:\n  #\tFILE\tLOCATION\tCODE\tMESSAGE"); err != nil {
			return fmt.Errorf("Fprintln: %w", err)
		}
		for i, issue := range report.Warnings {
			if _, err := fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\n", i+1, issue.File, validationLocation(issue), issue.Code, issue.Message); err != nil {
				return fmt.Errorf("Fprintf: %w", err)
			}
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("Flush: %w", err)
	}
	return nil
}

func validationLocation(issue config.ValidationIssue) string {
	if issue.Line == 0 {
		return "-"
	}
	return fmt.Sprintf("%d:%d", issue.Line, issue.Column)
}
