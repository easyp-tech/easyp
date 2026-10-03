package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestDescriptorPreflightDoesNotRunPlugins(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		second     string
		secondFile string
		outDir     bool
		wantError  string
	}{
		{name: "different file definitions without imports", second: "syntax = \"proto3\"; package other; message Item {}", secondFile: "item.proto", wantError: "conflicting descriptor"},
		{name: "same symbol in different files", second: "syntax = \"proto3\"; package item; message Item {}", secondFile: "other.proto", wantError: "name conflict"},
		{name: "later compilation failure", second: "not a protobuf file", secondFile: "other.proto", wantError: "Compile"},
		{name: "later compilation failure in directory mode", second: "not a protobuf file", secondFile: "other.proto", outDir: true, wantError: "Compile"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeDescriptorMarker(t, root)
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\ngenerate:\n  modules: [a, b]\n"+descriptorMarkerConfig)
			for _, name := range []string{"a", "b"} {
				writeV1GenerateFixture(t, root, name+"/protobuf.mod", "module example.com/"+name+"\n")
			}
			writeV1GenerateFixture(t, root, "a/item.proto", "syntax = \"proto3\"; package item; message Item {}")
			writeV1GenerateFixture(t, root, "b/"+tt.secondFile, tt.second)
			out := filepath.Join(root, "all.pb")
			require.NoError(t, os.WriteFile(out, []byte("previous output"), 0o600))
			request := Request{WorkDir: root, DescriptorSetOut: out}
			if tt.outDir {
				request.DescriptorSetOut, request.DescriptorSetOutDir = "", filepath.Join(root, "descriptors")
			}

			err := Run(t.Context(), logger.NewNop(), nil, request)

			require.ErrorContains(t, err, tt.wantError)
			assert.NoFileExists(t, filepath.Join(root, "plugin-ran.txt"))
			assert.NoDirExists(t, filepath.Join(root, "descriptors"))
			previous, err := os.ReadFile(out)
			require.NoError(t, err)
			assert.Equal(t, "previous output", string(previous))
		})
	}
}

func TestDescriptorDirectoryIsolatesDependenciesAndConsumers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		include bool
	}{
		{name: "targets only"}, {name: "complete graph", include: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeDescriptorMarker(t, root)
			for _, project := range []string{"backend", "frontend"} {
				writeV1GenerateFixture(t, root, project+"/easyp.gen.yaml", "version: v1\ngenerate:\n  modules: [proto/user, proto/order]\noptions:\n  go:\n    package_prefix: example.com/"+project+"/gen\n"+descriptorMarkerConfig)
			}
			for i, module := range []string{"user", "order"} {
				version := fmt.Sprintf("v1.%d.0", i)
				writeV1GenerateFixture(t, root, "proto/"+module+"/protobuf.mod", "module example.com/"+module+"\nrequire example.com/common "+version+"\nreplace example.com/common => ../../common-"+version+"\n")
				writeV1GenerateFixture(t, root, "proto/"+module+"/"+module+".proto", "syntax = \"proto3\"; package "+module+"; import \"money.proto\"; message Item { common.Money money = 1; }")
				writeV1GenerateFixture(t, root, "common-"+version+"/protobuf.mod", "module example.com/common\n")
				writeV1GenerateFixture(t, root, "common-"+version+"/money.proto", "syntax = \"proto3\"; package common; import \"google/protobuf/timestamp.proto\"; message Money { google.protobuf.Timestamp time = 1; string currency"+fmt.Sprint(i)+" = 2; }")
			}
			outDir := filepath.Join(root, "descriptors")
			request := Request{AllProjects: true, WorkDir: root, DescriptorSetOutDir: "descriptors", IncludeImports: tt.include}

			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))

			first := descriptorTree(t, outDir)
			require.Len(t, first, 4)
			for _, project := range []string{"backend", "frontend"} {
				for i, module := range []string{"user", "order"} {
					matches, err := filepath.Glob(filepath.Join(outDir, project, module+"-*.pb"))
					require.NoError(t, err)
					require.Len(t, matches, 1)
					set := readDescriptorSet(t, matches[0])
					want := []string{module + ".proto"}
					if tt.include {
						want = append(want, "money.proto", "google/protobuf/timestamp.proto")
						_, err := protodesc.NewFiles(set)
						require.NoError(t, err)
					}
					assert.ElementsMatch(t, want, descriptorNames(set))
					for _, file := range set.File {
						if file.GetName() == module+".proto" {
							assert.Contains(t, file.GetOptions().GetGoPackage(), "example.com/"+project+"/gen")
						}
						if file.GetName() == "money.proto" {
							assert.Equal(t, "currency"+fmt.Sprint(i), file.MessageType[0].Field[1].GetName())
						}
					}
				}
			}
			for _, project := range []string{"backend", "frontend"} {
				path := filepath.Join(root, project, "easyp.gen.yaml")
				raw, err := os.ReadFile(path)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(path, []byte(strings.ReplaceAll(string(raw), "[proto/user, proto/order]", "[proto/order, proto/user]")), 0o644))
			}
			require.NoError(t, Run(t.Context(), logger.NewNop(), nil, request))
			assert.Equal(t, first, descriptorTree(t, outDir))
			calls, err := os.ReadFile(filepath.Join(root, "plugin-ran.txt"))
			require.NoError(t, err)
			assert.Equal(t, "xxxxxxxx", string(calls))
		})
	}
}

func TestDescriptorConflictExplainsConsumerOptions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeDescriptorMarker(t, root)
	writeV1GenerateFixture(t, root, "proto/protobuf.mod", "module example.com/item\n")
	writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package item; message Item {}")
	for _, project := range []string{"backend", "frontend"} {
		writeV1GenerateFixture(t, root, project+"/easyp.gen.yaml", "version: v1\ngenerate:\n  modules: [proto]\noptions:\n  go:\n    package_prefix: example.com/"+project+"\n"+descriptorMarkerConfig)
	}
	err := Run(t.Context(), logger.NewNop(), nil, Request{AllProjects: true, WorkDir: root, DescriptorSetOut: "all.pb"})
	for _, want := range []string{"conflicting descriptor", "item.proto", "backend", "frontend", "example.com/item", "options.go_package", "descriptor_set_out_dir"} {
		require.ErrorContains(t, err, want)
	}
	assert.NoFileExists(t, filepath.Join(root, "plugin-ran.txt"))
	assert.NoFileExists(t, filepath.Join(root, "all.pb"))
	require.NoError(t, Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, Project: "backend", DescriptorSetOut: "all.pb"}))
	assert.Contains(t, readDescriptorSet(t, filepath.Join(root, "all.pb")).File[0].GetOptions().GetGoPackage(), "example.com/backend")
}

const descriptorMarkerConfig = "plugins:\n  - command: [sh, ./marker.sh]\n    out: gen\n"

func writeDescriptorMarker(t *testing.T, root string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fixture plugin uses a POSIX shell")
	}
	writeV1GenerateFixture(t, root, "marker.sh", "#!/bin/sh\ncat >/dev/null\nprintf x >> plugin-ran.txt\n")
}

func descriptorTree(t *testing.T, root string) map[string]string {
	t.Helper()
	files := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = string(raw)
		return nil
	})
	require.NoError(t, err)
	return files
}
