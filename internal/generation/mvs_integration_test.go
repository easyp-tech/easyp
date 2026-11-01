package generation

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitmodules"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/modules"
)

func TestTidyAndGenerationSelectSharedSemanticMinimum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		direct     string
		transitive string
	}{
		{name: "transitive raises direct minimum", direct: "v1.2.0", transitive: "v1.5.0"},
		{name: "direct raises transitive minimum", direct: "v1.5.0", transitive: "v1.2.0"},
	}
	for _, tt := range tests {
		for _, previousLock := range []bool{false, true} {
			for _, reverse := range []bool{false, true} {
				name := tt.name + "/fresh lock/common first"
				if previousLock {
					name = tt.name + "/existing lower lock/common first"
				}
				if reverse {
					name = strings.TrimSuffix(name, "common first") + "producer first"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					common := t.TempDir()
					writeV1GenerateFixture(t, common, v1.ModuleFile, "module "+common+"\nroots proto\n")
					runTestGit(t, common, "init", "-q")
					const original = "syntax = \"proto3\"; package common.v1; message Common { string value = 1; }\n"
					const selected = original + "message NewInMinor { string value = 1; }\n"
					commits := make(map[string]string)
					for _, revision := range []struct{ version, source string }{
						{version: "v1.2.0", source: original},
						{version: "v1.5.0", source: selected},
						{version: "v1.9.0", source: selected + "message NotRequired {}\n"},
					} {
						writeV1GenerateFixture(t, common, "proto/common/v1/types.proto", revision.source)
						runTestGit(t, common, "add", ".")
						runTestGit(t, common, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", revision.version)
						runTestGit(t, common, "tag", revision.version)
						commits[revision.version] = strings.TrimSpace(runTestGit(t, common, "rev-parse", "HEAD"))
					}

					producer := t.TempDir()
					writeV1GenerateFixture(t, producer, v1.ModuleFile, "module "+producer+"\nroots proto\nrequire "+common+" "+tt.transitive+"\n")
					writeV1GenerateFixture(t, producer, "proto/service/v1/service.proto", "syntax = \"proto3\"; package service.v1; import \"common/v1/types.proto\"; message Service { common.v1.Common value = 1; }\n")
					runTestGit(t, producer, "init", "-q")
					runTestGit(t, producer, "add", ".")
					runTestGit(t, producer, "-c", "user.name=EasyP Test", "-c", "user.email=test@example.com", "commit", "-qm", "producer")
					runTestGit(t, producer, "tag", "v1.0.0")

					root := t.TempDir()
					cache := gitmodules.New(t.TempDir())
					const header = "module example.test/consumer\nroots proto\n"
					if previousLock {
						writeV1GenerateFixture(t, root, v1.ModuleFile, header+"require "+common+" v1.2.0\n")
						writeV1GenerateFixture(t, root, "proto/request.proto", "syntax = \"proto3\"; import \"common/v1/types.proto\"; message Request { common.v1.Common value = 1; }\n")
						require.NoError(t, modules.Tidy(t.Context(), root, cache))
					}
					requirements := []string{"require " + common + " " + tt.direct, "require " + producer + " v1.0.0"}
					if reverse {
						slices.Reverse(requirements)
					}
					writeV1GenerateFixture(t, root, v1.ModuleFile, header+strings.Join(requirements, "\n")+"\n")
					writeV1GenerateFixture(t, root, "proto/request.proto", "syntax = \"proto3\"; package consumer; import \"common/v1/types.proto\"; import \"service/v1/service.proto\"; message Request { common.v1.NewInMinor value = 1; service.v1.Service service = 2; }\n")
					writeV1GenerateFixture(t, root, v1.GenerateFile, "version: v1\n")

					require.NoError(t, modules.Tidy(t.Context(), root, cache))

					lock, err := modules.ReadLock(filepath.Join(root, v1.LockFile))
					require.NoError(t, err)
					require.Len(t, lock.Modules, 2, "direct and transitive requirements share one selected module")
					versions := make(map[string]string, len(lock.Modules))
					for _, entry := range lock.Modules {
						versions[entry.Source] = entry.Version
						if entry.Source == common {
							assert.Equal(t, commits["v1.5.0"], entry.Commit)
						}
					}
					assert.Equal(t, map[string]string{common: "v1.5.0", producer: "v1.0.0"}, versions)
					manifestBefore, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
					require.NoError(t, err)
					lockBefore, err := os.ReadFile(filepath.Join(root, v1.LockFile))
					require.NoError(t, err)
					require.NoError(t, modules.Tidy(t.Context(), root, cache))

					for _, cold := range []bool{false, true} {
						generationCache := cache
						output := filepath.Join(root, "warm.pb")
						if cold {
							generationCache = gitmodules.New(t.TempDir())
							output = filepath.Join(root, "cold.pb")
						}
						require.NoError(t, Run(t.Context(), logger.NewNop(), generationCache, Request{WorkDir: root, WorkspaceRoot: root, Frozen: true, DescriptorSetOut: output, IncludeImports: true}))
						descriptors := readDescriptorSet(t, output)
						assert.ElementsMatch(t, []string{"request.proto", "service/v1/service.proto", "common/v1/types.proto"}, descriptorNames(descriptors))
						for _, file := range descriptors.File {
							if file.GetName() == "common/v1/types.proto" {
								var names []string
								for _, message := range file.MessageType {
									names = append(names, message.GetName())
								}
								assert.Equal(t, []string{"Common", "NewInMinor"}, names, "only the selected version supplies proto contents")
							}
						}
					}
					manifestAfter, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
					require.NoError(t, err)
					lockAfter, err := os.ReadFile(filepath.Join(root, v1.LockFile))
					require.NoError(t, err)
					assert.Equal(t, manifestBefore, manifestAfter)
					assert.Equal(t, lockBefore, lockAfter)
				})
			}
		}
	}
}
