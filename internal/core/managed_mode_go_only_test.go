package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGoPackageOnlyPreservesOtherOptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		enabled, disable bool
	}{
		{name: "point option"}, {name: "explicit managed mode", enabled: true}, {name: "Go option disabled", disable: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			file := &descriptorpb.FileDescriptorProto{Name: proto.String("demo/v1/item.proto"), Package: proto.String("demo.v1"), Options: &descriptorpb.FileOptions{
				GoPackage: proto.String("example.com/original;demov1"), JavaPackage: proto.String("original.java"), JavaMultipleFiles: proto.Bool(false), CcEnableArenas: proto.Bool(false),
			}, MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Item"), Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("value"), Number: proto.Int32(1), Type: descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(), Options: &descriptorpb.FieldOptions{Jstype: descriptorpb.FieldOptions_JS_NUMBER.Enum()}}}}}}
			before := proto.Clone(file).(*descriptorpb.FileDescriptorProto)
			cfg := ManagedModeConfig{Enabled: tt.enabled, GoPackageOnly: !tt.enabled, Override: []ManagedOverrideRule{
				{FileOption: FileOptionGoPackagePrefix, Value: "example.com/new"}, {FileOption: FileOptionJavaPackage, Value: "explicit.java"}, {FieldOption: FieldOptionJsType, Value: "JS_STRING"},
			}}
			if tt.disable {
				cfg.Disable = []ManagedDisableRule{{FileOption: FileOptionGoPackage}}
			}
			require.NoError(t, ApplyManagedMode([]*descriptorpb.FileDescriptorProto{file}, cfg, nil))
			if tt.disable {
				assert.True(t, proto.Equal(before, file))
				return
			}
			assert.Equal(t, "example.com/new/demo/v1;demov1", file.GetOptions().GetGoPackage())
			if tt.enabled {
				assert.Equal(t, "explicit.java", file.GetOptions().GetJavaPackage())
				assert.True(t, file.GetOptions().GetJavaMultipleFiles())
				assert.Equal(t, descriptorpb.FieldOptions_JS_STRING, file.MessageType[0].Field[0].GetOptions().GetJstype())
				return
			}
			file.Options.GoPackage = before.Options.GoPackage
			assert.True(t, proto.Equal(before, file), "non-Go options changed")
		})
	}
}
