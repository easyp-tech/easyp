package protoc_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
)

// Every type reference must resolve to the right kind of symbol, as in protoc.
func TestReferencesReportProtocDiagnostics(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{name: "extension_type", src: `import "google/protobuf/descriptor.proto"; extend google.protobuf.FieldOptions { Missing m = 50000; }`, want: `"Missing" is not defined.`},
		{name: "nested_extension_type", src: `import "google/protobuf/descriptor.proto"; message W { extend google.protobuf.FieldOptions { Missing m = 50000; } }`, want: `"Missing" is not defined.`},
		{name: "extendee", src: `extend Nope { string m = 50000; }`, want: `"Nope" is not defined.`},
		{name: "extendee_enum", src: `enum E { E_UNSPECIFIED = 0; } extend E { string m = 50000; }`, want: `"E" is not a message type.`},
		{name: "rpc_input", src: `message W {} service S { rpc M(Missing) returns (W); }`, want: `"Missing" is not defined.`},
		{name: "rpc_output_enum", src: `message W {} enum E { E_UNSPECIFIED = 0; } service S { rpc M(W) returns (E); }`, want: `"E" is not a message type.`},
		{name: "field_type_is_service", src: `message W {} service S { rpc M(W) returns (W); } message X { .S s = 1; }`, want: `".S" is not a type.`},
		{name: "field_type_skips_service", src: `message W {} service S { rpc M(W) returns (W); } message X { S s = 1; }`, want: `"S" is not defined.`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diags []string
			files := protoc.MemResolver{"a.proto": []byte(`syntax = "proto3"; ` + tc.src)}
			_, err := protoc.Compile(context.Background(), []string{"a.proto"}, protoc.WithResolver(files), protoc.WithStandardImports(true),
				protoc.WithErrorHandler(func(d protoc.Diagnostic) { diags = append(diags, d.Message) }))
			if !errors.Is(err, protoc.ErrCompilationFailed) {
				t.Fatalf("Compile error = %v, want ErrCompilationFailed", err)
			}
			if !strings.Contains(strings.Join(diags, "\n"), tc.want) {
				t.Fatalf("diagnostics = %q, want %q", diags, tc.want)
			}
		})
	}
}

// A single-component field type skips a same-named field, as protoc's LOOKUP_TYPES does.
func TestFieldTypeSkipsSameNamedField(t *testing.T) {
	files := protoc.MemResolver{"a.proto": []byte(`syntax = "proto3"; package p; message Status {} message M { Status Status = 1; }`)}
	res, err := protoc.Compile(context.Background(), []string{"a.proto"}, protoc.WithResolver(files))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	field := res.DescriptorSet().GetFile()[0].GetMessageType()[1].GetField()[0]
	if field.GetTypeName() != ".p.Status" {
		t.Fatalf("type_name = %q, want .p.Status", field.GetTypeName())
	}
}
