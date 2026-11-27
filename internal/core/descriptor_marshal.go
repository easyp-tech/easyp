package core

import (
	"context"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/parallel"
)

// MarshalDescriptorSet deterministically marshals a FileDescriptorSet of files,
// encoding the files in parallel. The bytes equal a deterministic Marshal of
// &descriptorpb.FileDescriptorSet{File: files}.
func MarshalDescriptorSet(ctx context.Context, files []*descriptorpb.FileDescriptorProto) ([]byte, error) {
	encoded := make([][]byte, len(files))
	err := parallel.Run(ctx, len(files), 0, func(index int) error {
		data, err := proto.MarshalOptions{Deterministic: true}.Marshal(files[index])
		encoded[index] = data
		return err
	})
	if err != nil {
		return nil, err
	}
	size := 0
	for _, data := range encoded {
		size += protowire.SizeTag(1) + protowire.SizeBytes(len(data))
	}
	out := make([]byte, 0, size)
	for _, data := range encoded {
		out = protowire.AppendTag(out, 1, protowire.BytesType)
		out = protowire.AppendBytes(out, data)
	}
	return out, nil
}
