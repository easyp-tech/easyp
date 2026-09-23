package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestGenerateSelectedV1ModuleReadsDependencyMetadata(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		body   string
	}{
		{name: "buf v2", config: "buf.yaml", body: "version: v2\nmodules:\n  - path: proto\n"},
		{name: "buf v1 workspace", config: "buf.work.yaml", body: "version: v1\ndirectories: [proto]\n"},
		{name: "legacy easyp", config: "easyp.yaml", body: "generate:\n  inputs:\n    - directory:\n        path: proto\n        root: proto\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []struct {
				name   string
				remote bool
			}{
				{name: "local replacement"},
				{name: "locked dependency", remote: true},
			} {
				t.Run(mode.name, func(t *testing.T) {
					root := t.TempDir()
					dependency := filepath.Join(root, "dependency")
					cacheBase := t.TempDir()
					t.Setenv("EASYPPATH", cacheBase)
					source := "example.com/dep"
					if mode.remote {
						source = dependency
					}
					writeV1GenerateFixture(t, dependency, tc.config, tc.body)
					if tc.name == "legacy easyp" {
						writeV1GenerateFixture(t, dependency, "protobuf.mod", "direct (\n)\n")
					}
					writeV1GenerateFixture(t, dependency, "proto/dep/v1/dep.proto", "syntax = \"proto3\"; package dep.v1; message Dep {}\n")
					manifest := fmt.Sprintf("module example.com/root\nrequire %s v1.0.0\n", source)
					if mode.remote {
						runTestGit(t, dependency, "init", "-q")
						runTestGit(t, dependency, "add", ".")
						runTestGit(t, dependency, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "initial")
						runTestGit(t, dependency, "tag", "v1.0.0")
						module, err := v1.ParseModule(strings.NewReader(manifest))
						require.NoError(t, err)
						lock, err := buildV1Lock(t.Context(), module, filepath.Join(cacheBase, "v1", "git"))
						require.NoError(t, err)
						require.NoError(t, writeV1Lock(root, lock))
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
					descriptors := generateV1Descriptors(t, root, func(ctx *cli.Context) error {
						return generateSelectedV1Module(ctx, logger.NewNop(), filepath.Join(root, "easyp.gen.yaml"), root, source, gen)
					})
					require.Len(t, descriptors, 1)
					require.Equal(t, "example.com/selected/dep/v1;depv1", descriptors["dep/v1/dep.proto"].GetOptions().GetGoPackage())
				})
			}
		})
	}
}
