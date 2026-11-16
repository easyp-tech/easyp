package options_test

import (
	"bytes"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/options"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

func parseFile(t *testing.T, path string, content string) *parser.ASTFile {
	t.Helper()
	src := []byte(content)
	ts := lexer.NewTokenStream(len(src)/8+16, len(src)/32+16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed for %s: %v", path, err)
	}
	p := parser.NewParser(src, ts, path)
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed for %s: %v", path, err)
	}
	return ast
}

func TestEvaluateCustomOptions(t *testing.T) {
	descProto := `
syntax = "proto2";
package google.protobuf;

message FileOptions {
    extensions 1000 to max;
}

message MessageOptions {
    extensions 1000 to max;
}

message FieldOptions {
    extensions 1000 to max;
}
`

	userProto := `
syntax = "proto2";
package test.opts;

import "google/protobuf/descriptor.proto";

enum Priority {
    LOW = 0;
    HIGH = 1;
}

message Meta {
    optional string author = 1;
    optional Priority prio = 2;
}

extend google.protobuf.FileOptions {
    optional string file_tag = 50001;
}

extend google.protobuf.MessageOptions {
    optional Meta msg_meta = 50002;
}

extend google.protobuf.FieldOptions {
    repeated int32 numbers = 50003;
}

option (file_tag) = "v1.0.0";

message SampleMsg {
    option (msg_meta) = {
        author: "Alice",
        prio: HIGH
    };

    optional string data = 1 [(numbers) = [10, 20]];
}
`

	fDesc := parseFile(t, "google/protobuf/descriptor.proto", descProto)
	fUser := parseFile(t, "user.proto", userProto)

	files := []*parser.ASTFile{fDesc, fUser}
	fileMap := map[string]*parser.ASTFile{
		"google/protobuf/descriptor.proto": fDesc,
		"user.proto":                       fUser,
	}

	table := linker.NewSymbolTable(files, fileMap)
	table.ResolveTypes()
	options.Evaluate(table)

	// 1. Verify FileOptions wire bytes
	if len(fUser.CustomOptionsWire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on fUser")
	}

	// 2. Verify MessageOptions wire bytes
	if len(fUser.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(fUser.Messages))
	}
	sampleMsg := &fUser.Messages[1]
	if len(sampleMsg.CustomOptionsWire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on SampleMsg")
	}

	// 3. Verify FieldOptions wire bytes
	if sampleMsg.Fields.Count != 1 {
		t.Fatalf("expected 1 field in SampleMsg")
	}
	dataField := &fUser.Fields[fUser.ChildIndices[sampleMsg.Fields.Start]]
	if len(dataField.CustomOptionsWire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on dataField")
	}
}

func TestNegativeFixedAndOneofOptions(t *testing.T) {
	descProto := `
syntax = "proto2";
package google.protobuf;

message OneofOptions {
    extensions 1000 to max;
}

message MessageOptions {
    extensions 1000 to max;
}
`
	userProto := `
syntax = "proto2";
package test.fixed;

import "google/protobuf/descriptor.proto";

message FixedScalars {
    optional sfixed32 sf32 = 1;
    optional sfixed64 sf64 = 2;
}

extend google.protobuf.OneofOptions {
    optional string oneof_tag = 60001;
}

extend google.protobuf.MessageOptions {
    optional FixedScalars fix_opt = 60002;
}

message TestMsg {
    option (fix_opt) = {
        sf32: -123,
        sf64: -456789
    };

    oneof my_oneof {
        option (oneof_tag) = "test_tag";
        int32 f1 = 1;
    }
}
`

	fDesc := parseFile(t, "google/protobuf/descriptor.proto", descProto)
	fUser := parseFile(t, "user.proto", userProto)

	files := []*parser.ASTFile{fDesc, fUser}
	fileMap := map[string]*parser.ASTFile{
		"google/protobuf/descriptor.proto": fDesc,
		"user.proto":                       fUser,
	}

	table := linker.NewSymbolTable(files, fileMap)
	table.ResolveTypes()
	options.Evaluate(table)

	if len(fUser.Oneofs) != 1 {
		t.Fatalf("expected 1 oneof, got %d", len(fUser.Oneofs))
	}
	oo := &fUser.Oneofs[0]
	if len(oo.CustomOptionsWire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on oneof")
	}

	if len(fUser.Messages) < 2 {
		t.Fatalf("expected at least 2 messages")
	}
	testMsg := &fUser.Messages[1]
	if len(testMsg.CustomOptionsWire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on TestMsg")
	}
	// Fixed32 for -123: 0xFFFFFF85 -> little-endian bytes: 85 ff ff ff
	// Fixed64 for -456789: 0xFFFFFFFFFFF907AB -> little-endian bytes: ab 07 f9 ff ff ff ff ff
	wire := testMsg.CustomOptionsWire
	expected32 := []byte{0x85, 0xff, 0xff, 0xff}
	expected64 := []byte{0xab, 0x07, 0xf9, 0xff, 0xff, 0xff, 0xff, 0xff}
	if !bytes.Contains(wire, expected32) {
		t.Errorf("expected wire to contain little-endian -123 (85 ff ff ff), wire: %x", wire)
	}
	if !bytes.Contains(wire, expected64) {
		t.Errorf("expected wire to contain little-endian -456789 (ab 07 f9 ff ff ff ff ff), wire: %x", wire)
	}
}

