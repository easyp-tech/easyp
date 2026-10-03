package modules

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestLocalOverlayGraph(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, aManifest, bManifest, wantError string
		absolute                              bool
	}{
		{name: "root replacements apply transitively", aManifest: "require example.com/B v1.0.0\n"},
		{name: "absolute root replacements", aManifest: "require example.com/B\n", absolute: true},
		{name: "ignore dependency replacements", aManifest: "require example.com/B\nreplace example.com/B => nonexistent\n"},
		{name: "ordinary dependency cycle deduplicates", aManifest: "require example.com/B\n", bManifest: "require example.com/A\n"},
		{name: "identity mismatch", aManifest: "require example.com/B\n", bManifest: "module example.com/wrong\n", wantError: "want example.com/B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "a/protobuf.mod", "module example.com/A\n"+tt.aManifest)
			b := "module example.com/B\n" + tt.bManifest
			if tt.wantError != "" {
				b = tt.bManifest
			}
			writeV1GenerateFixture(t, root, "b/protobuf.mod", b)
			target := "b"
			if tt.absolute {
				target = filepath.Join(root, "b")
			}
			module := v1.Module{Name: "example.com/app", Roots: []string{"proto"}, Requires: []v1.Requirement{{Module: "example.com/A"}}, Replaces: []v1.Replacement{
				{Module: "example.com/A", Target: "a"}, {Module: "example.com/B", Target: target}, {Module: "example.com/unused", Target: "missing"},
			}}
			repository := &fakeRepository{fetchErr: errors.New("unexpected remote fetch"), installErr: errors.New("unexpected remote install")}
			roots, err := EnsureSources(t.Context(), root, module, repository)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{filepath.Join(root, "a"), filepath.Join(root, "b")}, roots.Paths())
			_, err = os.Stat(filepath.Join(root, v1.LockFile))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestLocalOverlayOperationsPreserveFiles(t *testing.T) {
	t.Parallel()
	for _, existing := range []bool{false, true} {
		tests := []struct {
			name string
			run  func(*testing.T, string, *fakeRepository) error
		}{
			{name: "tidy", run: func(t *testing.T, root string, r *fakeRepository) error { return Tidy(t.Context(), root, r) }},
			{name: "download", run: func(t *testing.T, root string, r *fakeRepository) error { return Download(t.Context(), root, r) }},
			{name: "get", run: func(t *testing.T, root string, r *fakeRepository) error {
				return Get(t.Context(), root, v1.Requirement{Module: "example.com/A"}, r)
			}},
			{name: "update", run: func(t *testing.T, root string, r *fakeRepository) error { return Update(t.Context(), root, r) }},
		}
		for _, tt := range tests {
			t.Run(tt.name+map[bool]string{false: " without lock", true: " with lock"}[existing], func(t *testing.T) {
				t.Parallel()
				root := t.TempDir()
				manifest := "module example.com/app\nroots proto\nrequire example.com/A v1.0.0 // keep\nreplace example.com/A => a\n"
				lock := "# keep these bytes\nversion: 1\nmodules: []\n"
				writeV1GenerateFixture(t, root, v1.ModuleFile, manifest)
				writeV1GenerateFixture(t, root, "proto/app.proto", "syntax = \"proto3\"; import \"a.proto\";\n")
				writeV1GenerateFixture(t, root, "a/protobuf.mod", "module example.com/A\n")
				writeV1GenerateFixture(t, root, "a/a.proto", "syntax = \"proto3\";\n")
				if existing {
					writeV1GenerateFixture(t, root, v1.LockFile, lock)
				}
				repository := &fakeRepository{fetchErr: errors.New("unexpected fetch"), installErr: errors.New("unexpected install"), versionsErr: errors.New("unexpected tag lookup")}
				require.NoError(t, tt.run(t, root, repository))
				raw, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
				require.NoError(t, err)
				assert.Equal(t, manifest, string(raw))
				raw, err = os.ReadFile(filepath.Join(root, v1.LockFile))
				if existing {
					require.NoError(t, err)
					assert.Equal(t, lock, string(raw))
				} else {
					assert.ErrorIs(t, err, os.ErrNotExist)
				}
			})
		}
	}
}
