package lexer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
)

func TestTabstop8LineCol(t *testing.T) {
	// Empirical protoc 36.2 observations:
	// "\tint32 x = 0x;" -> col 9 (0-based 8 + 1)
	src1 := []byte("\tint32")
	ts1 := lexer.NewTokenStream(10, 10)
	if err := lexer.Lex(src1, ts1); err != nil {
		t.Fatal(err)
	}
	line, col := ts1.LineCol(src1, ts1.Starts[0])
	if line != 1 || col != 9 {
		t.Errorf("expected line 1 col 9 for \\tint32, got line %d col %d", line, col)
	}

	// 8 spaces + tab -> col 17
	src2 := []byte("        \tint32")
	ts2 := lexer.NewTokenStream(10, 10)
	if err := lexer.Lex(src2, ts2); err != nil {
		t.Fatal(err)
	}
	line, col = ts2.LineCol(src2, ts2.Starts[0])
	if line != 1 || col != 17 {
		t.Errorf("expected line 1 col 17 for 8 spaces + tab, got line %d col %d", line, col)
	}

	// Multi-byte UTF-8 character advances column by number of bytes
	// "/* я */ int32" -> "я" is 2 bytes (0xD1 0x8F).
	// "/* я */ " is 2 + 1 + 2 + 1 + 2 + 1 = 9 bytes.
	// So int32 is at offset 9 -> col 10.
	src3 := []byte("/* я */ int32")
	ts3 := lexer.NewTokenStream(10, 10)
	if err := lexer.Lex(src3, ts3); err != nil {
		t.Fatal(err)
	}
	line, col = ts3.LineCol(src3, ts3.Starts[1]) // int32 token
	if line != 1 || col != 10 {
		t.Errorf("expected line 1 col 10 for UTF-8 byte column, got line %d col %d", line, col)
	}
}

func TestNumberLexing(t *testing.T) {
	tests := []struct {
		input string
		kinds []lexer.Kind
	}{
		{"0", []lexer.Kind{lexer.KindInt}},
		{"1", []lexer.Kind{lexer.KindInt}},
		{"42", []lexer.Kind{lexer.KindInt}},
		{"017", []lexer.Kind{lexer.KindInt}},
		{"0x1F", []lexer.Kind{lexer.KindInt}},
		{"0x0", []lexer.Kind{lexer.KindInt}},
		{"1.", []lexer.Kind{lexer.KindFloat}},
		{"1.0", []lexer.Kind{lexer.KindFloat}},
		{"123.456", []lexer.Kind{lexer.KindFloat}},
		{".5", []lexer.Kind{lexer.KindFloat}},
		{".5e-3", []lexer.Kind{lexer.KindFloat}},
		{"1e5", []lexer.Kind{lexer.KindFloat}},
		{"1e+5", []lexer.Kind{lexer.KindFloat}},
		{"-1", []lexer.Kind{lexer.KindPunct, lexer.KindInt}},
		{"0x1e+5", []lexer.Kind{lexer.KindInt, lexer.KindPunct, lexer.KindInt}},
		{"09", []lexer.Kind{lexer.KindError}},
		{"08", []lexer.Kind{lexer.KindError}},
		{"0x", []lexer.Kind{lexer.KindError}},
		{"0X", []lexer.Kind{lexer.KindError}},
		{"1foo", []lexer.Kind{lexer.KindError}},
		{"1..2", []lexer.Kind{lexer.KindError}},
		{"1e", []lexer.Kind{lexer.KindError}},
		{"1e+", []lexer.Kind{lexer.KindError}},
		{"1.0f", []lexer.Kind{lexer.KindError}},
		{".e3", []lexer.Kind{lexer.KindPunct, lexer.KindIdent}},
	}

	for _, tc := range tests {
		ts := lexer.NewTokenStream(10, 10)
		if err := lexer.Lex([]byte(tc.input), ts); err != nil {
			t.Fatalf("Lex(%q) error: %v", tc.input, err)
		}
		if ts.Len() != len(tc.kinds) {
			t.Fatalf("Lex(%q) got %d tokens, want %d", tc.input, ts.Len(), len(tc.kinds))
		}
		for i, wantKind := range tc.kinds {
			if ts.Kinds[i] != wantKind {
				t.Errorf("Lex(%q) token %d: got %s, want %s", tc.input, i, ts.Kinds[i], wantKind)
			}
		}
	}
}

