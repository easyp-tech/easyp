package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/flags"
)

func TestValidateActionPreservesOutputErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		format string
	}{
		{name: "text output", format: flags.TextFormat},
		{name: "json output", format: flags.JSONFormat},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			path := filepath.Join(root, "easyp.yaml")
			require.NoError(t, os.WriteFile(path, []byte("version: v1\n"), 0o600))
			output, err := os.CreateTemp(root, "output-*")
			require.NoError(t, err)
			require.NoError(t, output.Close())
			ctx := validationContext(t, path, tt.format, output)

			err = (Validate{}).Action(ctx)

			require.ErrorIs(t, err, os.ErrClosed)
			var pathErr *os.PathError
			require.ErrorAs(t, err, &pathErr)
			assert.Equal(t, output.Name(), pathErr.Path)
		})
	}
}

func TestValidateActionWritesText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		contents        string
		expectedError   error
		expectedHeading string
		expectedMessage string
	}{
		{
			name:            "valid policy",
			contents:        "version: v1\n",
			expectedHeading: "VALID: true\n",
		},
		{
			name:            "invalid policy",
			contents:        "linters:\n  unknown: true\n",
			expectedError:   ErrHasValidateIssue,
			expectedHeading: "VALID: false\nERRORS:\n",
			expectedMessage: "yaml_validation",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "easyp.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o600))
			var output bytes.Buffer
			ctx := validationContext(t, path, flags.TextFormat, &output)

			err := (Validate{}).Action(ctx)

			require.ErrorIs(t, err, tt.expectedError)
			assert.True(t, bytes.HasPrefix(output.Bytes(), []byte(tt.expectedHeading)), output.String())
			if tt.expectedMessage != "" {
				assert.Contains(t, output.String(), "easyp.yaml")
				assert.Contains(t, output.String(), "2:3")
				assert.Contains(t, output.String(), tt.expectedMessage)
				return
			}
			assert.Equal(t, tt.expectedHeading, output.String())
		})
	}
}

func TestValidateActionWritesJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		contents      string
		expectedError error
		expectedValid bool
	}{
		{name: "valid policy", contents: "version: v1\n", expectedValid: true},
		{name: "invalid policy", contents: "linters:\n  unknown: true\n", expectedError: ErrHasValidateIssue},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "easyp.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o600))
			var output bytes.Buffer
			ctx := validationContext(t, path, flags.JSONFormat, &output)

			err := (Validate{}).Action(ctx)

			require.ErrorIs(t, err, tt.expectedError)
			var report validateResult
			require.NoError(t, json.Unmarshal(output.Bytes(), &report))
			assert.Equal(t, tt.expectedValid, report.Valid)
			assert.Empty(t, report.Warnings)
			if tt.expectedValid {
				assert.Empty(t, report.Errors)
				return
			}
			require.Len(t, report.Errors, 1)
			assert.Equal(t, "easyp.yaml", report.Errors[0].File)
			assert.Equal(t, "yaml_validation", report.Errors[0].Code)
			assert.Equal(t, config.SeverityError, report.Errors[0].Severity)
			assert.Equal(t, 2, report.Errors[0].Line)
			assert.Equal(t, 3, report.Errors[0].Column)
		})
	}
}

func TestWriteValidationText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		report   validateResult
		expected string
	}{
		{
			name: "warnings preserve a valid result",
			report: validateResult{
				Valid: true,
				Warnings: []config.ValidationIssue{
					{File: "easyp.yaml", Code: "warning", Message: "first warning"},
					{File: "nested/easyp.yaml", Line: 3, Column: 2, Code: "warning", Message: "second warning"},
				},
			},
			expected: "VALID: true\nWARNINGS:\n# FILE LOCATION CODE MESSAGE\n1 easyp.yaml - warning first warning\n2 nested/easyp.yaml 3:2 warning second warning",
		},
		{
			name: "errors precede separately numbered warnings",
			report: validateResult{
				Errors: []config.ValidationIssue{
					{File: "easyp.yaml", Line: 2, Column: 3, Code: "invalid", Message: "invalid field"},
				},
				Warnings: []config.ValidationIssue{
					{File: "easyp.gen.yaml", Code: "warning", Message: "check field"},
				},
			},
			expected: "VALID: false\nERRORS:\n# FILE LOCATION CODE MESSAGE\n1 easyp.yaml 2:3 invalid invalid field\nWARNINGS:\n# FILE LOCATION CODE MESSAGE\n1 easyp.gen.yaml - warning check field",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			err := writeValidationText(&output, tt.report)

			require.NoError(t, err)
			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			for i, line := range lines {
				lines[i] = strings.Join(strings.Fields(line), " ")
			}
			assert.Equal(t, tt.expected, strings.Join(lines, "\n"))
		})
	}
}

func TestWriteValidationTextPreservesOutputErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		byteLimit int
		operation string
	}{
		{name: "status line", operation: "Fprintf:"},
		{name: "section heading", byteLimit: len("VALID: false\n"), operation: "Fprintln:"},
		{name: "buffered rows", byteLimit: len("VALID: false\nERRORS:\n"), operation: "Flush:"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := &limitedValidationWriter{remaining: tt.byteLimit}
			report := validateResult{
				Errors: []config.ValidationIssue{{File: "easyp.yaml", Code: "invalid", Message: "invalid field"}},
			}

			err := writeValidationText(output, report)

			require.ErrorIs(t, err, io.ErrClosedPipe)
			assert.ErrorContains(t, err, tt.operation)
		})
	}
}

type limitedValidationWriter struct {
	remaining int
}

func (w *limitedValidationWriter) Write(p []byte) (int, error) {
	n := min(len(p), w.remaining)
	w.remaining -= n
	if n < len(p) {
		return n, io.ErrClosedPipe
	}
	return n, nil
}

func validationContext(t *testing.T, path, format string, output io.Writer) *cli.Context {
	t.Helper()

	set := flag.NewFlagSet("validate-config", flag.ContinueOnError)
	set.String(flags.Config.Name, "", "")
	set.String(flags.Format.Name, "", "")
	require.NoError(t, set.Set(flags.Config.Name, path))
	require.NoError(t, set.Set(flags.Format.Name, format))
	ctx := cli.NewContext(&cli.App{Writer: output}, set, nil)
	ctx.Context = t.Context()
	return ctx
}
