package protoast

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPositionRuneColumns pins the position contract (decision D-07a):
// 1-based lines, 1-based columns counted in runes, and a tab advancing the
// column by exactly one — the output contract of the legacy go-protoparser
// lint path, not protoc's byte columns with tab stops of 8. The
// model-vs-model proof against the legacy parser lives in
// internal/core/position_parity_test.go.
//
// Expected values are hand-computed over
// internal/core/testdata/position_contract_utf8.proto, whose multi-byte
// UTF-8 and tab-indented lines make rune and byte columns provably differ.
func TestPositionRuneColumns(t *testing.T) {
	t.Parallel()

	src, err := os.ReadFile("../core/testdata/position_contract_utf8.proto")
	require.NoError(t, err)

	f, err := Read("position_contract_utf8.proto", bytes.NewReader(src))
	require.NoError(t, err)
	require.Len(t, f.Packages, 1)
	require.Len(t, f.Messages, 1)
	require.Len(t, f.Enums, 1)
	require.Len(t, f.Services, 1)

	message := f.Messages[0]
	enum := f.Enums[0]
	service := f.Services[0]

	tests := []struct {
		name       string
		got        Position
		wantLine   int
		wantColumn int
	}{
		{"package keyword after a multi-byte string literal", f.Packages[0].Pos, 4, 39},
		{"package name after a multi-byte string literal", f.Packages[0].NamePos, 4, 47},
		{"message keyword", message.Pos, 6, 1},
		{"message name after an inline Cyrillic comment", message.NamePos, 6, 22},
		{"field after one leading tab", message.Fields[0].Pos, 7, 2},
		{"field name after one leading tab", message.Fields[0].NamePos, 7, 9},
		{"field after an emoji-bearing comment on the same line", message.Fields[1].Pos, 8, 33},
		{"field name after an emoji-bearing comment on the same line", message.Fields[1].NamePos, 8, 49},
		{"pure-ASCII space-indented field", message.Fields[2].Pos, 9, 3},
		{"pure-ASCII space-indented field name", message.Fields[2].NamePos, 9, 10},
		{"enum keyword", enum.Pos, 12, 1},
		{"enum name after an inline Cyrillic comment", enum.NamePos, 12, 16},
		{"enum value after one leading tab", enum.Values[0].Pos, 13, 2},
		{"enum value after a tab and an emoji-bearing comment", enum.Values[1].Pos, 14, 26},
		{"service keyword", service.Pos, 17, 1},
		{"service name after an inline Cyrillic comment", service.NamePos, 17, 22},
		{"rpc keyword after two leading tabs", service.RPCs[0].Pos, 18, 3},
		{"rpc name after an inline Cyrillic comment", service.RPCs[0].NamePos, 18, 19},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantLine, tt.got.Line, "line must be 1-based and independent of the column convention")
			assert.Equal(t, tt.wantColumn, tt.got.Column, "column must be 1-based, counted in runes, tab = +1")
		})
	}
}