func TestBracketedExtensionEvaluation(t *testing.T) {
	descProto := `
syntax = "proto2";
package google.protobuf;
message MessageOptions {
    extensions 1000 to max;
}
`
	userProto := `
syntax = "proto2";
package test.bracketopts;
import "google/protobuf/descriptor.proto";

message Container {
    optional string base_info = 1;
    extensions 100 to 1000;
}

extend Container {
    optional int32 ext_int = 101;
    optional string ext_str = 102;
    optional Container nested_ext = 103;
}

extend google.protobuf.MessageOptions {
    optional Container msg_opt = 60001;
}

message TestMsg {
    option (msg_opt) = {
        base_info: "root"
        [test.bracketopts.ext_int]: 42
        [ext_str]: "relative_bracket"
        [nested_ext]: <
            base_info: "nested"
            [ext_int]: 100
        >
    };
}
`

	fDesc := parseFile(t, "google/protobuf/descriptor.proto", descProto)
	fUser := parseFile(t, "user.proto", userProto)

	files := []*parser.ASTFile{fDesc, fUser}
	fileMap := map[string]*parser.ASTFile{
		"google/protobuf/descriptor.proto": fDesc,
		"user.proto":                       fUser,
	}

	table := linker.NewSymbolTable(files, fileMap)
	table.ResolveTypes()
	options.Evaluate(table)

	if len(fUser.Messages) < 2 {
		t.Fatalf("expected at least 2 messages")
	}
	testMsg := &fUser.Messages[1]
	wire := testMsg.CustomOptionsWire
	if len(wire) == 0 {
		t.Fatalf("expected non-empty CustomOptionsWire on TestMsg")
	}

	// Extension 101 is int32: tag = (101 << 3) | 0 = 0x328 -> varint [0xa8, 0x06]
	// value 42 = 0x2a
	ext42Bytes := []byte{0xa8, 0x06, 0x2a}
	if !bytes.Contains(wire, ext42Bytes) {
		t.Errorf("expected wire to contain ext_int=42 (a8 06 2a), got: %x", wire)
	}

	// Extension 101 nested value 100 = 0x64
	ext100Bytes := []byte{0xa8, 0x06, 0x64}
	if !bytes.Contains(wire, ext100Bytes) {
		t.Errorf("expected wire to contain nested ext_int=100 (a8 06 64), got: %x", wire)
	}

	// Verify strings "root", "relative_bracket", "nested"
	if !bytes.Contains(wire, []byte("root")) {
		t.Errorf("expected wire to contain 'root', got: %x", wire)
	}
	if !bytes.Contains(wire, []byte("relative_bracket")) {
		t.Errorf("expected wire to contain 'relative_bracket', got: %x", wire)
	}
	if !bytes.Contains(wire, []byte("nested")) {
		t.Errorf("expected wire to contain 'nested', got: %x", wire)
	}
}


