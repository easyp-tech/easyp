package moduleconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestReadGitDependencyBSRMetadata(t *testing.T) {
	t.Parallel()
	commit := strings.Repeat("a", 32)
	digest := "shake256:" + strings.Repeat("a", 128)
	tests := []struct {
		name  string
		files map[string]string
		want  []v1.BSRDependency
	}{
		{
			name:  "v1 without lock",
			files: map[string]string{"buf.yaml": "version: v1\ndeps: [buf.build/googleapis/googleapis:stable]\n"},
			want:  []v1.BSRDependency{{Module: "buf.build/googleapis/googleapis", Reference: "stable", Config: "buf.yaml"}},
		},
		{
			name:  "duplicate declaration",
			files: map[string]string{"buf.yaml": "version: v1\ndeps: [buf.build/googleapis/googleapis, buf.build/googleapis/googleapis]\n"},
			want:  []v1.BSRDependency{{Module: "buf.build/googleapis/googleapis", Config: "buf.yaml"}},
		},
		{
			name: "v1beta1 with lock",
			files: map[string]string{
				"buf.yaml": "version: v1beta1\nbuild: {roots: [proto]}\ndeps: [buf.build/googleapis/googleapis:stable]\n",
				"buf.lock": "version: v1beta1\ndeps:\n  - remote: buf.build\n    owner: googleapis\n    repository: googleapis\n    commit: " + commit + "\n    digest: " + digest + "\n",
			},
			want: []v1.BSRDependency{{Module: "buf.build/googleapis/googleapis", Reference: "stable", Commit: commit, Digest: digest, Config: "buf.yaml"}},
		},
		{
			name: "v2 locked transitive dependencies",
			files: map[string]string{
				"buf.yaml": "version: v2\nmodules: [{path: proto}]\ndeps: [buf.build/googleapis/googleapis]\n",
				"buf.lock": "version: v2\ndeps:\n  - name: buf.build/googleapis/googleapis\n    commit: " + commit + "\n  - name: buf.build/bufbuild/protovalidate\n    commit: " + strings.Repeat("b", 32) + "\n",
			},
			want: []v1.BSRDependency{
				{Module: "buf.build/googleapis/googleapis", Commit: commit, Config: "buf.yaml"},
				{Module: "buf.build/bufbuild/protovalidate", Commit: strings.Repeat("b", 32), Config: "buf.yaml"},
			},
		},
		{
			name: "workspace preserves both origins",
			files: map[string]string{
				"buf.work.yaml":    "version: v1\ndirectories: [proto/a, proto/b]\n",
				"proto/a/buf.yaml": "version: v1\ndeps: [buf.build/googleapis/googleapis]\n",
				"proto/a/buf.lock": "version: v1\ndeps: [{remote: buf.build, owner: googleapis, repository: googleapis, commit: " + commit + "}]\n",
				"proto/b/buf.yaml": "version: v1\ndeps: [buf.build/googleapis/googleapis:stable]\n",
			},
			want: []v1.BSRDependency{
				{Module: "buf.build/googleapis/googleapis", Commit: commit, Config: "proto/a/buf.yaml"},
				{Module: "buf.build/googleapis/googleapis", Reference: "stable", Config: "proto/b/buf.yaml"},
			},
		},
		{
			name:  "native manifest takes precedence",
			files: map[string]string{"protobuf.mod": "module github.com/acme/parent\n", "buf.yaml": "version: v1\ndeps: [invalid]\n", "buf.lock": "broken YAML: ["},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for name, contents := range tt.files {
				path := filepath.Join(root, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
			}

			module, err := ReadGitDependency(root, "github.com/acme/parent")

			require.NoError(t, err)
			assert.Equal(t, tt.want, module.BSRDependencies)
			assert.Empty(t, module.Requires)
		})
	}
}

func TestReadBufDependencyRejectsInvalidBSRMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    string
		lock    string
		wantErr string
	}{
		{name: "invalid identity", deps: "[github.com/googleapis]", wantErr: "invalid BSR module"},
		{name: "empty reference", deps: "[buf.build/googleapis/googleapis:]", wantErr: "empty BSR reference"},
		{name: "conflicting declarations", deps: "[buf.build/googleapis/googleapis:stable, buf.build/googleapis/googleapis:next]", wantErr: "conflicting BSR references"},
		{name: "malformed lock", deps: "[buf.build/googleapis/googleapis]", lock: "version: [", wantErr: "Decode"},
		{name: "unsupported lock format", deps: "[buf.build/googleapis/googleapis]", lock: "version: v3\n", wantErr: "unsupported buf.lock version"},
		{name: "missing pin", deps: "[buf.build/googleapis/googleapis]", lock: "version: v2\ndeps: []\n", wantErr: "missing from buf.lock"},
		{name: "lock without declarations", deps: "[]", lock: "version: v2\ndeps: [{name: buf.build/googleapis/googleapis, commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}]\n", wantErr: "declares none"},
		{name: "empty commit", deps: "[buf.build/googleapis/googleapis]", lock: "version: v2\ndeps: [{name: buf.build/googleapis/googleapis}]\n", wantErr: "missing BSR commit"},
		{name: "duplicate lock entries", deps: "[buf.build/googleapis/googleapis]", lock: "version: v2\ndeps: [{name: buf.build/googleapis/googleapis, commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}, {name: buf.build/googleapis/googleapis, commit: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa}]\n", wantErr: "duplicate BSR pin"},
		{name: "multiple lock documents", deps: "[buf.build/googleapis/googleapis]", lock: "version: v2\ndeps: []\n---\nversion: v2\n", wantErr: "one YAML document"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "buf.yaml")
			require.NoError(t, os.WriteFile(path, []byte("version: v1\ndeps: "+tt.deps+"\n"), 0o600))
			if tt.lock != "" {
				require.NoError(t, os.WriteFile(filepath.Join(root, "buf.lock"), []byte(tt.lock), 0o600))
			}

			_, err := readBufDependencyModule(path)

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestReadBufDependencyRejectsNonRegularLock(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		link bool
	}{
		{name: "symlink", link: true},
		{name: "directory"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			config := filepath.Join(root, "buf.yaml")
			require.NoError(t, os.WriteFile(config, []byte("version: v1\n"), 0o600))
			lockPath := filepath.Join(root, "buf.lock")
			if tt.link {
				require.NoError(t, os.WriteFile(filepath.Join(root, "pin.yaml"), []byte("version: v1\ndeps: []\n"), 0o600))
				require.NoError(t, os.Symlink("pin.yaml", lockPath))
			} else {
				require.NoError(t, os.Mkdir(lockPath, 0o755))
			}

			_, err := readBufDependencyModule(config)

			require.ErrorContains(t, err, "non-regular dependency config")
			assert.Contains(t, err.Error(), "buf.lock")
		})
	}
}
