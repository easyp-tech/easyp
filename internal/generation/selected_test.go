package generation

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestGenerateSelectedV1ModuleReadsDependencyMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		files  map[string]string
		locked bool
	}{
		{
			name:  "buf_v2_replacement",
			files: map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n"},
		},
		{
			name:   "buf_v2_locked",
			files:  map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n"},
			locked: true,
		},
		{
			name:  "buf_workspace_replacement",
			files: map[string]string{"buf.work.yaml": "version: v1\ndirectories: [proto]\n"},
		},
		{
			name:   "buf_workspace_locked",
			files:  map[string]string{"buf.work.yaml": "version: v1\ndirectories: [proto]\n"},
			locked: true,
		},
		{
			name: "legacy_easyp_replacement",
			files: map[string]string{
				"easyp.yaml":   "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n",
				"protobuf.mod": "direct (\n)\n",
			},
		},
		{
			name: "legacy_easyp_locked",
			files: map[string]string{
				"easyp.yaml":   "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n",
				"protobuf.mod": "direct (\n)\n",
			},
			locked: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dependency := filepath.Join(root, "dependency")
			cacheBase := t.TempDir()
			cache := gitmodules.New(cacheBase)
			source := "example.com/dep"
			if tt.locked {
				source = dependency
			}
			for path, content := range tt.files {
				writeV1GenerateFixture(t, dependency, path, content)
			}
			writeV1GenerateFixture(t, dependency, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
			manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", source)
			if tt.locked {
				runTestGit(t, dependency, "init", "-q")
				runTestGit(t, dependency, "add", ".")
				runTestGit(t, dependency, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
				runTestGit(t, dependency, "tag", "v1.0.0")
				module, err := v1.ParseModule(strings.NewReader(manifest))
				require.NoError(t, err)
				lock, err := modules.Resolve(t.Context(), module, gitmodules.New(cacheBase), nil)
				require.NoError(t, err)
				require.NoError(t, writeTestLock(root, lock))
			} else {
				manifest += "replace example.com/dep => ./dependency\n"
			}
			writeV1GenerateFixture(t, root, "protobuf.mod", manifest)
			gen, err := v1.ParseGenerate(strings.NewReader(fmt.Sprintf(`version: v1
generate:
  modules: [%s]
  managed:
    enabled: true
    override:
      - module: %s
        file_option: go_package_prefix
        value: example.com/selected
`, source, source)))
			require.NoError(t, err)

			descriptors := generateV1Descriptors(t, root, func(request Request) error {
				return generateSelectedV1Module(t.Context(), logger.NewNop(), cache, request, filepath.Join(root, "easyp.gen.yaml"), root, v1ModuleSelection{source: source}, gen)
			})

			require.Contains(t, descriptors, "dep/v1/dep.proto")
			assert.Len(t, descriptors, 1)
			assert.Equal(t, "example.com/selected/dep/v1;depv1", descriptors["dep/v1/dep.proto"].GetOptions().GetGoPackage())
		})
	}
}

func TestSelectV1ModulesPrefersGeneratorSiblingThenRepositoryRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		manifests     map[string]string
		wantDirectory string
		wantError     string
	}{
		{
			name: "sibling_manifest",
			manifests: map[string]string{
				"protobuf.mod":              "module example.com/root\n",
				"projects/protobuf.mod":     "module example.com/intermediate\n",
				"projects/app/protobuf.mod": "module example.com/app\n",
			},
			wantDirectory: "projects/app",
		},
		{
			name: "root_manifest_skips_intermediate",
			manifests: map[string]string{
				"protobuf.mod":          "module example.com/root\n",
				"projects/protobuf.mod": "module example.com/intermediate\n",
			},
			wantDirectory: ".",
		},
		{
			name:      "missing_manifest",
			manifests: map[string]string{"projects/protobuf.mod": "module example.com/intermediate\n"},
			wantError: "no protobuf.mod for generator",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			configDir := filepath.Join(root, "projects", "app")
			for path, content := range tt.manifests {
				writeV1GenerateFixture(t, root, path, content)
			}

			modules, err := selectV1Modules(root, configDir, nil)

			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []v1ModuleSelection{{directory: filepath.Join(root, tt.wantDirectory)}}, modules)
		})
	}
}

func TestGenerateSelectedV1ModuleUsesGeneratorSiblingRequirements(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		rootManifest    string
		siblingManifest string
	}{
		{
			name:            "sibling_replacement_over_root_replacement",
			rootManifest:    "module example.com/root\nrequire example.com/dep v1.0.0\nreplace example.com/dep => ./missing\n",
			siblingManifest: "module example.com/app\nrequire example.com/dep v1.0.0\nreplace example.com/dep => ./dependency\n",
		},
		{
			name:            "sibling_dependency_absent_from_root",
			rootManifest:    "module example.com/root\n",
			siblingManifest: "module example.com/app\nrequire example.com/dep v1.0.0\nreplace example.com/dep => ./dependency\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			cache := gitmodules.New(t.TempDir())
			configDir := filepath.Join(root, "app")
			dependency := filepath.Join(configDir, "dependency")
			writeV1GenerateFixture(t, root, "protobuf.mod", tt.rootManifest)
			writeV1GenerateFixture(t, configDir, "protobuf.mod", tt.siblingManifest)
			writeV1GenerateFixture(t, dependency, "buf.yaml", "version: v2\nmodules:\n  - path: proto\n")
			writeV1GenerateFixture(t, dependency, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
			gen, err := v1.ParseGenerate(strings.NewReader("version: v1\ngenerate:\n  modules: [example.com/dep]\n"))
			require.NoError(t, err)

			descriptors := generateV1Descriptors(t, root, func(request Request) error {
				return generateSelectedV1Module(t.Context(), logger.NewNop(), cache, request, filepath.Join(configDir, "easyp.gen.yaml"), root, v1ModuleSelection{source: "example.com/dep"}, gen)
			})

			assert.Contains(t, descriptors, "dep/v1/dep.proto")
		})
	}
}
