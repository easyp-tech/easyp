package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModuleEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want Module
	}{
		{
			name: "initial_utf8_bom",
			raw:  "\ufeffmodule example.com/app\n",
			want: Module{Name: "example.com/app", Roots: []string{"."}},
		},
		{
			name: "bom_comments_and_crlf",
			raw:  "\ufeff# heading\r\n\r\nmodule example.com/app # identity\r\nroots proto\r\n",
			want: Module{Name: "example.com/app", Roots: []string{"proto"}},
		},
		{
			name: "hash_comments_at_token_start",
			raw: "# heading\nmodule example.com/app # identity\nroots ( # roots\n" +
				" proto # directory\n) # end\nrequire ( # dependencies\n" +
				" example.com/x # versionless\n example.com/y v1.2.0\t# pinned\n)\n" +
				"replace ( # local\n example.com/x => ../x # target\n) # end\n#",
			want: Module{
				Name:     "example.com/app",
				Roots:    []string{"proto"},
				Requires: []Requirement{{Module: "example.com/x"}, {Module: "example.com/y", Version: "v1.2.0"}},
				Replaces: []Replacement{{Module: "example.com/x", Target: "../x"}},
			},
		},
		{
			name: "unicode_whitespace_before_comments",
			raw:  "\u00a0# heading\nmodule example.com/app\u2003// identity\nrequire example.com/x\u00a0# dependency\n",
			want: Module{Name: "example.com/app", Roots: []string{"."}, Requires: []Requirement{{Module: "example.com/x"}}},
		},
		{
			name: "hash_and_slashes_inside_url_tokens",
			raw: "module https://example.com/app.git#identity\n" +
				"require https://example.com/x.git?path=a//b#fragment v1.2.0 # comment\n" +
				"require ssh://git@example.com:2222/y.git // comment\n",
			want: Module{
				Name:  "https://example.com/app.git#identity",
				Roots: []string{"."},
				Requires: []Requirement{
					{Module: "https://example.com/x.git?path=a//b#fragment", Version: "v1.2.0"},
					{Module: "ssh://git@example.com:2222/y.git"},
				},
			},
		},
		{
			name: "url_parentheses_are_not_blocks",
			raw: "module https://example.com/app(\n" +
				"require https://example.com/x)\n" +
				"replace https://example.com/x) => ./local(\n",
			want: Module{
				Name:     "https://example.com/app(",
				Roots:    []string{"."},
				Requires: []Requirement{{Module: "https://example.com/x)"}},
				Replaces: []Replacement{{Module: "https://example.com/x)", Target: "./local("}},
			},
		},
		{
			name: "versionless_sha_and_major_versions_are_unchanged",
			raw: "module example.com/app\nrequire example.com/unversioned\nrequire(\n" +
				" example.com/sha1 0123456789abcdef0123456789abcdef01234567\n" +
				" example.com/sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef\n" +
				" example.com/v2 v1.2.0\n example.com/plain v2.3.4\n)\n",
			want: Module{
				Name:  "example.com/app",
				Roots: []string{"."},
				Requires: []Requirement{
					{Module: "example.com/unversioned"},
					{Module: "example.com/sha1", Version: "0123456789abcdef0123456789abcdef01234567"},
					{Module: "example.com/sha256", Version: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
					{Module: "example.com/v2", Version: "v1.2.0"},
					{Module: "example.com/plain", Version: "v2.3.4"},
				},
			},
		},
		{
			name: "relative_and_absolute_replacement_targets",
			raw: "module example.com/app\nreplace example.com/x => ../x\nreplace(\n" +
				" example.com/y => /opt/protos/y\n example.com/z => local(z)#copy\n)\n",
			want: Module{
				Name:  "example.com/app",
				Roots: []string{"."},
				Replaces: []Replacement{
					{Module: "example.com/x", Target: "../x"},
					{Module: "example.com/y", Target: "/opt/protos/y"},
					{Module: "example.com/z", Target: "local(z)#copy"},
				},
			},
		},
		{
			name: "empty_blocks_and_compact_openers",
			raw:  "module example.com/app\nroots(\n)\nrequire (\n)\nreplace(\n)\n",
			want: Module{Name: "example.com/app", Roots: []string{"."}},
		},
		{
			name: "sources_are_distinct_from_versions_and_targets",
			raw: "module example.com/app\nrequire example.com/x v1.2.0\nrequire example.com/y v1.2.0\n" +
				"replace example.com/x => ./shared\nreplace example.com/y => ./shared\n",
			want: Module{
				Name:     "example.com/app",
				Roots:    []string{"."},
				Requires: []Requirement{{Module: "example.com/x", Version: "v1.2.0"}, {Module: "example.com/y", Version: "v1.2.0"}},
				Replaces: []Replacement{{Module: "example.com/x", Target: "./shared"}, {Module: "example.com/y", Target: "./shared"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseModule(strings.NewReader(tt.raw))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsModuleManifest([]byte(tt.raw)))
		})
	}
}

func TestParseModuleRejectsMalformedDirectives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "same_line_require_block", raw: "require (example.com/x v1.2.0)"},
		{name: "spaced_same_line_require_block", raw: "require ( example.com/x v1.2.0 )"},
		{name: "versionless_same_line_require_block", raw: "require (example.com/x)"},
		{name: "compact_same_line_require_block", raw: "require(example.com/x v1.2.0)"},
		{name: "same_line_roots_block", raw: "roots (proto)"},
		{name: "same_line_replace_block", raw: "replace (example.com/x => ./local)"},
		{name: "module_parenthesis", raw: "module (example.com/app)"},
		{name: "module_arrow_token", raw: "module =>"},
		{name: "module_quoted_identity", raw: "module \"example.com/app\""},
		{name: "missing_module", raw: "module"},
		{name: "extra_module_token", raw: "module example.com/app extra"},
		{name: "missing_roots", raw: "roots # missing path"},
		{name: "missing_requirement", raw: "require # missing source"},
		{name: "extra_requirement_token", raw: "require example.com/x v1.2.0 extra"},
		{name: "require_arrow_source", raw: "require => v1.2.0"},
		{name: "require_arrow_version", raw: "require example.com/x =>"},
		{name: "require_closing_parenthesis", raw: "require example.com/x )"},
		{name: "require_closing_version", raw: "require example.com/x v1.2.0)"},
		{name: "missing_replace_target", raw: "replace example.com/x =>"},
		{name: "malformed_replace_arrow", raw: "replace example.com/x -> ./local"},
		{name: "replace_arrow_source", raw: "replace => => ./local"},
		{name: "replace_arrow_target", raw: "replace example.com/x => =>"},
		{name: "replace_parenthesis_target", raw: "replace example.com/x => )"},
		{name: "extra_replace_token", raw: "replace example.com/x => ./local extra"},
		{name: "roots_arrow_token", raw: "roots =>"},
		{name: "roots_parenthesis_token", raw: "roots proto )"},
		{name: "block_opener_extra_token", raw: "require extra ("},
		{name: "unknown_block", raw: "unknown ("},
		{name: "unknown_directive", raw: "unknown example.com/x"},
		{name: "embedded_bom", raw: "require \ufeffexample.com/x"},
		{name: "nul_in_requirement", raw: "require example.com/\x00x"},
		{name: "quoted_requirement", raw: "require \"example.com/x\""},
		{name: "backtick_requirement", raw: "require `example.com/x`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prefix := "module example.com/app\n"
			if strings.Fields(tt.raw)[0] == "module" {
				prefix = "\n"
			}
			got, err := ParseModule(strings.NewReader(prefix + tt.raw + "\n"))
			require.ErrorContains(t, err, "protobuf.mod:2:")
			assert.Equal(t, Module{}, got)
			assert.Equal(t, 1, strings.Count(err.Error(), "protobuf.mod:"))
		})
	}
}

func TestParseModuleRejectsMalformedIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "arrow", raw: "module =>"},
		{name: "closing_parenthesis", raw: "module )"},
		{name: "quoted_identity", raw: "module \"example.com/app\""},
		{name: "repeated_bom", raw: "\ufeff\ufeffmodule example.com/app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseModule(strings.NewReader(tt.raw))
			require.ErrorContains(t, err, "protobuf.mod:1:")
			assert.Equal(t, Module{}, got)
			assert.Equal(t, 1, strings.Count(err.Error(), "protobuf.mod:"))
		})
	}
}

func TestParseModuleBlockLocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		wantErr []string
	}{
		{
			name:    "unmatched_closing_parenthesis",
			raw:     "module example.com/app\n\n)\n",
			wantErr: []string{"protobuf.mod:3:", "unmatched closing parenthesis"},
		},
		{
			name:    "unclosed_require_block",
			raw:     "module example.com/app\nrequire (\n example.com/x\n",
			wantErr: []string{"protobuf.mod:2:", "unclosed require block"},
		},
		{
			name:    "unclosed_roots_block_after_comments",
			raw:     "\ufeff# heading\nmodule example.com/app\n\nroots( # paths\n proto\n",
			wantErr: []string{"protobuf.mod:4:", "unclosed roots block"},
		},
		{
			name:    "unclosed_replace_block",
			raw:     "module example.com/app\nreplace (\n example.com/x => ./local\n",
			wantErr: []string{"protobuf.mod:2:", "unclosed replace block"},
		},
		{
			name:    "nested_named_block",
			raw:     "module example.com/app\nrequire (\n\nreplace (\n)\n)\n",
			wantErr: []string{"protobuf.mod:4:", "nested block", "opened at line 2"},
		},
		{
			name:    "nested_anonymous_block",
			raw:     "module example.com/app\nroots (\n (\n )\n)\n",
			wantErr: []string{"protobuf.mod:3:", "nested block", "opened at line 2"},
		},
		{
			name:    "closing_parenthesis_with_extra_tokens",
			raw:     "module example.com/app\nroots (\n ) extra\n",
			wantErr: []string{"protobuf.mod:3:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseModule(strings.NewReader(tt.raw))
			require.Error(t, err)
			for _, message := range tt.wantErr {
				assert.Contains(t, err.Error(), message)
			}
			assert.Equal(t, Module{}, got)
			assert.Equal(t, 1, strings.Count(err.Error(), "protobuf.mod:"))
		})
	}
}

