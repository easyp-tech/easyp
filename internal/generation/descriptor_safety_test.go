package generation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestDescriptorOutputFlagsAreExclusive(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: "all.pb", DescriptorSetOutDir: "sets"})
	require.ErrorContains(t, err, "mutually exclusive")
	files, err := os.ReadDir(root)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestDescriptorOutputPreflight(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		parentFile      bool
		sameIdentity    bool
		duplicateTarget bool
	}{
		{name: "parent is a file", parentFile: true},
		{name: "different sources claim one identity", sameIdentity: true},
		{name: "duplicate selection runs once", duplicateTarget: true},
		{name: "same basename different identities"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeDescriptorMarker(t, root)
			modules := "[a, b]"
			if tt.duplicateTarget {
				modules = "[a, a]"
			}
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  modules: "+modules+"\n"+descriptorMarkerConfig)
			for _, module := range []string{"a", "b"} {
				identity := module + ".example.com/same"
				if tt.sameIdentity {
					identity = "example.com/same"
				}
				writeV1GenerateFixture(t, root, module+"/protobuf.mod", "module "+identity+"\n")
				writeV1GenerateFixture(t, root, module+"/item.proto", "syntax = \"proto3\"; package "+module+"; message Item {}")
			}
			outDir := filepath.Join(root, "sets")
			if tt.parentFile {
				require.NoError(t, os.WriteFile(outDir, []byte("keep"), 0o600))
			}
			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOutDir: outDir})
			if tt.parentFile || tt.sameIdentity {
				require.Error(t, err)
				assert.NoFileExists(t, filepath.Join(root, "plugin-ran.txt"))
				if tt.sameIdentity {
					assert.Contains(t, err.Error(), "output collision")
				}
				return
			}
			require.NoError(t, err)
			want := 2
			if tt.duplicateTarget {
				want = 1
			}
			assert.Len(t, descriptorTree(t, outDir), want)
			calls, err := os.ReadFile(filepath.Join(root, "plugin-ran.txt"))
			require.NoError(t, err)
			assert.Len(t, calls, want)
		})
	}
}

func TestDescriptorExportUsesPreparedGraph(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeDescriptorMarker(t, root)
	writeV1GenerateFixture(t, root, "marker.sh", "#!/bin/sh\ncat >/dev/null\nprintf x >> plugin-ran.txt\nprintf invalid > b/item.proto\n")
	writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  modules: [a, b]\n"+descriptorMarkerConfig)
	for _, module := range []string{"a", "b"} {
		writeV1GenerateFixture(t, root, module+"/protobuf.mod", "module example.com/"+module+"\n")
		writeV1GenerateFixture(t, root, module+"/item.proto", "syntax = \"proto3\"; package "+module+"; message Item {}")
	}
	outDir := filepath.Join(root, "sets")
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOutDir: outDir, IncludeImports: true}))
	assert.Len(t, descriptorTree(t, outDir), 2)
	calls, err := os.ReadFile(filepath.Join(root, "plugin-ran.txt"))
	require.NoError(t, err)
	assert.Equal(t, "xx", string(calls))
}
