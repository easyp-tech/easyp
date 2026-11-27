package core

import (
	"context"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/compilerbackend"
)

func TestMarshalDescriptorSetMatchesMarshal(t *testing.T) {
	t.Parallel()

	sources := map[string]string{
		"options.proto": `syntax = "proto2"; package custom; import "google/protobuf/descriptor.proto"; extend google.protobuf.MessageOptions { optional string note = 50001; }`,
		"shared.proto":  `syntax = "proto3"; package shared; message Value { string text = 1; map<string, int32> counts = 2; }`,
		"first.proto": `syntax = "proto3"; package app;
import "options.proto";
import "shared.proto";
// First documents a message with custom options.
message First {
  option (custom.note) = "kept";
  shared.Value value = 1;
}`,
		"second.proto": `syntax = "proto3"; package app; import "shared.proto"; message Second { repeated shared.Value items = 1; }`,
	}
	compiled, err := compilerbackend.CompileDescriptorSet(t.Context(), compilerbackend.Request{
		Files:           []string{"first.proto", "second.proto"},
		StandardImports: true,
		SourceInfo:      true,
		Accessor: func(name string) (io.ReadCloser, error) {
			content, ok := sources[name]
			if !ok {
				return nil, fs.ErrNotExist
			}
			return io.NopCloser(strings.NewReader(content)), nil
		},
	})
	require.NoError(t, err)
	all := compiled.Files
	require.Greater(t, len(all), 2)

	for _, files := range [][]*descriptorpb.FileDescriptorProto{all, all[:1], {}, nil} {
		want, err := proto.MarshalOptions{Deterministic: true}.Marshal(&descriptorpb.FileDescriptorSet{File: files})
		require.NoError(t, err)
		got, err := MarshalDescriptorSet(t.Context(), files)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}

	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = MarshalDescriptorSet(canceled, all)
	require.ErrorIs(t, err, context.Canceled)
}
