package v1

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestGeneratePathSelectorValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		valid bool
	}{
		{name: ".", valid: true}, {name: "mcp", valid: true},
		{name: "mcp/options/v1/options.proto", valid: true}, {name: ".sources/api.proto", valid: true},
		{name: "api.proto/messages.proto", valid: true}, {name: "api-v1/sub_dir", valid: true},
		{name: "..sources/api.proto", valid: true}, {name: ".../api.proto", valid: true},
		{name: ""}, {name: "./mcp"}, {name: "../mcp"}, {name: "mcp/../other"},
		{name: "mcp/./options"}, {name: "mcp/"}, {name: "mcp//options"}, {name: "/mcp"},
		{name: "C:/mcp"}, {name: `mcp\options`}, {name: "mcp/*"}, {name: "mcp/**"},
		{name: "mcp[ab]"}, {name: "mcp?"}, {name: "mcp options"}, {name: "mcp\noptions"},
		{name: "mcp\toptions"}, {name: "mcp\u00a0options"}, {name: "mcp\u2003options"},
		{name: "mcp\u2028options"}, {name: "mcp\x00options"}, {name: "mcp|options"},
		{name: "mcp\n"}, {name: "mcp\r"}, {name: "mcp\u2028"}, {name: "mcp\ufeff"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := "version: v1\ngenerate:\n  paths: [" + strconv.Quote(tt.name) + "]\nplugins: []\n"
			_, err := ParseGenerate(strings.NewReader(raw))
			assert.Equal(t, !tt.valid, config.HasErrors(ValidateGenerateYAML([]byte(raw))))
			if tt.valid {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, "generate.paths")
		})
	}
	require.NoError(t, ValidatePathSelectors(nil))
	require.NoError(t, ValidatePathSelectors([]string{"mcp", "mcp"}))
}

func TestPathSelectorMatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		selector string
		file     string
		want     bool
	}{
		{name: "all", selector: ".", file: "mcp/options.proto", want: true},
		{name: "exact_file", selector: "mcp/options.proto", file: "mcp/options.proto", want: true},
		{name: "subtree", selector: "mcp", file: "mcp/options/v1/options.proto", want: true},
		{name: "directory_with_proto_suffix", selector: "api.proto", file: "api.proto/messages.proto", want: true},
		{name: "prefix_lookalike", selector: "mcp", file: "mcp-copy/options.proto"},
		{name: "file_prefix_lookalike", selector: "mcp/options.proto", file: "mcp/options.proto-copy"},
		{name: "different_root", selector: "mcp", file: "examples/mcp/options.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, PathSelectorMatches(tt.selector, tt.file))
		})
	}
}
