package compiler

import (
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestCompileDescriptorSetPreservesOrderedClosureAndOwnedMessages(t *testing.T) {
	t.Parallel()
	sources := map[string]string{
		"options.proto": `syntax = "proto2"; package custom; import "google/protobuf/descriptor.proto"; extend google.protobuf.MessageOptions { optional string note = 50001; }`,
		"shared.proto":  `syntax = "proto3"; package shared; message Value { string text = 1; }`,
		"first.proto": `syntax = "proto3"; package app;
import "options.proto"; import "shared.proto";
// First is a documented target.
message First { option (custom.note) = "kept"; shared.Value value = 1; }`,
		"second.proto":         `syntax = "proto3"; package app; message Second {}`,
		"preview/shared.proto": `syntax = "proto3"; package shared; message Value { bool text = 1; }`,
		"preview/broken.proto": `this unimported file must never be parsed`,
	}
	resolve := StandardImports(func(name string) (io.ReadCloser, error) {
		source, ok := sources[name]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return io.NopCloser(strings.NewReader(source)), nil
	})
	options := Options{Resolve: resolve, SourceInfo: true}
	targets := []string{"second.proto", "first.proto", "second.proto"}
	compiled, err := CompileDescriptorSet(t.Context(), targets, options)
	require.NoError(t, err)
	var names []string
	for _, file := range compiled.Files {
		names = append(names, file.GetName())
	}
	assert.Equal(t, []string{"second.proto", "google/protobuf/descriptor.proto", "options.proto", "shared.proto", "first.proto"}, names)
	assert.Equal(t, []string{"google/protobuf/descriptor.proto", "options.proto", "shared.proto"}, compiled.Dependencies)
	set := &descriptorpb.FileDescriptorSet{File: compiled.Files}
	registry, err := protodesc.NewFiles(set)
	require.NoError(t, err)
	first, err := registry.FindFileByPath("first.proto")
	require.NoError(t, err)
	location := first.SourceLocations().ByDescriptor(first.Messages().ByName("First"))
	assert.Contains(t, location.LeadingComments, "First is a documented target.")
	messageOptions := compiled.Files[len(compiled.Files)-1].MessageType[0].GetOptions()
	assert.Empty(t, messageOptions.ProtoReflect().GetUnknown())
	var note string
	messageOptions.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if field.FullName() == "custom.note" {
			note = value.String()
		}
		return true
	})
	assert.Equal(t, "kept", note)
	want := proto.Clone(set).(*descriptorpb.FileDescriptorSet)
	compiled.Files[0].Name = proto.String("mutated.proto")
	compiled.Files[len(compiled.Files)-1].MessageType[0].Name = proto.String("Mutated")
	compiled.Dependencies[0] = "mutated.proto"
	again, err := CompileDescriptorSet(t.Context(), targets, options)
	require.NoError(t, err)
	assert.True(t, proto.Equal(want, &descriptorpb.FileDescriptorSet{File: again.Files}))
	assert.Equal(t, []string{"google/protobuf/descriptor.proto", "options.proto", "shared.proto"}, again.Dependencies)
}

func TestCompileDescriptorSetKeepsResolverFailure(t *testing.T) {
	t.Parallel()
	compiled, err := CompileDescriptorSet(t.Context(), []string{"denied.proto"}, Options{
		Resolve: func(string) (io.ReadCloser, error) { return nil, fs.ErrPermission },
	})
	require.ErrorIs(t, err, fs.ErrPermission)
	assert.Empty(t, compiled.Files)
	assert.Empty(t, compiled.Dependencies)
}
