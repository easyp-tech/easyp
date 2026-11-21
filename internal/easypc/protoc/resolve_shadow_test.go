package protoc_test

import (
	"context"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
)

// protoc skips non-aggregate symbols (fields) when resolving the first
// component of a qualified type name.
func TestQualifiedNameSkipsShadowingField(t *testing.T) {
	files := protoc.MemResolver{
		"dep.proto":  []byte(`syntax = "proto3"; package service; message Service {}`),
		"main.proto": []byte(`syntax = "proto3"; import "dep.proto"; message C { service.Service service = 1; }`),
	}
	if _, err := protoc.Compile(context.Background(), []string{"main.proto"}, protoc.WithResolver(files)); err != nil {
		t.Fatalf("Compile: %v", err)
	}
}
