package moduleconfig

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestReadGitDependencyModuleFormats(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		roots    []string
		requires []v1.Requirement
	}{
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
			if err != nil {
				t.Fatal(err)
			}
			if module.Name != "example.com/dependency" || !reflect.DeepEqual(module.Roots, tc.roots) || !reflect.DeepEqual(module.Requires, tc.requires) {
				t.Fatalf("module = %#v; want roots=%v requires=%v", module, tc.roots, tc.requires)
			}
		})
	}
}

func TestReadGitDependencyModuleRejectsEscapingRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "buf.yaml"), []byte("version: v2\nmodules:\n  - path: ../outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGitDependency(dir, "example.com/dependency"); err == nil || !strings.Contains(err.Error(), "leaves") {
		t.Fatalf("expected root traversal error, got %v", err)
	}
}

func TestReadGitDependencyModuleRejectsBufFilters(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "buf.yaml"), []byte("version: v2\nmodules:\n  - path: proto\n    excludes: [proto/unused]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGitDependency(dir, "example.com/dependency"); err == nil || !strings.Contains(err.Error(), "includes/excludes") {
		t.Fatalf("expected unsupported filter error, got %v", err)
	}
}

func TestParseLegacyV1RequirementWithoutVersion(t *testing.T) {
	got, err := parseLegacyV1Requirement("github.com/acme/common")
	if err != nil {
		t.Fatal(err)
	}
	if got.Module != "github.com/acme/common" || got.Version != "" {
		t.Fatalf("requirement = %#v", got)
	}
}

func TestParseLegacyV1RequirementRejectsBranchRef(t *testing.T) {
	if _, err := parseLegacyV1Requirement("github.com/acme/common@main"); err == nil || !strings.Contains(err.Error(), "non-SemVer") {
		t.Fatalf("expected branch-ref error, got %v", err)
	}
}
