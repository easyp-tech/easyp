package core

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestManagedOptionMetadataMatchesHandlers(t *testing.T) {
	t.Parallel()
	fileOptions := ManagedFileOptionMetadata()
	fieldOptions := ManagedFieldOptionMetadata()
	require.Len(t, fileOptions, len(fileOptionHandlers))
	require.Len(t, fieldOptions, len(fieldOptionHandlers))
	tests := []struct {
		name     string
		metadata ManagedOptionMetadata
		apply    func(any) bool
	}{}
	for i, handler := range fileOptionHandlers {
		tests = append(tests, struct {
			name     string
			metadata ManagedOptionMetadata
			apply    func(any) bool
		}{
			name: string(handler.Option), metadata: fileOptions[i],
			apply: func(value any) bool {
				fd := &descriptorpb.FileDescriptorProto{Name: proto.String("api/v1/service.proto"), Options: &descriptorpb.FileOptions{}}
				before := proto.Clone(fd)
				handler.Apply(fd, value, "acme.v1", fd.GetName())
				return !proto.Equal(before, fd)
			},
		})
	}
	for i, handler := range fieldOptionHandlers {
		tests = append(tests, struct {
			name     string
			metadata ManagedOptionMetadata
			apply    func(any) bool
		}{
			name: string(handler.Option), metadata: fieldOptions[i],
			apply: func(value any) bool {
				field := &descriptorpb.FieldDescriptorProto{Type: descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(), Options: &descriptorpb.FieldOptions{}}
				before := proto.Clone(field)
				handler.Apply(field, value)
				return !proto.Equal(before, field)
			},
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.name, tt.metadata.Name)
			require.Contains(t, []string{"string", "boolean"}, tt.metadata.ValueType)
			values := []any{nil, true, false, 0, 1, "", "value", "true", "SPEED", "CODE_SIZE", "LITE_RUNTIME", "JS_NORMAL", "JS_STRING", "JS_NUMBER", "invalid", []string{"value"}, map[string]string{"value": "value"}}
			for _, value := range tt.metadata.Enum {
				values = append(values, value)
			}
			for _, value := range values {
				accepted := false
				switch value := value.(type) {
				case string:
					accepted = tt.metadata.ValueType == "string" && (len(tt.metadata.Enum) == 0 || slices.Contains(tt.metadata.Enum, value))
				case bool:
					accepted = tt.metadata.ValueType == "boolean"
				}
				assert.Equal(t, accepted, tt.apply(value), fmt.Sprintf("metadata/handler mismatch for %T(%v)", value, value))
			}
		})
	}
}

func TestManagedOptionMetadataDoesNotExposeSharedState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		load func() []ManagedOptionMetadata
	}{
		{name: "file", load: ManagedFileOptionMetadata},
		{name: "field", load: ManagedFieldOptionMetadata},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := tt.load()
			changed := tt.load()
			for i := range changed {
				changed[i].Name = "changed"
				for j := range changed[i].Enum {
					changed[i].Enum[j] = "changed"
				}
			}
			assert.Equal(t, before, tt.load())
		})
	}
}
