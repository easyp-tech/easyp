package serialize_test

import (
	"os"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/serialize"
)

func TestSerializeDescriptorProto(t *testing.T) {
	path := "../../protoc/wkt/google/protobuf/descriptor.proto"
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}

	ts := lexer.NewTokenStream(len(src)/8+16, len(src)/32+16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "google/protobuf/descriptor.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	s := serialize.NewWireSerializer()
	outBytes := s.SerializeFileDescriptorSet([]*parser.ASTFile{ast}, false)
	if len(outBytes) == 0 {
		t.Fatalf("serialized bytes empty")
	}

	// FileDescriptorSet has tag 1, wire type 2 -> byte 0x0a
	if outBytes[0] != 0x0a {
		t.Errorf("expected tag 1 wire type 2 (0x0a), got 0x%02x", outBytes[0])
	}

	_ = os.WriteFile("/tmp/easyp_wkt.pb", outBytes, 0644)
	t.Logf("Serialized FileDescriptorSet: %d bytes", len(outBytes))
}

func TestSerializeProto3DesugaredElements(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.pkg;

message ComplexMsg {
    int32 id = 1;
    optional string opt_val = 2;
    map<string, int32> entries = 3;
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex: %v", err)
	}
	p := parser.NewParser(src, ts, "complex.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	s := serialize.NewWireSerializer()
	outBytes := s.SerializeFileDescriptorSet([]*parser.ASTFile{ast}, false)
	if len(outBytes) == 0 {
		t.Fatalf("serialized bytes empty")
	}
	if outBytes[0] != 0x0a {
		t.Errorf("expected 0x0a, got 0x%02x", outBytes[0])
	}
}

func TestSerializeMultiFileIsolation(t *testing.T) {
	src1 := []byte(`syntax = "proto3"; package test.dep; message DepMessage { int32 id = 1; }`)
	src2 := []byte(`syntax = "proto3"; package test.main; import "dep.proto"; message MainMessage { int32 id = 1; }`)

	ts1 := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src1, ts1); err != nil {
		t.Fatal(err)
	}
	p1 := parser.NewParser(src1, ts1, "dep.proto")
	ast1, err := p1.ParseFile()
	if err != nil {
		t.Fatal(err)
	}

	ts2 := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src2, ts2); err != nil {
		t.Fatal(err)
	}
	p2 := parser.NewParser(src2, ts2, "main.proto")
	ast2, err := p2.ParseFile()
	if err != nil {
		t.Fatal(err)
	}

	s := serialize.NewWireSerializer()
	outBytes := s.SerializeFileDescriptorSet([]*parser.ASTFile{ast1, ast2}, false)
	if len(outBytes) == 0 {
		t.Fatalf("expected non-empty serialized bytes")
	}

	// Verify that the first byte is Tag 1, WireType 2 (0x0a)
	if outBytes[0] != 0x0a {
		t.Errorf("expected 0x0a, got 0x%02x", outBytes[0])
	}
}
