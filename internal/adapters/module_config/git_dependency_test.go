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
				"buf.yaml":     "version: v2\nmodules:\n  - path: old\n",
			},
			roots:    []string{"proto"},
			requires: []v1.Requirement{{Module: "example.com/common", Version: "v1.2.0"}},
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
	require.NoError(t, os.WriteFile(filepath.Join(dir, bufV1ConfigFile), []byte("version: v2\ndirectories: [proto]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, bufV2ConfigFile), []byte("version: v1\n"), 0o644))
	_, err := ReadGitDependency(dir, "example.com/dependency")
	require.ErrorContains(t, err, bufV1ConfigFile)
}

func TestReadOptionalDependencyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(string) error
		wantFound bool
		wantError bool
	}{
		{name: "missing file"},
		{
			name: "present file",
			setup: func(path string) error {
				return os.WriteFile(path, []byte("version: v1\n"), 0o644)
			},
			wantFound: true,
		},
		{
			name: "file cannot be read",
			setup: func(path string) error {
				return os.Mkdir(path, 0o755)
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if tc.setup != nil {
				require.NoError(t, tc.setup(filepath.Join(dir, "config.yaml")))
			}
			data, found, err := readOptionalDependencyConfig(dir, "config.yaml")
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantFound, found)
			if found {
				require.Equal(t, "version: v1\n", string(data))
			}
		})
	}
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
