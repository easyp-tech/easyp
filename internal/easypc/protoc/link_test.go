package protoc

import (
	"context"
	"testing"
)

func TestLinkRecordsUsesAndMethodTypes(t *testing.T) {
	files := MemResolver{
		"a.proto": []byte(`syntax = "proto3";
package a;
import "b.proto";
import "unused.proto";
import public "pub.proto";
import "opt.proto";
message Req { b.B b = 1; }
service S {
  rpc Get(Req) returns (.b.B) { option (o.tag) = 1; }
}
`),
		"b.proto":      []byte(`syntax = "proto3"; package b; message B {}`),
		"unused.proto": []byte(`syntax = "proto3"; package u; message U {}`),
		"pub.proto":    []byte(`syntax = "proto3"; package p; message P {}`),
		"opt.proto": []byte(`syntax = "proto3"; package o; import "google/protobuf/descriptor.proto";
extend google.protobuf.MethodOptions { int32 tag = 50000; }`),
	}
	linked, err := Link(context.Background(), []string{"a.proto"}, WithResolver(files), WithStandardImports(true))
	if err != nil {
		t.Fatalf("Link: %v", err)
	}
	a, ok := linked.File("a.proto")
	if !ok {
		t.Fatal("a.proto missing")
	}
	for _, want := range []string{"b.proto", "opt.proto"} {
		if !a.Uses[want] {
			t.Errorf("a.proto should use %s, uses %v", want, a.Uses)
		}
	}
	for _, unwanted := range []string{"unused.proto", "pub.proto", "a.proto"} {
		if a.Uses[unwanted] {
			t.Errorf("a.proto should not use %s", unwanted)
		}
	}
	if len(a.Imports) != 4 || !a.Imports[2].Public || a.Imports[2].Path != "pub.proto" {
		t.Errorf("imports = %+v", a.Imports)
	}
	if len(a.Methods) != 1 || a.Methods[0].Input != "a.Req" || a.Methods[0].Output != "b.B" {
		t.Errorf("methods = %+v", a.Methods)
	}
	if _, ok := linked.File("b.proto"); !ok {
		t.Error("imports are linked too")
	}
}

func TestLinkNeverReadsTheFileSystem(t *testing.T) {
	_, err := Link(context.Background(), []string{"link_test.go"}, WithResolver(MemResolver{}))
	if err == nil {
		t.Fatal("Link read a file that no resolver provides")
	}
}
