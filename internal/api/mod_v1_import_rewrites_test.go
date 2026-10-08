package api

import (
	"bytes"
	"flag"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestModTidyReportsOnlyCommittedImportMappings(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantErr  bool
	}{
		{name: "committed_edits", typeName: "service.Service"},
		{name: "failed_compile_has_no_applied_report", typeName: "service.Missing", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The API reads process cwd and EASYPPATH, so these cases are sequential.
			root, remote, storage := t.TempDir(), t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, remote, "api/v1/svc.proto", `syntax = "proto3"; package service; message Service {}`)
			runTestGit(t, remote, "init", "-q")
			runTestGit(t, remote, "add", ".")
			runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "old contract")
			runTestGit(t, remote, "tag", "v0.4.0")
			writeV1GenerateFixture(t, root, v1.ModuleFile, "module example.test/consumer\n")
			cache := gitmodules.New(storage)
			require.NoError(t, modules.GetWithRoots(t.Context(), root, v1.Requirement{Module: remote, Version: "v0.4.0"}, cache, []string{"api/v1"}))
			consumer := `syntax = "proto3"; import "svc.proto"; message Consumer { ` + tt.typeName + ` service = 1; }`
			writeV1GenerateFixture(t, root, "consumer.proto", consumer)
			writeV1GenerateFixture(t, root, "build/copy.proto", consumer)
			writeV1GenerateFixture(t, remote, v1.ModuleFile, "module "+remote+"\nroots api\n")
			runTestGit(t, remote, "add", ".")
			runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "declare root")
			runTestGit(t, remote, "tag", "v0.5.0")
			manifest, _, err := modules.ReadManifest(root)
			require.NoError(t, err)
			writeV1GenerateFixture(t, root, v1.ModuleFile, strings.ReplaceAll(string(manifest), "v0.4.0", "v0.5.0"))
			t.Chdir(root)
			t.Setenv("EASYPPATH", storage)
			var output bytes.Buffer
			ctx := cli.NewContext(&cli.App{Writer: &output, Metadata: map[string]any{}}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
			ctx.Context = t.Context()

			err = (Mod{}).Tidy(ctx)

			if tt.wantErr {
				require.ErrorContains(t, err, "unknown type")
				assert.Empty(t, output.String())
				return
			}
			require.NoError(t, err)
			text := output.String()
			assert.Contains(t, text, `build/copy.proto: import "svc.proto" -> "v1/svc.proto"`)
			assert.Contains(t, text, `consumer.proto: import "svc.proto" -> "v1/svc.proto"`)
			assert.Less(t, strings.Index(text, "build/copy.proto:"), strings.Index(text, "consumer.proto:"))
			for _, detail := range []string{remote, "v0.4.0", "v0.5.0", "[api/v1]", "[api]", "generation selectors", "SDK"} {
				assert.Contains(t, text, detail)
			}
			lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
			require.NoError(t, err)
			assert.Equal(t, "v0.5.0", lock.Modules[0].Version)
			output.Reset()
			require.NoError(t, (Mod{}).Tidy(ctx))
			assert.Empty(t, output.String())
		})
	}
}
