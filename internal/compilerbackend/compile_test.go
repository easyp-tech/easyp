package compilerbackend

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// The want descriptors in testdata were produced once by protocompile, the
// compiler easyp replaced, so the backend keeps matching its output.
func TestBackendsProduceSameDescriptors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
		roots []string
		// protocompile emitted one span for json_name, easyp-protoc emits protoc's two.
		differentSourceInfo bool
	}{
		{
			name: "import_paths_and_wkt",
			files: map[string]string{
				"a/common/v1/common.proto": `syntax = "proto3"; package common.v1; import "google/protobuf/timestamp.proto"; message Common { google.protobuf.Timestamp at = 1; }`,
				"b/api/v1/api.proto":       `syntax = "proto3"; package api.v1; import "common/v1/common.proto"; message Req { common.v1.Common common = 1; } service Api { rpc Get(Req) returns (Req); }`,
			},
			roots: []string{"b", "a"},
		},
		{
			name: "field_shadows_package",
			files: map[string]string{
				"svc.proto":  `syntax = "proto3"; package service; message Service {}`,
				"main.proto": `syntax = "proto3"; import "svc.proto"; message C { service.Service service = 1; }`,
			},
		},
		{
			name: "comments",
			files: map[string]string{
				"main.proto": "syntax = \"proto3\";\npackage p;\n// Leading.\nmessage M {\n  // Field.\n  string a_b = 1; // Trailing.\n}\n",
			},
		},
		{
			name: "json_name",
			files: map[string]string{
				"main.proto": "syntax = \"proto3\";\npackage p;\nmessage M {\n  string a_b = 1 [json_name = \"ab\"];\n}\n",
			},
			differentSourceInfo: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			var targets []string
			for name, src := range tc.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(src), 0o644))
				target := name
				for _, root := range tc.roots {
					target = strings.TrimPrefix(target, root+"/")
				}
				targets = append(targets, target)
			}
			roots := []string{dir}
			if len(tc.roots) > 0 {
				roots = nil
				for _, root := range tc.roots {
					roots = append(roots, filepath.Join(dir, root))
				}
			}
			req := Request{
				Files:           targets,
				ImportPaths:     roots,
				Accessor:        func(path string) (io.ReadCloser, error) { return os.Open(path) },
				StandardImports: true,
				SourceInfo:      true,
			}

			raw, err := os.ReadFile(filepath.Join("testdata", tc.name+".binpb"))
			require.NoError(t, err)
			want := &descriptorpb.FileDescriptorSet{}
			require.NoError(t, proto.Unmarshal(raw, want))
			got, err := compileEasyp(context.Background(), req)
			require.NoError(t, err)

			require.Len(t, got.Files, len(want.File))
			for _, wantProto := range want.File {
				gotProto := got.Protos[wantProto.GetName()]
				require.NotNil(t, gotProto, wantProto.GetName())
				wantInfo, gotInfo := wantProto.GetSourceCodeInfo(), gotProto.GetSourceCodeInfo()
				wantProto, gotProto = proto.CloneOf(wantProto), proto.CloneOf(gotProto)
				wantProto.SourceCodeInfo, gotProto.SourceCodeInfo = nil, nil
				assert.True(t, proto.Equal(wantProto, gotProto), "descriptor mismatch for %s", wantProto.GetName())
				assert.Equal(t, !tc.differentSourceInfo, proto.Equal(wantInfo, gotInfo), "source info equality for %s", wantProto.GetName())
			}
		})
	}
}

func TestEasypReportsDiagnostics(t *testing.T) {
	t.Parallel()

	req := Request{
		Files:    []string{"bad.proto"},
		Accessor: BytesAccessor(func(string) ([]byte, error) { return []byte(`syntax = "proto3"; message M { Missing m = 1; }`), nil }),
	}
	_, err := compileEasyp(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"Missing" is not defined`)
}

func TestDecodeDescriptorSetMatchesUnmarshal(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("testdata")
	require.NoError(t, err)
	var all []byte
	for _, entry := range entries {
		wire, err := os.ReadFile(filepath.Join("testdata", entry.Name()))
		require.NoError(t, err)
		all = append(all, wire...)
		var want descriptorpb.FileDescriptorSet
		require.NoError(t, proto.Unmarshal(wire, &want))
		got, err := decodeDescriptorSet(t.Context(), wire)
		require.NoError(t, err)
		assert.True(t, proto.Equal(&want, got), entry.Name())
	}
	// Concatenated sets are one set; a non-file field falls back to Unmarshal.
	withOther := append(append([]byte(nil), all...), 0x10, 0x01)
	for _, wire := range [][]byte{all, withOther, nil} {
		var want descriptorpb.FileDescriptorSet
		require.NoError(t, proto.Unmarshal(wire, &want))
		got, err := decodeDescriptorSet(t.Context(), wire)
		require.NoError(t, err)
		assert.True(t, proto.Equal(&want, got))
	}
	_, err = decodeDescriptorSet(t.Context(), []byte{0x0a, 0x05, 0x01})
	require.Error(t, err)
}
