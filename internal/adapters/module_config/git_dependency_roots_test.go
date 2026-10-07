package moduleconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitDependencyRootsProvenance(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name          string
		files         map[string]string
		authoritative bool
	}{
		{name: "no metadata"},
		{name: "legacy requirements", files: map[string]string{"protobuf.mod": "direct (\n  example.test/common@v1.0.0\n)\n"}},
		{name: "legacy deps only", files: map[string]string{"easyp.yaml": "deps: [example.test/common@v1.0.0]\n"}},
		{name: "legacy Git inputs only", files: map[string]string{"easyp.yaml": "generate:\n  inputs:\n    - git_repo: {url: example.test/common@v1.0.0}\n"}},
		{name: "native implicit default", files: map[string]string{"protobuf.mod": "module example.test/dependency\n"}, authoritative: true},
		{name: "native explicit default", files: map[string]string{"protobuf.mod": "module example.test/dependency\nroots .\n"}, authoritative: true},
		{name: "Buf default", files: map[string]string{"buf.yaml": "version: v1\n"}, authoritative: true},
		{name: "Buf workspace default", files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [.]\n"}, authoritative: true},
		{name: "legacy directory default", files: map[string]string{"easyp.yaml": "generate:\n  inputs:\n    - directory: api\n"}, authoritative: true},
		{name: "nested native default", files: map[string]string{"api/protobuf.mod": "module example.test/dependency\n"}, authoritative: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for name, content := range tt.files {
				filename := filepath.Join(directory, filepath.FromSlash(name))
				require.NoError(t, os.MkdirAll(filepath.Dir(filename), 0o755))
				require.NoError(t, os.WriteFile(filename, []byte(content), 0o644))
			}
			module, err := ReadGitDependency(directory, "example.test/dependency")
			require.NoError(t, err)
			assert.Equal(t, tt.authoritative, module.RootsFromMetadata)
		})
	}
}

func TestGitDependencyRejectsMalformedRootMetadata(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, file, content string
	}{
		{name: "numeric Buf path", file: "buf.yaml", content: "version: v2\nmodules: [{path: 123}]\n"},
		{name: "merged numeric Buf path", file: "buf.yaml", content: "version: v2\nmodules: [{<<: {path: 123}}]\n"},
		{name: "numeric workspace directory", file: "buf.work.yaml", content: "version: v1\ndirectories: [123]\n"},
		{name: "null Buf roots", file: "buf.yaml", content: "version: v1beta1\nbuild: {roots: null}\n"},
		{name: "null legacy directory", file: "easyp.yaml", content: "generate:\n  inputs: [{directory: null}]\n"},
		{name: "numeric legacy root", file: "easyp.yaml", content: "generate:\n  inputs: [{directory: {path: api, root: 123}}]\n"},
		{name: "merged numeric legacy root", file: "easyp.yaml", content: "generate:\n  inputs: [{directory: {path: api, <<: {root: 123}}}]\n"},
		{name: "multiple Buf documents", file: "buf.yaml", content: "version: v1\n---\nversion: v2\n"},
		{name: "multiple EasyP documents", file: "easyp.yaml", content: "deps: []\n---\ngenerate:\n  inputs: [{directory: api}]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(directory, tt.file), []byte(tt.content), 0o644))
			_, err := ReadGitDependency(directory, "example.test/dependency")
			require.Error(t, err)
		})
	}
}
