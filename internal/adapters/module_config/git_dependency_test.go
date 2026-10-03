package moduleconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/stretchr/testify/require"
)

func TestReadGitDependencyModuleFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string]string
		roots    []string
		requires []v1.Requirement
	}{
		{
			name:  "no dependency metadata",
			roots: []string{"."},
		},
		{
			name: "buf v2",
			files: map[string]string{
				"buf.yaml": "version: v2\nmodules:\n  - path: proto/user\n  - path: proto/common\n",
			},
			roots: []string{"proto/user", "proto/common"},
		},
		{
			name: "buf v1 workspace",
			files: map[string]string{
				"buf.work.yaml":         "version: v1\ndirectories:\n  - proto/user\n  - proto/common\n",
				"proto/user/buf.yaml":   "version: v1\n",
				"proto/common/buf.yaml": "version: v1\n",
			},
			roots: []string{"proto/user", "proto/common"},
		},
		{
			name:  "buf v1 single module",
			files: map[string]string{"buf.yaml": "version: v1\n"},
			roots: []string{"."},
		},
		{
			name:  "buf v1beta1 roots",
			files: map[string]string{"buf.yaml": "version: v1beta1\nbuild:\n  roots: [proto, third_party]\n"},
			roots: []string{"proto", "third_party"},
		},
		{
			name: "legacy easyp and protobuf.mod",
			files: map[string]string{
				"easyp.yaml":   "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n",
				"protobuf.mod": "direct (\n  example.com/common@v1.2.0\n)\n",
			},
			roots:    []string{"proto"},
			requires: []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}},
		},
		{
			name: "legacy git input",
			files: map[string]string{
				"easyp.yaml": "generate:\n  inputs:\n    - directory: proto\n    - git_repo:\n        url: example.com/common@v1.1.0\n",
			},
			roots:    []string{"."},
			requires: []v1.Requirement{{Module: "example.com/common", Version: "v1.1.0"}},
		},
		{
			name: "v1 manifest wins",
			files: map[string]string{
				"protobuf.mod": "module example.com/dependency\nroots proto\nrequire example.com/common v1.2.0\n",
				"buf.yaml":     "invalid: [lower priority]",
				"easyp.yaml":   "generate: [invalid legacy config]",
			},
			roots:    []string{"proto"},
			requires: []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}},
		},
		{
			name: "legacy requirements and buf roots are combined",
			files: map[string]string{
				"protobuf.mod":  "direct (\n  example.com/common@v1.2.0\n)\n",
				"buf.work.yaml": "version: v1\ndirectories: [proto/buf]\n",
				"buf.yaml":      "invalid: [lower priority]",
				"easyp.yaml":    "generate:\n  inputs:\n    - directory:\n        path: proto/legacy\n        root: proto/legacy\n    - git_repo:\n        url: example.com/other@v1.3.0\n",
			},
			roots: []string{"proto/buf"},
			requires: []v1.Requirement{
				{Module: "example.com/common", Version: "v1.2.0"},
				{Module: "example.com/other", Version: "v1.3.0"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for name, body := range tc.files {
				path := filepath.Join(dir, name)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			module, err := ReadGitDependency(dir, "example.com/dependency")
			require.NoError(t, err)
			require.Equal(t, "example.com/dependency", module.Name)
			require.Equal(t, tc.roots, module.Roots)
			require.Equal(t, tc.requires, module.Requires)
		})
	}
}

func TestReadGitDependencyRejectsInvalidBufWorkspaceBeforeFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, bufWorkConfigFile), []byte("version: v2\ndirectories: [proto]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, bufModuleConfigFile), []byte("version: v1\n"), 0o644))
	_, err := ReadGitDependency(dir, "example.com/dependency")
	require.ErrorContains(t, err, bufWorkConfigFile)
}

func TestDetectGitDependencyModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files []string
		want  []gitDependencyMode
	}{
		{name: "missing files"},
		{
			name:  "all formats in stable order",
			files: []string{legacyEasyPConfigFile, bufModuleConfigFile, dependencyManifestFile, bufWorkConfigFile},
			want:  []gitDependencyMode{gitDependencyManifest, gitDependencyBufWorkspace, gitDependencyBufModule, gitDependencyLegacyEasyP},
		},
		{
			name:  "nested manifest is not a root mode",
			files: []string{"nested/" + dependencyManifestFile},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for _, name := range tc.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(""), 0o644))
			}
			modes, err := detectGitDependencyModes(dir)
			require.NoError(t, err)
			require.Equal(t, tc.want, modes)
		})
	}
}

func TestReadGitDependencyReportsUnreadableSelectedConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, bufModuleConfigFile), 0o755))
	_, err := ReadGitDependency(dir, "example.com/dependency")
	require.ErrorContains(t, err, bufModuleConfigFile)
}

func TestReadGitDependencySelectsNestedModuleByIdentity(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"protobuf.mod":         "module example.com/repo\nroots rootproto\n",
		"foo/protobuf.mod":     "module example.com/repo/foo\nroots proto\nrequire example.com/common v1.2.0\n",
		"bar/protobuf.mod":     "module example.com/repo/bar\nroots proto\n",
		"foo/proto/foo.proto":  "syntax = \"proto3\";\n",
		"bar/proto/bar.proto":  "syntax = \"proto3\";\n",
		"rootproto/root.proto": "syntax = \"proto3\";\n",
	} {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}

	module, err := ReadGitDependency(dir, "example.com/repo/foo")
	require.NoError(t, err)
	require.Equal(t, "example.com/repo/foo", module.Name)
	require.Equal(t, []string{"foo/proto"}, module.Roots)
	require.Equal(t, []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}}, module.Requires)

	root, err := ReadGitDependency(dir, "example.com/repo")
	require.NoError(t, err)
	require.Equal(t, []string{"rootproto"}, root.Roots)
}

func TestReadGitDependencyRejectsUnknownNestedModule(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "foo"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "foo", "protobuf.mod"), []byte("module example.com/repo/foo\n"), 0o644))
	_, err := ReadGitDependency(dir, "example.com/repo/bar")
	require.ErrorContains(t, err, "example.com/repo/bar")
}

func TestReadGitDependencyModuleRejectsEscapingRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "buf.yaml"), []byte("version: v2\nmodules:\n  - path: ../outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGitDependency(dir, "example.com/dependency"); err == nil || !strings.Contains(err.Error(), "leaves") {
		t.Fatalf("expected root traversal error, got %v", err)
	}
}

func TestReadGitDependencyBufProtoFilters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		files   map[string]string
		roots   []string
		filters []v1.ProtoFileFilter
	}{
		{
			name:    "v2 includes and excludes",
			files:   map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n    includes: [proto/api]\n    excludes: [proto/api/test]\n"},
			roots:   []string{"proto"},
			filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/api"}, Excludes: []string{"proto/api/test"}}},
		},
		{
			name:    "v2 shared root is traversed once",
			files:   map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n    includes: [proto/a]\n  - path: proto\n    includes: [proto/b]\n"},
			roots:   []string{"proto"},
			filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/a"}}, {Root: "proto", Includes: []string{"proto/b"}}},
		},
		{
			name:    "v1 excludes",
			files:   map[string]string{"buf.yaml": "version: v1\nbuild:\n  excludes: [test]\n"},
			roots:   []string{"."},
			filters: []v1.ProtoFileFilter{{Root: ".", Excludes: []string{"test"}}},
		},
		{
			name:    "v1beta1 excludes belong to their root",
			files:   map[string]string{"buf.yaml": "version: v1beta1\nbuild:\n  roots: [proto, third_party]\n  excludes: [proto/test]\n"},
			roots:   []string{"proto", "third_party"},
			filters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/test"}}, {Root: "third_party"}},
		},
		{
			name: "v1 workspace rebases nested excludes",
			files: map[string]string{
				"buf.work.yaml":  "version: v1\ndirectories: [proto]\n",
				"proto/buf.yaml": "version: v1\nbuild:\n  excludes: [test]\n",
			},
			roots:   []string{"proto"},
			filters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/test"}}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for name, body := range tt.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
			}

			module, err := ReadGitDependency(dir, "example.com/dependency")

			require.NoError(t, err)
			require.Equal(t, tt.roots, module.Roots)
			require.Equal(t, tt.filters, module.ProtoFilters)
		})
	}
}

