package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyAndGenerationUseProducerBufExcludes(t *testing.T) {
	t.Parallel()
	for _, metadata := range []struct {
		name   string
		prefix string
		files  map[string]string
	}{
		{name: "v1", files: map[string]string{"buf.yaml": "version: v1\nbuild:\n  excludes: [preview]\n"}},
		{name: "v2", prefix: "proto/", files: map[string]string{"buf.yaml": "version: v2\nmodules:\n  - path: proto\n    excludes: [proto/preview]\n"}},
		{name: "v1_workspace", prefix: "proto/", files: map[string]string{
			"buf.work.yaml":  "version: v1\ndirectories: [proto]\n",
			"proto/buf.yaml": "version: v1\nbuild:\n  excludes: [preview]\n",
		}},
	} {
		for _, mode := range []string{"pinned git", "local replace"} {
			t.Run(metadata.name+"/"+mode, func(t *testing.T) {
				t.Parallel()
				fixture := newBufExcludesGenerationFixture(t, mode, metadata.prefix, metadata.files)
				producerFiles := []string{
					metadata.prefix + "types.proto", metadata.prefix + "preview/types.proto",
					metadata.prefix + "preview/broken.proto", metadata.prefix + "previewed/retained.proto",
				}
				for name := range metadata.files {
					producerFiles = append(producerFiles, name)
				}
				producerBefore := bufExcludesGenerationBytes(t, fixture.producer, producerFiles...)

				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))

				manifestBefore := bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile)
				output := filepath.Join(fixture.consumer, "all.pb")
				request := Request{
					WorkDir: fixture.consumer, WorkspaceRoot: fixture.consumer, Frozen: mode == "pinned git",
					DescriptorSetOut: output, IncludeImports: true,
				}
				require.NoError(t, Run(t.Context(), logger.NewNop(), fixture.cache, request))
				descriptors := readDescriptorSet(t, output)
				assert.ElementsMatch(t, []string{"request.proto", "types.proto", "previewed/retained.proto"}, descriptorNames(descriptors))
				_, err := protodesc.NewFiles(descriptors)
				require.NoError(t, err)
				for _, file := range descriptors.File {
					if file.GetName() == "types.proto" {
						require.Len(t, file.MessageType, 1)
						require.Len(t, file.MessageType[0].Field, 1)
						assert.Equal(t, descriptorpb.FieldDescriptorProto_TYPE_STRING, file.MessageType[0].Field[0].GetType())
					}
				}
				require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
				assert.Equal(t, manifestBefore, bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile))

				writeV1GenerateFixture(t, fixture.consumer, "api/request.proto", "syntax = \"proto3\"; package app; import \"preview/types.proto\"; message Request { types.Value value = 1; }\n")
				failureBefore := bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile, "api/request.proto", "all.pb")
				err = modules.Tidy(t.Context(), fixture.consumer, fixture.cache)
				require.ErrorContains(t, err, "cannot resolve imports")
				assert.ErrorContains(t, err, "preview/types.proto")
				assert.Equal(t, failureBefore, bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile, "api/request.proto", "all.pb"))

				err = Run(t.Context(), logger.NewNop(), fixture.cache, request)
				require.ErrorContains(t, err, "preview/types.proto")
				assert.Equal(t, failureBefore, bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile, "api/request.proto", "all.pb"))
				request.DescriptorSetOut = ""
				request.DescriptorSetOutDir = filepath.Join(fixture.consumer, "failed-descriptors")
				require.Error(t, Run(t.Context(), logger.NewNop(), fixture.cache, request))
				assert.NoDirExists(t, request.DescriptorSetOutDir)
				assert.Equal(t, failureBefore, bufExcludesGenerationBytes(t, fixture.consumer, v1.ModuleFile, v1.LockFile, "api/request.proto", "all.pb"))
				assert.Equal(t, producerBefore, bufExcludesGenerationBytes(t, fixture.producer, producerFiles...))
			})
		}
	}
}