func TestStringLexing(t *testing.T) {
	tests := []struct {
		input    string
		wantKind lexer.Kind
	}{
		{`""`, lexer.KindString},
		{`''`, lexer.KindString},
		{`"hello world"`, lexer.KindString},
		{`"escaped \" quote"`, lexer.KindString},
		{`"\a\b\f\n\r\t\v\\\'\"\?"`, lexer.KindString},
		{`"\x0\x1f\xA\xFF"`, lexer.KindString},
		{`"\0\7\77\377"`, lexer.KindString},
		{`"\u1234\uABCD\U0001F600"`, lexer.KindString},
		{`"hello мир 🎉"`, lexer.KindString},
		{`"\c"`, lexer.KindError},       // bad escape
		{`"\x"`, lexer.KindError},       // incomplete hex
		{`"\xz"`, lexer.KindError},      // invalid hex digit
		{`"\u12"`, lexer.KindError},     // incomplete unicode
		{`"\U0001"`, lexer.KindError},   // incomplete unicode 8-hex
		{"\"unclosed", lexer.KindError}, // unclosed string at EOF
		{"\"unclosed\n\"", lexer.KindError},
	}

	for _, tc := range tests {
		ts := lexer.NewTokenStream(10, 10)
		if err := lexer.Lex([]byte(tc.input), ts); err != nil {
			t.Fatalf("Lex(%q) error: %v", tc.input, err)
		}
		if ts.Len() == 0 {
			t.Fatalf("Lex(%q) emitted 0 tokens", tc.input)
		}
		if ts.Kinds[0] != tc.wantKind {
			t.Errorf("Lex(%q): got %s, want %s", tc.input, ts.Kinds[0], tc.wantKind)
		}
	}
}

func TestBOMHandling(t *testing.T) {
	// BOM at byte 0
	bomInput := append([]byte{0xEF, 0xBB, 0xBF}, []byte("syntax = \"proto3\";")...)
	ts := lexer.NewTokenStream(10, 10)
	if err := lexer.Lex(bomInput, ts); err != nil {
		t.Fatal(err)
	}
	if ts.Len() != 4 {
		t.Fatalf("expected 4 tokens, got %d", ts.Len())
	}
	// First token 'syntax' must start at byte 3
	if ts.Starts[0] != 3 || ts.Lens[0] != 6 {
		t.Errorf("expected token 0 at start 3 len 6, got start %d len %d", ts.Starts[0], ts.Lens[0])
	}
}

func TestNullInComment(t *testing.T) {
	input := []byte("// comment with \x00 null\n/* block with \x00 null */ ident")
	ts := lexer.NewTokenStream(10, 10)
	if err := lexer.Lex(input, ts); err != nil {
		t.Fatal(err)
	}
	if ts.Len() != 3 {
		t.Fatalf("expected 3 tokens, got %d", ts.Len())
	}
	if ts.Kinds[0] != lexer.KindLineComment {
		t.Errorf("token 0: got %s, want line_comment", ts.Kinds[0])
	}
	if ts.Kinds[1] != lexer.KindBlockComment {
		t.Errorf("token 1: got %s, want block_comment", ts.Kinds[1])
	}
	if ts.Kinds[2] != lexer.KindIdent {
		t.Errorf("token 2: got %s, want ident", ts.Kinds[2])
	}
}

func TestGoldenFiles(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata", "lexer")
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("testdata dir not readable: %v", err)
	}

	protoCount := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".proto") {
			continue
		}
		protoCount++
		protoPath := filepath.Join(testdataDir, entry.Name())
		tokensPath := filepath.Join(testdataDir, strings.TrimSuffix(entry.Name(), ".proto")+".tokens")

		protoBytes, err := os.ReadFile(protoPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", protoPath, err)
		}

		tokensBytes, err := os.ReadFile(tokensPath)
		if err != nil {
			t.Fatalf("failed to read %s: %v", tokensPath, err)
		}

		ts := lexer.NewTokenStream(64, 16)
		if err := lexer.Lex(protoBytes, ts); err != nil {
			t.Fatalf("Lex(%s) failed: %v", entry.Name(), err)
		}

		actualDump := ts.FormatDump(entry.Name())
		expectedDump := string(tokensBytes)

		if actualDump != expectedDump {
			t.Errorf("golden mismatch for %s:\nGOT:\n%s\nWANT:\n%s", entry.Name(), actualDump, expectedDump)
		}
	}

	if protoCount != 30 {
		t.Fatalf("expected exactly 30 golden test files, found %d", protoCount)
	}
}

func TestCorpusChecksumDeterministic(t *testing.T) {
	fc1 := uint64(0x1234567890abcdef)
	fc2 := uint64(0xfedcba0987654321)
	c1 := lexer.CorpusChecksum([]uint64{fc1, fc2})
	c2 := lexer.CorpusChecksum([]uint64{fc1, fc2})
	if c1 != c2 {
		t.Errorf("corpus checksum not deterministic: %016x != %016x", c1, c2)
	}
}