func TestParseModuleDuplicateSourceLocations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		raw       string
		directive string
		firstLine string
		lastLine  string
	}{
		{
			name: "require_repeated_inline", directive: "require",
			raw:       "require example.com/x v1.2.0\nrequire example.com/x v1.2.0\n",
			firstLine: "first declared at line 2", lastLine: "protobuf.mod:3:",
		},
		{
			name: "require_conflicting_versions_in_block", directive: "require",
			raw:       "require (\n example.com/x v1.2.0\n example.com/x v2.0.0\n)\n",
			firstLine: "first declared at line 3", lastLine: "protobuf.mod:4:",
		},
		{
			name: "require_versionless_then_block", directive: "require",
			raw:       "require example.com/x\nrequire (\n example.com/x v1.2.0\n)\n",
			firstLine: "first declared at line 2", lastLine: "protobuf.mod:4:",
		},
		{
			name: "require_across_blocks", directive: "require",
			raw:       "require (\n example.com/x\n)\nrequire (\n example.com/x\n)\n",
			firstLine: "first declared at line 3", lastLine: "protobuf.mod:6:",
		},
		{
			name: "replace_repeated_inline", directive: "replace",
			raw:       "replace example.com/x => ./local\nreplace example.com/x => ./local\n",
			firstLine: "first declared at line 2", lastLine: "protobuf.mod:3:",
		},
		{
			name: "replace_conflicting_targets_in_block", directive: "replace",
			raw:       "replace (\n example.com/x => ./first\n example.com/x => ./second\n)\n",
			firstLine: "first declared at line 3", lastLine: "protobuf.mod:4:",
		},
		{
			name: "replace_block_then_inline", directive: "replace",
			raw:       "replace (\n example.com/x => ./first\n)\nreplace example.com/x => /opt/second\n",
			firstLine: "first declared at line 3", lastLine: "protobuf.mod:5:",
		},
		{
			name: "replace_across_blocks", directive: "replace",
			raw:       "replace (\n example.com/x => ./local\n)\nreplace (\n example.com/x => ./local\n)\n",
			firstLine: "first declared at line 3", lastLine: "protobuf.mod:6:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseModule(strings.NewReader("module example.com/app\n" + tt.raw))
			require.ErrorContains(t, err, "duplicate "+tt.directive)
			assert.Contains(t, err.Error(), "example.com/x")
			assert.Contains(t, err.Error(), tt.firstLine)
			assert.Contains(t, err.Error(), tt.lastLine)
			assert.Equal(t, Module{}, got)
			assert.Equal(t, 1, strings.Count(err.Error(), "protobuf.mod:"))
		})
	}
}

func TestIsModuleManifestEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{
			name: "bom_and_hash_comments_preserve_v1_parse_errors",
			raw:  "\ufeff# heading\nmodule example.com/app\nunknown directive\n",
			want: true,
		},
		{
			name: "valid_block_before_module_with_bom",
			raw:  "\ufeffrequire ( # dependencies\n example.com/x\n)\nmodule example.com/app\n",
			want: true,
		},
		{
			name: "legacy_with_bom_and_hash_comment",
			raw:  "\ufeff# heading\ndirect (\n example.com/x@v1.2.0\n)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, IsModuleManifest([]byte(tt.raw)))
		})
	}
}
