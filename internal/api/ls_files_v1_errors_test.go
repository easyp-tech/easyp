package api

import (
	"bytes"
	"encoding/json"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/flags"
)

// Action reads the process cwd, so these CLI cases must run sequentially.
func TestLsFilesActionReportsErrorsBeforeFailing(t *testing.T) {
	tests := []struct {
		name           string
		format         string
		includeImports bool
		broken         bool
	}{
		{name: "json_errors", format: flags.JSONFormat, includeImports: true, broken: true},
		{name: "text_errors", format: flags.TextFormat, includeImports: true, broken: true},
		{name: "json_success", format: flags.JSONFormat, includeImports: true},
		{name: "text_success", format: flags.TextFormat, includeImports: true},
		{name: "local_only_ignores_import_errors", format: flags.JSONFormat, broken: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\n")
			writeV1GenerateFixture(t, root, "valid.proto", `syntax = "proto3"; import "google/protobuf/timestamp.proto";`)
			if tt.broken {
				writeV1GenerateFixture(t, root, "missing.proto", `syntax = "proto3"; import "absent.proto"; import "google/protobuf/not_real.proto";`)
				writeV1GenerateFixture(t, root, "syntax.proto", `syntax = "proto3"; message Broken {`)
				writeV1GenerateFixture(t, root, "traversal.proto", `syntax = "proto3"; import "nested/../valid.proto";`)
			}
			t.Chdir(root)
			var output, diagnostics bytes.Buffer
			app := cli.NewApp()
			app.Writer, app.ErrWriter = &output, &diagnostics
			set := flag.NewFlagSet("ls-files", flag.ContinueOnError)
			set.Bool(flagLsFilesIncludeImports.Name, tt.includeImports, "")
			set.String(flags.Format.Name, flags.JSONFormat, "")
			require.NoError(t, set.Set(flags.Format.Name, tt.format))
			ctx := cli.NewContext(app, set, nil)
			ctx.Context = t.Context()

			err := (LsFiles{}).Action(ctx)

			wantErrors := tt.broken && tt.includeImports
			if wantErrors {
				var exitErr cli.ExitCoder
				require.ErrorAs(t, err, &exitErr)
				assert.Equal(t, 1, exitErr.ExitCode())
			} else {
				assert.NoError(t, err)
			}
			if tt.format == flags.TextFormat {
				assert.Contains(t, output.String(), "valid.proto\tworkspace\t")
				assert.Contains(t, output.String(), "google/protobuf/timestamp.proto\twellknown\t")
				if wantErrors {
					assert.Contains(t, diagnostics.String(), "import_not_found: missing.proto imports \"absent.proto\"")
					assert.Contains(t, diagnostics.String(), "google/protobuf/not_real.proto")
					assert.Contains(t, diagnostics.String(), "parse_error: syntax.proto")
					assert.Contains(t, diagnostics.String(), "invalid_import: traversal.proto")
				} else {
					assert.Empty(t, diagnostics.String())
				}
				return
			}
			var result v1ListResult
			require.NoError(t, json.Unmarshal(output.Bytes(), &result))
			assert.NotEmpty(t, result.Roots)
			var paths []string
			for _, file := range result.Files {
				paths = append(paths, file.ImportPath)
			}
			assert.Contains(t, paths, "valid.proto")
			if tt.includeImports {
				assert.Contains(t, paths, "google/protobuf/timestamp.proto")
			} else {
				assert.NotContains(t, paths, "google/protobuf/timestamp.proto")
			}
			if wantErrors {
				assert.Len(t, result.Errors, 4)
				assert.Contains(t, result.Errors, v1ListError{Code: "import_not_found", Message: `missing.proto imports "absent.proto"`})
				assert.Contains(t, result.Errors, v1ListError{Code: "import_not_found", Message: `missing.proto imports "google/protobuf/not_real.proto"`})
				assert.Contains(t, result.Errors, v1ListError{Code: "invalid_import", Message: `traversal.proto imports "nested/../valid.proto"`})
			} else {
				assert.Empty(t, result.Errors)
			}
			assert.Empty(t, diagnostics.String())
		})
	}
}

func TestListV1FilesRejectsInvalidAlreadyListedImport(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "main.proto", `syntax = "proto3"; import "invalid:name.proto";`)
	writeV1GenerateFixture(t, root, "invalid:name.proto", `syntax = "proto3";`)
	module := v1.Module{Name: "example.com/app", Roots: []string{"."}}

	result, err := listV1Files(t.Context(), root, module, true, nil)

	require.NoError(t, err)
	assert.Equal(t, []v1ListError{{Code: "invalid_import", Message: `main.proto imports "invalid:name.proto"`}}, result.Errors)
}
