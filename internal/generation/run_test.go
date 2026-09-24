package generation

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestRunUsesExplicitWorkingDirectory(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX shell")
	}
	tests := []struct {
		name            string
		absoluteOutput  bool
		relativeWorkDir bool
		directory       string
	}{
		{name: "relative descriptor output"},
		{name: "absolute descriptor output", absoluteOutput: true},
		{name: "working directory with spaces", directory: "project with spaces"},
		{name: "relative working directory", relativeWorkDir: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(t.TempDir(), tt.directory)
			workDir := root
			if tt.relativeWorkDir {
				cwd, err := os.Getwd()
				require.NoError(t, err)
				workDir, err = filepath.Rel(cwd, root)
				require.NoError(t, err)
			}
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\nroots proto\n")
			writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\nplugins:\n  - name: ./tools/test-plugin\n    out: gen\n")
			script := filepath.Join(root, "tools", "test-plugin")
			require.NoError(t, os.MkdirAll(filepath.Dir(script), 0o755))
			require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\npwd > plugin-workdir.txt\n"), 0o755))
			output := "descriptor.pb"
			if tt.absoluteOutput {
				output = filepath.Join(root, output)
			}

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: workDir, DescriptorSetOut: output})

			require.NoError(t, err)
			if !filepath.IsAbs(output) {
				output = filepath.Join(root, output)
			}
			raw, err := os.ReadFile(output)
			require.NoError(t, err)
			var descriptors descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(raw, &descriptors))
			require.Len(t, descriptors.File, 1)
			assert.Equal(t, "item.proto", descriptors.File[0].GetName())
			workingDirectory, err := os.ReadFile(filepath.Join(root, "plugin-workdir.txt"))
			require.NoError(t, err)
			resolvedRoot, err := filepath.EvalSymlinks(root)
			require.NoError(t, err)
			resolvedWorkDir, err := filepath.EvalSymlinks(strings.TrimSpace(string(workingDirectory)))
			require.NoError(t, err)
			assert.Equal(t, resolvedRoot, resolvedWorkDir)
		})
	}
}

func TestRunUsesExplicitPluginSources(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX shell")
	}

	tests := []struct {
		name         string
		pluginSource string
	}{
		{name: "binary path", pluginSource: "path: ./tools/test-plugin"},
		{name: "custom command", pluginSource: "command: [sh, ./tools/test-plugin]"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/app\nroots proto\n")
			writeV1GenerateFixture(t, root, "proto/item.proto", "syntax = \"proto3\"; package item.v1; message Item {}")
			writeV1GenerateFixture(t, root, "easyp.gen.yaml", "version: v1\nplugins:\n  - "+tt.pluginSource+"\n    out: gen\n")
			script := filepath.Join(root, "tools", "test-plugin")
			require.NoError(t, os.MkdirAll(filepath.Dir(script), 0o755))
			require.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\ncat >/dev/null\nprintf called > plugin-called.txt\n"), 0o755))

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root})

			require.NoError(t, err)
			called, err := os.ReadFile(filepath.Join(root, "plugin-called.txt"))
			require.NoError(t, err)
			assert.Equal(t, "called", string(called))
		})
	}
}

func TestRunWithoutManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configDir string
	}{
		{name: "root generator"},
		{name: "nested generator", configDir: "app"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, filepath.Join(tt.configDir, "easyp.gen.yaml"), "version: v1\nplugins:\n  - name: python\n    out: gen/python\n")
			writeV1GenerateFixture(t, root, filepath.Join(tt.configDir, "proto/item.proto"), "syntax = \"proto3\"; package item.v1; message Item {}\n")
			output := filepath.Join(root, "descriptor.pb")

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: output})

			require.NoError(t, err)
			raw, err := os.ReadFile(output)
			require.NoError(t, err)
			var descriptors descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(raw, &descriptors))
			require.Len(t, descriptors.File, 1)
			assert.Equal(t, "proto/item.proto", descriptors.File[0].GetName())
		})
	}
}

func TestRunRejectsMultipleDescriptorTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "two modules in one generator",
			files: map[string]string{
				"easyp.gen.yaml":    "version: v1\ngenerate:\n  modules: [m1, m2]\n",
				"m1/protobuf.mod":   "module example.com/m1\nroots proto\n",
				"m1/proto/m1.proto": "syntax = \"proto3\"; package m1.v1; message M1 {}\n",
				"m2/protobuf.mod":   "module example.com/m2\nroots proto\n",
				"m2/proto/m2.proto": "syntax = \"proto3\"; package m2.v1; message M2 {}\n",
			},
		},
		{
			name: "two generator files",
			files: map[string]string{
				"easyp.gen.yaml":          "version: v1\nplugins:\n  - name: python\n    out: gen/python\n",
				"protobuf.mod":            "module example.com/root\nroots proto\n",
				"proto/root.proto":        "syntax = \"proto3\"; package root.v1; message Root {}\n",
				"child/easyp.gen.yaml":    "version: v1\nplugins:\n  - name: python\n    out: gen/python\n",
				"child/protobuf.mod":      "module example.com/child\nroots proto\n",
				"child/proto/child.proto": "syntax = \"proto3\"; package child.v1; message Child {}\n",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for path, contents := range tt.files {
				writeV1GenerateFixture(t, root, path, contents)
			}
			output := filepath.Join(root, "all.pb")
			require.NoError(t, os.WriteFile(output, []byte("existing descriptor"), 0o644))

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: output})

			require.ErrorContains(t, err, "descriptor set requires exactly one selected module")
			contents, readErr := os.ReadFile(output)
			require.NoError(t, readErr)
			assert.Equal(t, "existing descriptor", string(contents))
		})
	}
}

func TestRunWritesDescriptorWithoutPlugins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     map[string]string
		wantProto string
	}{
		{
			name: "sole generator has no plugins",
			files: map[string]string{
				"easyp.gen.yaml":  "version: v1\n",
				"protobuf.mod":    "module example.com/app\nroots proto\n",
				"proto/app.proto": "syntax = \"proto3\"; package app.v1; message App {}\n",
			},
			wantProto: "app.proto",
		},
		{
			name: "parent options file does not become a descriptor target",
			files: map[string]string{
				"easyp.gen.yaml":          "version: v1\noptions:\n  go:\n    package_prefix: example.com/gen\n",
				"child/easyp.gen.yaml":    "version: v1\nplugins:\n  - name: python\n    out: gen/python\n",
				"child/protobuf.mod":      "module example.com/child\nroots proto\n",
				"child/proto/child.proto": "syntax = \"proto3\"; package child.v1; message Child {}\n",
			},
			wantProto: "child.proto",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			for path, contents := range tt.files {
				writeV1GenerateFixture(t, root, path, contents)
			}
			output := filepath.Join(root, "all.pb")

			err := Run(t.Context(), logger.NewNop(), nil, Request{WorkDir: root, DescriptorSetOut: output})

			require.NoError(t, err)
			contents, err := os.ReadFile(output)
			require.NoError(t, err)
			var descriptors descriptorpb.FileDescriptorSet
			require.NoError(t, proto.Unmarshal(contents, &descriptors))
			require.Len(t, descriptors.File, 1)
			assert.Equal(t, tt.wantProto, descriptors.File[0].GetName())
		})
	}
}