func TestTidyAndGenerationDoNotInventBufExcludes(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"pinned git", "local replace"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			fixture := newBufExcludesGenerationFixture(t, mode, "", nil)
			writeV1GenerateFixture(t, fixture.consumer, "api/request.proto", "syntax = \"proto3\"; package app; import \"preview/types.proto\"; message Request { preview.Value value = 1; }\n")

			require.NoError(t, modules.Tidy(t.Context(), fixture.consumer, fixture.cache))
			output := filepath.Join(fixture.consumer, "all.pb")
			require.NoError(t, Run(t.Context(), logger.NewNop(), fixture.cache, Request{
				WorkDir: fixture.consumer, WorkspaceRoot: fixture.consumer, Frozen: mode == "pinned git",
				DescriptorSetOut: output, IncludeImports: true,
			}))

			descriptors := readDescriptorSet(t, output)
			assert.ElementsMatch(t, []string{"request.proto", "types.proto", "preview/types.proto", "previewed/retained.proto"}, descriptorNames(descriptors))
			_, err := protodesc.NewFiles(descriptors)
			require.NoError(t, err)
			assert.NoFileExists(t, filepath.Join(fixture.producer, "buf.yaml"))
			assert.NoFileExists(t, filepath.Join(fixture.producer, v1.ModuleFile))
		})
	}
}

type bufExcludesGenerationFixture struct {
	consumer string
	producer string
	cache    *gitmodules.Cache
}

func newBufExcludesGenerationFixture(t *testing.T, mode, prefix string, metadata map[string]string) bufExcludesGenerationFixture {
	t.Helper()
	producer, consumer := t.TempDir(), t.TempDir()
	for name, content := range metadata {
		writeV1GenerateFixture(t, producer, name, content)
	}
	writeV1GenerateFixture(t, producer, prefix+"types.proto", "syntax = \"proto3\"; package types; message Value { string value = 1; }\n")
	writeV1GenerateFixture(t, producer, prefix+"previewed/retained.proto", "syntax = \"proto3\"; package retained; message Retained {}\n")
	preview := "syntax = \"proto3\"; package preview; message Value { bool value = 1; }\n"
	if len(metadata) > 0 {
		preview = strings.Replace(preview, "package preview", "package types", 1)
		writeV1GenerateFixture(t, producer, prefix+"preview/broken.proto", "this excluded source must not be parsed")
	}
	writeV1GenerateFixture(t, producer, prefix+"preview/types.proto", preview)
	source := "example.test/types"
	manifest := "module example.test/app\nroots api\nrequire " + source + "\nreplace " + source + " => " + producer + "\n"
	if mode == "pinned git" {
		runTestGit(t, producer, "init", "-q")
		runTestGit(t, producer, "add", ".")
		runTestGit(t, producer, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "producer Buf selection")
		commit := strings.TrimSpace(runTestGit(t, producer, "rev-parse", "HEAD"))
		source = producer
		manifest = "module example.test/app\nroots api\nrequire " + source + " " + commit + "\n"
	} else {
		writeV1GenerateFixture(t, consumer, v1.LockFile, "version: 1\nmodules: []\n# published lock remains unchanged for local overlays\n")
	}
	writeV1GenerateFixture(t, consumer, v1.ModuleFile, manifest)
	writeV1GenerateFixture(t, consumer, "api/request.proto", "syntax = \"proto3\"; package app; import \"types.proto\"; message Request { types.Value value = 1; }\n")
	writeV1GenerateFixture(t, consumer, v1.GenerateFile, fmt.Sprintf("version: v1\ngenerate:\n  modules: [., %q]\n", source))
	return bufExcludesGenerationFixture{consumer: consumer, producer: producer, cache: gitmodules.New(t.TempDir())}
}

func bufExcludesGenerationBytes(t *testing.T, root string, names ...string) map[string][]byte {
	t.Helper()
	contents := make(map[string][]byte, len(names))
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(root, name))
		require.NoError(t, err)
		contents[name] = raw
	}
	return contents
}
