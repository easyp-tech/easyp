package generation

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestGenerationCompilesOnlyImportedDependencyFiles(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"pinned git", "local replace"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			producer, consumer := t.TempDir(), t.TempDir()
			writeV1GenerateFixture(t, producer, "google/type/money.proto", `syntax = "proto3"; package google.type; message Money { int64 units = 1; }`)
			writeV1GenerateFixture(t, producer, "preview/google/type/money.proto", `syntax = "proto3"; package google.type; message Money { string units = 1; }`)
			writeV1GenerateFixture(t, producer, "service.proto", `syntax = "proto3"; package service; import "google/type/money.proto"; message Response { google.type.Money value = 1; }`)
			writeV1GenerateFixture(t, producer, "preview/broken.proto", "this unimported file must never be parsed")
			writeV1GenerateFixture(t, producer, "unused.proto", `syntax = "proto3"; package unused; message Unused {}`)
			manifest := "module example.test/app\nroots api\nrequire example.test/types\nreplace example.test/types => " + producer + "\n"
			if mode == "pinned git" {
				runTestGit(t, producer, "init", "-q")
				runTestGit(t, producer, "add", ".")
				runTestGit(t, producer, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "dependency import closure")
				commit := strings.TrimSpace(runTestGit(t, producer, "rev-parse", "HEAD"))
				manifest = "module example.test/app\nroots api\nrequire " + producer + " " + commit + "\n"
			} else {
				writeV1GenerateFixture(t, consumer, v1.LockFile, "version: 1\nmodules: []\n")
			}
			writeV1GenerateFixture(t, consumer, v1.ModuleFile, manifest)
			// Only the consumer is a compilation target. The dependency has no
			// selection metadata, and both copies must remain importable.
			writeV1GenerateFixture(t, consumer, v1.GenerateFile, "version: v1\n")
			cache := gitmodules.New(t.TempDir())
			output := filepath.Join(consumer, "all.pb")
			request := Request{
				WorkDir: consumer, WorkspaceRoot: consumer, Frozen: mode == "pinned git",
				DescriptorSetOut: output, IncludeImports: true,
			}
			for _, source := range []struct {
				name    string
				imports string
				value   string
				files   []string
				kind    protoreflect.Kind
			}{
				{
					name:    "primary through a transitive import",
					imports: `import "service.proto";`,
					value:   "service.Response",
					files:   []string{"request.proto", "service.proto", "google/type/money.proto"},
					kind:    protoreflect.Int64Kind,
				},
				{
					name:    "preview independently",
					imports: `import "preview/google/type/money.proto";`,
					value:   "google.type.Money",
					files:   []string{"request.proto", "preview/google/type/money.proto"},
					kind:    protoreflect.StringKind,
				},
			} {
				writeV1GenerateFixture(t, consumer, "api/request.proto", `syntax = "proto3"; package app; `+source.imports+` message Request { `+source.value+` value = 1; }`)
				require.NoError(t, modules.Tidy(t.Context(), consumer, cache), source.name)
				before := bufExcludesGenerationBytes(t, consumer, v1.ModuleFile, v1.LockFile)
				for _, cold := range []bool{false, true} {
					generationCache := cache
					if cold {
						generationCache = gitmodules.New(t.TempDir())
					}
					require.NoError(t, Run(t.Context(), logger.NewNop(), generationCache, request), source.name)
					set := readDescriptorSet(t, output)
					assert.ElementsMatch(t, source.files, descriptorNames(set), source.name)
					graph, err := protodesc.NewFiles(set)
					require.NoError(t, err)
					money, err := graph.FindDescriptorByName("google.type.Money")
					require.NoError(t, err)
					assert.Equal(t, source.kind, money.(protoreflect.MessageDescriptor).Fields().ByName("units").Kind())
				}
				request.IncludeImports = false
				require.NoError(t, Run(t.Context(), logger.NewNop(), cache, request), source.name)
				assert.Equal(t, []string{"request.proto"}, descriptorNames(readDescriptorSet(t, output)))
				request.IncludeImports = true
				assert.Equal(t, before, bufExcludesGenerationBytes(t, consumer, v1.ModuleFile, v1.LockFile))
			}

			writeV1GenerateFixture(t, consumer, "api/request.proto", `syntax = "proto3"; package app; import "service.proto"; import "preview/google/type/money.proto"; message Request { service.Response value = 1; }`)
			before := bufExcludesGenerationBytes(t, consumer, v1.ModuleFile, v1.LockFile, "all.pb")
			err := Run(t.Context(), logger.NewNop(), cache, request)
			require.ErrorContains(t, err, `"Money" is already defined`)
			assert.ErrorContains(t, err, "money.proto")
			assert.Equal(t, before, bufExcludesGenerationBytes(t, consumer, v1.ModuleFile, v1.LockFile, "all.pb"))
			for _, name := range []string{"buf.yaml", "buf.work.yaml", v1.ModuleFile} {
				assert.NoFileExists(t, filepath.Join(producer, name))
			}
			for _, name := range []string{"google/type/money.proto", "preview/google/type/money.proto", "preview/broken.proto", "unused.proto"} {
				assert.FileExists(t, filepath.Join(producer, name))
			}
		})
	}
}