func TestParseLegacyV1RequirementWithoutVersion(t *testing.T) {
	t.Parallel()

	got, err := parseLegacyV1Requirement("github.com/acme/common")
	if err != nil {
		t.Fatal(err)
	}
	if got.Module != "github.com/acme/common" || got.Version != "" {
		t.Fatalf("requirement = %#v", got)
	}
}

func TestParseLegacyV1RequirementRejectsBranchRef(t *testing.T) {
	t.Parallel()

	if _, err := parseLegacyV1Requirement("github.com/acme/common@main"); err == nil || !strings.Contains(err.Error(), "SemVer tag or full Git commit") {
		t.Fatalf("expected branch-ref error, got %v", err)
	}
}

func TestParseLegacyV1RequirementByCommit(t *testing.T) {
	commit := strings.Repeat("a", 40)
	got, err := parseLegacyV1Requirement("github.com/acme/common@" + commit)
	require.NoError(t, err)
	require.Equal(t, v1.Requirement{Module: "github.com/acme/common", Version: commit}, got)
}

func TestReadGitDependencyBufRegistryMetadataKeepsGitIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
		roots []string
	}{
		{
			name: "buf v2",
			files: map[string]string{
				"buf.yaml": "version: v2\nmodules:\n  - path: proto\ndeps:\n  - buf.build/googleapis/googleapis\n",
			},
			roots: []string{"proto"},
		},
		{
			name: "buf v1",
			files: map[string]string{
				"buf.yaml": "version: v1\ndeps:\n  - buf.build/acme/payments:deadbeef\n",
			},
			roots: []string{"."},
		},
		{
			name: "buf v1beta1",
			files: map[string]string{
				"buf.yaml": "version: v1beta1\nbuild:\n  roots: [proto]\ndeps:\n  - buf.build/acme/common:v1.2.3\n",
			},
			roots: []string{"proto"},
		},
		{
			name: "buf v1 workspace with nested registry metadata",
			files: map[string]string{
				"buf.work.yaml":    "version: v1\ndirectories: [proto/a, proto/b]\n",
				"proto/a/buf.yaml": "version: v1\ndeps: [buf.build/acme/a-dep]\n",
				"proto/b/buf.yaml": "version: v1\ndeps: [buf.build/acme/b-dep:v1.0.0]\n",
			},
			roots: []string{"proto/a", "proto/b"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			for name, body := range tt.files {
				path := filepath.Join(dir, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
			}

			module, err := ReadGitDependency(dir, "example.com/dependency")

			require.NoError(t, err)
			require.Equal(t, "example.com/dependency", module.Name)
			require.Equal(t, tt.roots, module.Roots)
			require.Empty(t, module.Requires, "BSR identities must not become Git requirements")
		})
	}
}

func TestReadGitDependencyBufWorkspaceIgnoresUnselectedRootBufMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := map[string]string{
		"buf.work.yaml":    "version: v1\ndirectories: [proto/a]\n",
		"buf.yaml":         "version: v1\ndeps: [buf.build/acme/not-in-workspace]\n",
		"proto/a/buf.yaml": "version: v1\n",
		"proto/a/a.proto":  "syntax = \"proto3\";\n",
	}
	for name, body := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	}

	module, err := ReadGitDependency(dir, "example.com/dependency")

	require.NoError(t, err)
	require.Equal(t, []string{"proto/a"}, module.Roots)
}

func TestReadGitDependencyBufWorkspaceRejectsEscapingDirectoryBeforeMetadataRead(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside")
	require.NoError(t, os.MkdirAll(outside, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(outside, bufModuleConfigFile), []byte(
		"version: v1\ndeps: [buf.build/acme/outside]\n",
	), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, bufWorkConfigFile), []byte(
		"version: v1\ndirectories: [../outside]\n",
	), 0o644))

	_, err := ReadGitDependency(root, "example.com/dependency")

	require.ErrorContains(t, err, "leaves the repository")
	require.NotContains(t, err.Error(), "buf.build/acme/outside")
}
