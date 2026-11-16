package protoc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
)

// Malformed declarations must yield a diagnostic, not a panic, a hang or an
// index aliasing an earlier declaration.
func TestNamelessDeclarationsReportErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{name: "message_first", src: `package p; message {`},
		{name: "message_after_message", src: `syntax = "proto3"; message A {} message {`},
		{name: "enum_first", src: `syntax = "proto3"; enum {`},
		{name: "service_first", src: `syntax = "proto3"; service {`},
		{name: "nested_message", src: `syntax = "proto3"; message A { message { } }`},
		{name: "nested_enum", src: `syntax = "proto3"; message A { enum { } }`},
		{name: "stray_close_brace", src: "syntax = \"proto3\";\n}\nmessage A {}\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := protoc.MemResolver{"bad.proto": []byte(tc.src)}
			_, err := protoc.Compile(context.Background(), []string{"bad.proto"}, protoc.WithResolver(files))
			if !errors.Is(err, protoc.ErrCompilationFailed) {
				t.Fatalf("Compile error = %v, want ErrCompilationFailed", err)
			}
		})
	}
}
