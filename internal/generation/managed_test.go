package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestGenerateV1ManagedRulesMatchModuleIdentities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		remote bool
	}{
		{name: "local_replacement"},
		{name: "locked_dependency", remote: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dependency := t.TempDir()
			cacheBase := t.TempDir()
			cache := gitmodules.New(cacheBase)
			source := "example.com/dep"
			if tt.remote {
				source = dependency
			}
			writeV1GenerateFixture(t, dependency, "protobuf.mod", "module "+source+"\nroots proto\n")
			writeV1GenerateFixture(t, dependency, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; option go_package = \"example.com/original/dep/v1;depv1\"; message Dep {}\n")
			manifest := fmt.Sprintf("module example.com/root\nroots proto\nrequire %s v1.0.0\n", source)
			if tt.remote {
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
				replacement, err := filepath.Rel(root, dependency)
				require.NoError(t, err)
				manifest += fmt.Sprintf("replace %s => %s\n", source, replacement)
			}
			writeV1GenerateFixture(t, root, "protobuf.mod", manifest)
			writeV1GenerateFixture(t, root, "proto/root/v1/root.proto", "syntax = \"proto3\"; package root.v1; import \"dep/v1/dep.proto\"; message Root { dep.v1.Dep value = 1; }\n")
			gen, err := v1.ParseGenerate(strings.NewReader(fmt.Sprintf(`version: v1
generate:
  managed:
    enabled: true
    disable:
      - module: %s
    override:
      - file_option: go_package_prefix
        value: example.com/global
      - module: example.com/root
        file_option: go_package_prefix
        value: example.com/current
`, source)))
			require.NoError(t, err)
			descriptors := generateV1Descriptors(t, root, func(request Request) error {
				return generateSelectedV1Module(t.Context(), logger.NewNop(), cache, request, filepath.Join(root, "easyp.gen.yaml"), root, v1ModuleSelection{directory: root}, gen)
			})
			require.Contains(t, descriptors, "dep/v1/dep.proto")
			require.Contains(t, descriptors, "root/v1/root.proto")
			assert.Equal(t, "example.com/original/dep/v1;depv1", descriptors["dep/v1/dep.proto"].GetOptions().GetGoPackage())
			assert.Equal(t, "example.com/current/root/v1;rootv1", descriptors["root/v1/root.proto"].GetOptions().GetGoPackage())
		})
	}
}

func writeV1GenerateFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte(content), 0o644)
	require.NoError(t, err)
}

func generateV1Descriptors(t *testing.T, root string, generate func(Request) error) map[string]*descriptorpb.FileDescriptorProto {
	t.Helper()
	output := filepath.Join(root, "descriptors.pb")
	require.NoError(t, generate(Request{DescriptorSetOut: output, IncludeImports: true}))
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	var set descriptorpb.FileDescriptorSet
	require.NoError(t, proto.Unmarshal(data, &set))
	files := make(map[string]*descriptorpb.FileDescriptorProto, len(set.File))
	for _, file := range set.File {
		files[file.GetName()] = file
	}
	return files
}
