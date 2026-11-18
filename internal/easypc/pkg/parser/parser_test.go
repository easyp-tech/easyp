package parser_test

import (
	"os"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

func TestParseDescriptorProto(t *testing.T) {
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

	if ast.Package.String(ast.Src) != "google.protobuf" {
		t.Errorf("expected package google.protobuf, got %q", ast.Package.String(ast.Src))
	}

	if ast.TopLevelMessages.Count != 23 {
		t.Errorf("expected exactly 23 top-level messages, got %d", ast.TopLevelMessages.Count)
	}
	if ast.TopLevelEnums.Count != 2 {
		t.Errorf("expected exactly 2 top-level enums, got %d", ast.TopLevelEnums.Count)
	}

	t.Logf("Successfully parsed descriptor.proto: %d top messages (%d total), %d fields, %d top enums (%d total)",
		ast.TopLevelMessages.Count, len(ast.Messages), len(ast.Fields), ast.TopLevelEnums.Count, len(ast.Enums))
}

func TestJSONName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"foo_bar", "fooBar"},
		{"_foo", "Foo"},
		{"foo__bar", "fooBar"},
		{"FOO_BAR", "FOOBAR"},
		{"foo_", "foo"},
		{"camelCase", "camelCase"},
		{"a_b_c", "aBC"},
	}

	for _, tc := range cases {
		got := parser.JSONName(tc.input)
		if got != tc.want {
			t.Errorf("JSONName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseProto2Features(t *testing.T) {
	src := []byte(`
syntax = "proto2";
package test.proto2;

message Outer {
    required int32 id = 1;
    optional string name = 2 [default = "hello"];
    repeated int64 tags = 3;

    optional group MyGroup = 4 {
        optional int32 nested_val = 1;
    }

    extensions 100 to 199;
}

extend Outer {
    optional string ext_field = 101;
}
`)

	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "test_proto2.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Syntax != parser.SyntaxProto2 {
		t.Errorf("expected syntax proto2, got %v", ast.Syntax)
	}

	if ast.TopLevelMessages.Count != 1 {
		t.Fatalf("expected 1 top-level message, got %d", ast.TopLevelMessages.Count)
	}

	outerMsgIdx := ast.ChildIndices[ast.TopLevelMessages.Start]
	outer := &ast.Messages[outerMsgIdx]
	if outer.Name.String(ast.Src) != "Outer" {
		t.Errorf("expected message name Outer, got %q", outer.Name.String(ast.Src))
	}

	// Group desugaring: Outer should have a nested message "MyGroup"
	if outer.NestedMessages.Count != 1 {
		t.Fatalf("expected 1 nested message for group, got %d", outer.NestedMessages.Count)
	}
	nestedIdx := ast.ChildIndices[outer.NestedMessages.Start]
	if ast.Messages[nestedIdx].Name.String(ast.Src) != "MyGroup" {
		t.Errorf("expected nested group message MyGroup, got %q", ast.Messages[nestedIdx].Name.String(ast.Src))
	}

	// Extensions
	if len(ast.Extensions) != 1 {
		t.Fatalf("expected 1 extension field, got %d", len(ast.Extensions))
	}
	ext := &ast.Extensions[0]
	if ext.Extendee.String(ast.Src) != ".test.proto2.Outer" {
		t.Errorf("expected extendee .test.proto2.Outer, got %q", ext.Extendee.String(ast.Src))
	}
}

func TestParseProto3Features(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.proto3;

message SampleMessage {
    int32 regular_field = 1;
    optional string opt_field = 2;
    map<string, int64> score_map = 3;

    oneof choice {
        string choice_a = 4;
        int32 choice_b = 5;
    }
}

service SampleService {
    rpc UnaryCall(SampleMessage) returns (SampleMessage);
    rpc StreamCall(stream SampleMessage) returns (stream SampleMessage);
}
`)

	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "test_proto3.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Syntax != parser.SyntaxProto3 {
		t.Errorf("expected syntax proto3, got %v", ast.Syntax)
	}

	msg := &ast.Messages[ast.ChildIndices[ast.TopLevelMessages.Start]]

	// Proto3 optional desugaring: synthetic oneof named _opt_field
	// We have 1 user oneof (choice) + 1 synthetic oneof (_opt_field) = 2 oneofs
	if msg.Oneofs.Count != 2 {
		t.Fatalf("expected 2 oneofs (1 regular + 1 synthetic optional), got %d", msg.Oneofs.Count)
	}
	synthOneofIdx := ast.ChildIndices[msg.Oneofs.Start+1]
	synthOneof := &ast.Oneofs[synthOneofIdx]
	if synthOneof.Name.String(ast.Src) != "_opt_field" {
		t.Errorf("expected synthetic oneof _opt_field, got %q", synthOneof.Name.String(ast.Src))
	}

	// Map desugaring: nested ScoreMapEntry message with map_entry = true
	if msg.NestedMessages.Count != 1 {
		t.Fatalf("expected 1 nested message for map entry, got %d", msg.NestedMessages.Count)
	}
	mapEntryIdx := ast.ChildIndices[msg.NestedMessages.Start]
	mapEntry := &ast.Messages[mapEntryIdx]
	if mapEntry.Name.String(ast.Src) != "ScoreMapEntry" {
		t.Errorf("expected map entry name ScoreMapEntry, got %q", mapEntry.Name.String(ast.Src))
	}
	if !mapEntry.IsMapEntry {
		t.Errorf("expected IsMapEntry=true")
	}

	// Service
	if len(ast.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(ast.Services))
	}
	svc := &ast.Services[0]
	if svc.Methods.Count != 2 {
		t.Fatalf("expected 2 methods, got %d", svc.Methods.Count)
	}
	m0 := &ast.Methods[ast.ChildIndices[svc.Methods.Start]]
	m1 := &ast.Methods[ast.ChildIndices[svc.Methods.Start+1]]
	if m0.ClientStreaming || m0.ServerStreaming {
		t.Errorf("expected unary for m0")
	}
	if !m1.ClientStreaming || !m1.ServerStreaming {
		t.Errorf("expected bidi streaming for m1")
	}
}

func TestParseEditions2023(t *testing.T) {
	src := []byte(`
edition = "2023";
package test.editions;

message EditionMessage {
    int32 id = 1;
    string title = 2;
    reserved 10 to 20;
    reserved "old_name", "deprecated_field";
}
`)

	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "test_editions.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Syntax != parser.SyntaxEditions || ast.Edition != parser.Edition2023 {
		t.Errorf("expected edition 2023, got syntax=%v edition=%v", ast.Syntax, ast.Edition)
	}

	msg := &ast.Messages[ast.ChildIndices[ast.TopLevelMessages.Start]]
	if msg.ReservedRanges.Count != 1 {
		t.Errorf("expected 1 reserved range, got %d", msg.ReservedRanges.Count)
	}
	if msg.ReservedNames.Count != 2 {
		t.Errorf("expected 2 reserved names, got %d", msg.ReservedNames.Count)
	}
}

func TestSyntaxErrorNoPanic(t *testing.T) {
	src := []byte(`syntax = "proto3"; message Foo { invalid syntax here }`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "syntax_err.proto")
	ast, err := p.ParseFile()
	if err == nil {
		t.Fatalf("expected parse error, got nil")
	}
	if len(ast.Errors) == 0 {
		t.Fatalf("expected ast errors to be populated")
	}
}

func TestUnclosedBraceRejection(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unclosed_message", "syntax = \"proto3\"; message Foo { int32 a = 1;"},
		{"unclosed_enum", "syntax = \"proto3\"; enum Bar { ZERO = 0;"},
		{"unclosed_service", "syntax = \"proto3\"; service Svc { rpc Call(A) returns (B);"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.src)
			ts := lexer.NewTokenStream(64, 16)
			_ = lexer.Lex(src, ts)
			p := parser.NewParser(src, ts, "unclosed.proto")
			ast, err := p.ParseFile()
			if err == nil {
				t.Fatalf("expected parse error for unclosed brace, got nil")
			}
			foundExpectedBrace := false
			for _, diag := range ast.Errors {
				if diag.Message == "expected '}'" {
					foundExpectedBrace = true
					break
				}
			}
			if !foundExpectedBrace {
				t.Errorf("expected diagnostic \"expected '}'\", got: %v", ast.Errors)
			}
		})
	}
}

func TestMalformedOptionRecovery(t *testing.T) {
	src := []byte(`syntax = "proto2"; message M { optional string s = 1 [default = 'single '\'quote']; }`)
	ts := lexer.NewTokenStream(64, 16)
	_ = lexer.Lex(src, ts)

	p := parser.NewParser(src, ts, "repro_oom.proto")
	ast, err := p.ParseFile()
	if err == nil {
		t.Fatalf("expected parse error for malformed option, got nil")
	}
	if len(ast.Errors) == 0 {
		t.Fatalf("expected errors recorded")
	}
}

func TestNegativeDefaultValue(t *testing.T) {
	src := []byte(`syntax = "proto2"; package test; message M { optional int32 num = 1 [default = -67890]; }`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "neg_def.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(ast.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(ast.Fields))
	}
	fld := &ast.Fields[0]
	defaultStr := fld.DefaultVal.String(ast.Src)
	if defaultStr != "-67890" {
		t.Errorf("expected default value -67890, got %q", defaultStr)
	}
}

func TestServiceMethodAndMapQualification(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.pkg;

message Ping {}
message Pong {}

message Container {
    map<string, int32> dict = 1;
}

service TestService {
    rpc Exchange(Ping) returns (Pong);
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "svc_qual.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(ast.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(ast.Methods))
	}
	m := &ast.Methods[0]
	if m.InputType.String(ast.Src) != ".test.pkg.Ping" {
		t.Errorf("expected input type .test.pkg.Ping, got %q", m.InputType.String(ast.Src))
	}
	if m.OutputType.String(ast.Src) != ".test.pkg.Pong" {
		t.Errorf("expected output type .test.pkg.Pong, got %q", m.OutputType.String(ast.Src))
	}

	// Check map field synthetic type name qualification
	foundMapField := false
	for _, fld := range ast.Fields {
		if fld.Name.String(ast.Src) == "dict" {
			foundMapField = true
			if fld.TypeName.String(ast.Src) != ".test.pkg.Container.DictEntry" {
				t.Errorf("expected map entry type .test.pkg.Container.DictEntry, got %q", fld.TypeName.String(ast.Src))
			}
		}
	}
	if !foundMapField {
		t.Errorf("did not find map field dict")
	}
}

func TestRPCMethodOptionAggregate(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.rpc;

message Req {}
message Resp {}

service ApiService {
    rpc Call(Req) returns (Resp) {
        option (custom.rule) = {
            get: "/v1/test"
            additional_bindings {
                get: "/v1/alt"
            }
        };
    }
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "rpc_opt.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ast.Errors) != 0 {
		t.Fatalf("expected 0 errors, got: %v", ast.Errors)
	}
}

func TestOptionListAndAngleBracketLiterals(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.opt;

message MyMsg {
    option (custom.meta) = <
        key: "val"
        items: [ "a", "b", "c" ]
        nested: < count: 42 >
    >;
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "opt_test.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ast.Errors) != 0 {
		t.Fatalf("expected 0 errors, got: %v", ast.Errors)
	}

	if len(ast.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(ast.Messages))
	}
	msg := &ast.Messages[0]
	if msg.Options.Count != 1 {
		t.Fatalf("expected 1 option, got %d", msg.Options.Count)
	}
	opt := &ast.Options[ast.ChildIndices[msg.Options.Start]]
	if opt.ValKind != parser.OptionValAggregate {
		t.Errorf("expected OptionValAggregate, got %v", opt.ValKind)
	}
}

func TestOneofOptions(t *testing.T) {
	src := []byte(`
syntax = "proto3";
package test.oneof;

message Msg {
    oneof choice {
        option (custom_oneof_opt) = "oneof_val";
        string a = 1;
        int32 b = 2;
    }
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "oneof_opt.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ast.Errors) != 0 {
		t.Fatalf("expected 0 errors, got: %v", ast.Errors)
	}
	if len(ast.Oneofs) != 1 {
		t.Fatalf("expected 1 oneof, got %d", len(ast.Oneofs))
	}
	oo := &ast.Oneofs[0]
	if oo.Options.Count != 1 {
		t.Fatalf("expected 1 oneof option, got %d", oo.Options.Count)
	}
	opt := &ast.Options[ast.ChildIndices[oo.Options.Start]]
	if opt.ValSpan.String(ast.Src) != "oneof_val" {
		t.Errorf("expected option value \"oneof_val\", got %s", opt.ValSpan.String(ast.Src))
	}
}

func TestBracketedExtensionInAggregate(t *testing.T) {
	src := []byte(`
syntax = "proto2";
package test.brackets;

message MyMsg {
    option (my_opt) = {
        normal_field: "hello"
        [pkg.ext_int]: 42
        (pkg.ext_str): "paren_ext"
        [nested_ext]: <
            [sub_ext]: 100
        >
    };
}
`)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(src, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}

	p := parser.NewParser(src, ts, "brackets_test.proto")
	ast, err := p.ParseFile()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(ast.Errors) != 0 {
		t.Fatalf("expected 0 errors, got: %v", ast.Errors)
	}

	if len(ast.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(ast.Messages))
	}
	msg := &ast.Messages[0]
	if msg.Options.Count != 1 {
		t.Fatalf("expected 1 option, got %d", msg.Options.Count)
	}
	opt := &ast.Options[ast.ChildIndices[msg.Options.Start]]
	if opt.ValKind != parser.OptionValAggregate {
		t.Fatalf("expected OptionValAggregate, got %v", opt.ValKind)
	}

	// Verify suboptions
	if opt.SubOptions.Count != 4 {
		t.Fatalf("expected 4 sub-options, got %d", opt.SubOptions.Count)
	}

	// 1: normal_field (IsExtension = false)
	sub0 := &ast.Options[ast.ChildIndices[opt.SubOptions.Start]]
	part0 := &ast.OptionNameParts[sub0.NameParts.Start]
	if part0.IsExtension || part0.Name.String(ast.Src) != "normal_field" {
		t.Errorf("expected normal_field non-extension, got %s (ext=%v)", part0.Name.String(ast.Src), part0.IsExtension)
	}

	// 2: [pkg.ext_int] (IsExtension = true)
	sub1 := &ast.Options[ast.ChildIndices[opt.SubOptions.Start+1]]
	part1 := &ast.OptionNameParts[sub1.NameParts.Start]
	if !part1.IsExtension || part1.Name.String(ast.Src) != "pkg.ext_int" {
		t.Errorf("expected pkg.ext_int extension, got %s (ext=%v)", part1.Name.String(ast.Src), part1.IsExtension)
	}

	// 3: (pkg.ext_str) (IsExtension = true)
	sub2 := &ast.Options[ast.ChildIndices[opt.SubOptions.Start+2]]
	part2 := &ast.OptionNameParts[sub2.NameParts.Start]
	if !part2.IsExtension || part2.Name.String(ast.Src) != "pkg.ext_str" {
		t.Errorf("expected pkg.ext_str extension, got %s (ext=%v)", part2.Name.String(ast.Src), part2.IsExtension)
	}

	// 4: [nested_ext] with sub_ext in angle brackets
	sub3 := &ast.Options[ast.ChildIndices[opt.SubOptions.Start+3]]
	part3 := &ast.OptionNameParts[sub3.NameParts.Start]
	if !part3.IsExtension || part3.Name.String(ast.Src) != "nested_ext" {
		t.Errorf("expected nested_ext extension, got %s (ext=%v)", part3.Name.String(ast.Src), part3.IsExtension)
	}
	if sub3.SubOptions.Count != 1 {
		t.Fatalf("expected 1 nested sub-option, got %d", sub3.SubOptions.Count)
	}
	nestedSub := &ast.Options[ast.ChildIndices[sub3.SubOptions.Start]]
	nestedPart := &ast.OptionNameParts[nestedSub.NameParts.Start]
	if !nestedPart.IsExtension || nestedPart.Name.String(ast.Src) != "sub_ext" {
		t.Errorf("expected sub_ext extension, got %s (ext=%v)", nestedPart.Name.String(ast.Src), nestedPart.IsExtension)
	}
}

