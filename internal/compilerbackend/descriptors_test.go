package compilerbackend

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
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
	req := Request{
		Files:    []string{"second.proto", "first.proto", "second.proto"},
		Accessor: mapAccessor(sources), StandardImports: true, SourceInfo: true,
	}
	compiled, err := CompileDescriptorSet(t.Context(), req)
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
	raw, err := proto.Marshal(set)
	require.NoError(t, err)
	var decoded descriptorpb.FileDescriptorSet
	require.NoError(t, (proto.UnmarshalOptions{Resolver: dynamicpb.NewTypes(registry)}).Unmarshal(raw, &decoded))
	decodedOptions := decoded.File[len(decoded.File)-1].MessageType[0].GetOptions()
	assert.Empty(t, decodedOptions.ProtoReflect().GetUnknown())
	var note string
	decodedOptions.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
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
	again, err := CompileDescriptorSet(t.Context(), req)
	require.NoError(t, err)
	assert.True(t, proto.Equal(want, &descriptorpb.FileDescriptorSet{File: again.Files}))
	assert.Equal(t, []string{"google/protobuf/descriptor.proto", "options.proto", "shared.proto"}, again.Dependencies)

	req.ResolveExtensions = true
	resolved, err := CompileDescriptorSet(t.Context(), req)
	require.NoError(t, err)
	messageOptions := resolved.Files[len(resolved.Files)-1].MessageType[0].GetOptions()
	assert.Empty(t, messageOptions.ProtoReflect().GetUnknown())
	var resolvedNote string
	messageOptions.ProtoReflect().Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if field.FullName() == "custom.note" {
			resolvedNote = value.String()
		}
		return true
	})
	assert.Equal(t, "kept", resolvedNote)
}

func TestCompileDescriptorSetKeepsResolverFailure(t *testing.T) {
	t.Parallel()
	compiled, err := CompileDescriptorSet(t.Context(), Request{
		Files:    []string{"denied.proto"},
		Accessor: func(string) (io.ReadCloser, error) { return nil, fs.ErrPermission },
	})
	require.ErrorIs(t, err, fs.ErrPermission)
	assert.Empty(t, compiled.Files)
	assert.Empty(t, compiled.Dependencies)
}
