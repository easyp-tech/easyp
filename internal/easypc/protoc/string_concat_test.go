package protoc_test

import (
	"context"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
)

// protoc's ConsumeString joins adjacent string literals in syntax, import and
// reserved statements, not only in option values.
func TestAdjacentStringLiteralsConcatenate(t *testing.T) {
	files := protoc.MemResolver{
		"svc.proto":  []byte(`syntax = "proto3"; package service; message Service {}`),
		"main.proto": []byte(`syntax = "pro" "to3"; import "s\x76c." /* c */ "proto"; message C { service.Service s = 1; reserved "fo" "o"; }`),
	}
	res, err := protoc.Compile(context.Background(), []string{"main.proto"}, protoc.WithResolver(files))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	file := res.DescriptorSet().GetFile()[0]
	if got := file.GetDependency(); len(got) != 1 || got[0] != "svc.proto" {
		t.Fatalf("dependency = %v, want [svc.proto]", got)
	}
	if got := file.GetMessageType()[0].GetReservedName(); len(got) != 1 || got[0] != "foo" {
		t.Fatalf("reserved names = %v, want [foo]", got)
	}
}
