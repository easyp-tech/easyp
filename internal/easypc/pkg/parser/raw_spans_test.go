package parser_test

import (
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

func parseSrc(t *testing.T, src string) *parser.ASTFile {
	t.Helper()
	b := []byte(src)
	ts := lexer.NewTokenStream(64, 16)
	if err := lexer.Lex(b, ts); err != nil {
		t.Fatalf("lex failed: %v", err)
	}
	ast, _ := parser.NewParser(b, ts, "raw.proto").ParseFile()
	return ast
}

func TestOptionRawSpans(t *testing.T) {
	src := `syntax = "proto2";
option go_package = "a/b" /* c */ "c;d"; // trailing
option ( foo.bar ).baz = -12;
message M {
  optional string s = 1 [default = "x" "y", (ext).sub = { a: 1 }, json_name = "S"];
}
`
	ast := parseSrc(t, src)
	if len(ast.Errors) != 0 {
		t.Fatalf("unexpected errors: %v", ast.Errors)
	}
	if got := ast.SrcLen; got != uint32(len(src)) {
		t.Errorf("SrcLen = %d, want %d", got, len(src))
	}

	want := map[string]string{
		"go_package":      `"a/b" /* c */ "c;d"`,
		"( foo.bar ).baz": `-12`,
		"(ext).sub":       `{ a: 1 }`,
	}
	got := map[string]string{}
	for _, opt := range ast.Options {
		if opt.NameSpan.IsEmpty() {
			continue
		}
		got[opt.NameSpan.String(ast.Src)] = opt.ValRawSpan.String(ast.Src)
	}
	for name, raw := range want {
		if got[name] != raw {
			t.Errorf("option %q: raw value %q, want %q (all: %q)", name, got[name], raw, got)
		}
	}

	fld := ast.Fields[0]
	if s := fld.DefaultOpt.String(ast.Src); s != `default = "x" "y"` {
		t.Errorf("DefaultOpt = %q", s)
	}
	if s := fld.CustomJSONNameOpt.String(ast.Src); s != `json_name = "S"` {
		t.Errorf("CustomJSONNameOpt = %q", s)
	}
}

func TestSemanticDiagnostics(t *testing.T) {
	ast := parseSrc(t, `syntax = "proto3"; message M { int32 a = 1; int32 a = 1; int32 b = 0; }`)
	if len(ast.Errors) == 0 {
		t.Fatal("expected diagnostics")
	}
	for _, d := range ast.Errors {
		if !d.Semantic {
			t.Errorf("diagnostic %q should be semantic", d.Message)
		}
	}

	ast = parseSrc(t, `syntax = "proto3"; message M { int32 a = ; }`)
	if len(ast.Errors) == 0 || ast.Errors[0].Semantic {
		t.Errorf("syntax error must not be semantic: %v", ast.Errors)
	}
}
