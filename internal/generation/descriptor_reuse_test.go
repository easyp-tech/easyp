package generation

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestCombinePreparedDescriptorTargetReusesValidatedSnapshot(t *testing.T) {
	t.Parallel()
	target := prepareDescriptorReuseTarget(t, "item.proto", descriptorReuseProto)
	before := proto.Clone(target.full).(*descriptorpb.FileDescriptorSet)

	full, err := combineDescriptorTargets([]preparedDescriptorTarget{target}, true)
	require.NoError(t, err)
	assert.Same(t, target.full, full)

	selected, err := combineDescriptorTargets([]preparedDescriptorTarget{target}, false)
	require.NoError(t, err)
	require.Equal(t, []string{"item.proto"}, descriptorNames(selected))
	for _, file := range target.full.File {
		if file.GetName() == "item.proto" {
			assert.Same(t, file, selected.File[0])
		}
	}
	assert.True(t, proto.Equal(before, target.full))
}

func TestCombineDescriptorTargetValidatesUnpreparedGraph(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		includeImports bool
	}{
		{name: "targets only"},
		{name: "complete graph", includeImports: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target := preparedDescriptorTarget{
				plan: &core.GenerationPlan{},
				full: &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
					Name:       proto.String("item.proto"),
					Syntax:     proto.String("proto3"),
					Dependency: []string{"missing.proto"},
				}}},
				label: "unprepared target",
			}

			_, err := combineDescriptorTargets([]preparedDescriptorTarget{target}, tt.includeImports)

			require.ErrorContains(t, err, "validate combined descriptor set for [unprepared target]")
			assert.ErrorContains(t, err, "missing.proto")
		})
	}
}

func TestCombinePreparedDescriptorTargetsValidatesCombinedGraph(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		includeImports bool
	}{
		{name: "targets only"},
		{name: "complete graph", includeImports: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first := prepareDescriptorReuseTarget(t, "first.proto", "syntax = \"proto3\"; package duplicate; message Item {}")
			second := prepareDescriptorReuseTarget(t, "second.proto", "syntax = \"proto3\"; package duplicate; message Item {}")
			first.label, second.label = "first validated target", "second validated target"

			_, err := combineDescriptorTargets([]preparedDescriptorTarget{first, second}, tt.includeImports)

			require.ErrorContains(t, err, "validate combined descriptor set for [first validated target; second validated target]")
			assert.ErrorContains(t, err, "name conflict")
		})
	}
}

func TestDescriptorDirectorySnapshotPreservesMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		includeImports bool
	}{
		{name: "targets only"},
		{name: "complete graph", includeImports: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target := prepareDescriptorReuseTarget(t, "item.proto", descriptorReuseProto)
			before := proto.Clone(target.full).(*descriptorpb.FileDescriptorSet)
			// Callers may mutate public descriptor copies without changing exports.
			copy := target.plan.DescriptorSet(true)
			for _, file := range copy.File {
				file.Name = proto.String("changed.proto")
				file.Options = nil
				file.SourceCodeInfo = nil
			}
			request := Request{
				WorkDir:             target.selected.directory,
				DescriptorSetOutDir: filepath.Join(t.TempDir(), "sets"),
				IncludeImports:      tt.includeImports,
			}

			outputs, err := planDescriptorOutputs(t.Context(), request, []preparedDescriptorTarget{target})

			require.NoError(t, err)
			require.Len(t, outputs, 1)
			want := &descriptorpb.FileDescriptorSet{}
			for _, file := range before.File {
				if tt.includeImports || file.GetName() == "item.proto" {
					want.File = append(want.File, file)
				}
			}
			expected, err := proto.MarshalOptions{Deterministic: true}.Marshal(want)
			require.NoError(t, err)
			assert.Equal(t, expected, outputs[0].raw)
			assert.True(t, proto.Equal(before, target.full))
			assert.True(t, bytes.Contains(outputs[0].raw, []byte("prepared owner")))
			assert.True(t, bytes.Contains(outputs[0].raw, []byte("Prepared source comment.")))
			assert.True(t, bytes.Contains(outputs[0].raw, []byte("example.com/original;reuse")))
		})
	}
}

func prepareDescriptorReuseTarget(t *testing.T, name, source string) preparedDescriptorTarget {
	t.Helper()
	root := t.TempDir()
	writeV1GenerateFixture(t, root, "protobuf.mod", "module example.com/reuse\n")
	writeV1GenerateFixture(t, root, name, source)
	configPath := filepath.Join(root, "easyp.gen.yaml")
	request := Request{WorkDir: root, WorkspaceRoot: root, DescriptorSetOut: filepath.Join(root, "unused.pb")}
	targets, err := prepareDescriptorTargets(t.Context(), logger.NewNop(), nil, request, []generationTarget{{
		configPath: configPath,
		module:     v1ModuleSelection{directory: root},
	}})
	require.NoError(t, err)
	require.Len(t, targets, 1)
	return targets[0]
}

const descriptorReuseProto = `syntax = "proto2";
package reuse;
import "google/protobuf/descriptor.proto";
option go_package = "example.com/original;reuse";
extend google.protobuf.FileOptions {
  optional string owner = 50001;
}
option (owner) = "prepared owner";
// Prepared source comment.
message Item {}
`
