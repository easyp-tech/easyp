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

func TestReadGitDependencyModuleRejectsBufFilters(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "buf.yaml"), []byte("version: v2\nmodules:\n  - path: proto\n    excludes: [proto/unused]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGitDependency(dir, "example.com/dependency"); err == nil || !strings.Contains(err.Error(), "includes/excludes") {
		t.Fatalf("expected unsupported filter error, got %v", err)
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

	if _, err := parseLegacyV1Requirement("github.com/acme/common@main"); err == nil || !strings.Contains(err.Error(), "non-SemVer") {
		t.Fatalf("expected branch-ref error, got %v", err)
	}
}
