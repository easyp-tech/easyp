package gitindex_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/adapters/gitindex"
)

func TestParseModes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode string
		want gitindex.Mode
	}{
		{name: "regular_file", mode: "100644", want: gitindex.RegularFile},
		{name: "executable_file", mode: "100755", want: gitindex.ExecutableFile},
		{name: "symlink", mode: "120000", want: gitindex.Symlink},
		{name: "gitlink", mode: "160000", want: gitindex.Gitlink},
		{name: "unknown_mode", mode: "100600", want: "100600"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := tt.mode + " 0123456789abcdef0123456789abcdef01234567 0\tfile.proto"

			entries, err := gitindex.Parse(raw + "\x00")

			require.NoError(t, err)
			assert.Equal(t, []gitindex.Entry{{Mode: tt.want, Path: "file.proto", Raw: raw}}, entries)
		})
	}
}

func TestParsePreservesPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{name: "spaces", path: "directory name/file name.proto"},
		{name: "leading_and_trailing_spaces", path: " file.proto "},
		{name: "tabs", path: "directory\tname/file\tname.proto"},
		{name: "newlines", path: "directory\nname/file\nname.proto"},
		{name: "leading_and_trailing_whitespace", path: "\t\n file.proto \n\t"},
		{name: "non_utf8_bytes", path: "file\xff.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := "100644 0123456789abcdef0123456789abcdef01234567 0\t" + tt.path

			entries, err := gitindex.Parse(raw + "\x00")

			require.NoError(t, err)
			assert.Equal(t, []gitindex.Entry{{Mode: gitindex.RegularFile, Path: tt.path, Raw: raw}}, entries)
		})
	}
}

func TestParseRecords(t *testing.T) {
	t.Parallel()
	first := "100644 0123456789abcdef0123456789abcdef01234567 0\tfirst.proto"
	second := "100755 abcdef0123456789abcdef0123456789abcdef0123 0\tsecond.proto"
	want := []gitindex.Entry{
		{Mode: gitindex.RegularFile, Path: "first.proto", Raw: first},
		{Mode: gitindex.ExecutableFile, Path: "second.proto", Raw: second},
	}
	tests := []struct {
		name string
		raw  string
		want []gitindex.Entry
	}{
		{name: "empty_output", raw: ""},
		{name: "empty_records", raw: "\x00\x00\x00"},
		{name: "multiple_records", raw: first + "\x00" + second + "\x00", want: want},
		{name: "empty_records_are_skipped", raw: "\x00" + first + "\x00\x00" + second + "\x00\x00", want: want},
		{name: "no_trailing_nul", raw: first + "\x00" + second, want: want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			entries, err := gitindex.Parse(tt.raw)

			require.NoError(t, err)
			assert.Equal(t, tt.want, entries)
		})
	}
}

func TestParseMetadata(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		metadata string
	}{
		{name: "opaque_object_id", metadata: "100644 opaque-object-id 0"},
		{name: "sha256_object_id", metadata: "100644 " + strings.Repeat("a", 64) + " 0"},
		{name: "metadata_whitespace", metadata: " 100644  opaque-object-id   0 "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := tt.metadata + "\tfile.proto"

			entries, err := gitindex.Parse(raw + "\x00")

			require.NoError(t, err)
			assert.Equal(t, []gitindex.Entry{{Mode: gitindex.RegularFile, Path: "file.proto", Raw: raw}}, entries)
		})
	}
}

func TestParseInvalidRecords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
	}{
		{name: "missing_separator", raw: "100644 object-id 0 file.proto"},
		{name: "missing_metadata", raw: "\tfile.proto"},
		{name: "missing_field", raw: "100644 0\tfile.proto"},
		{name: "extra_field", raw: "100644 object-id 0 extra\tfile.proto"},
		{name: "unmerged_base_stage", raw: "100644 object-id 1\tfile.proto"},
		{name: "unmerged_ours_stage", raw: "100644 object-id 2\tfile.proto"},
		{name: "unmerged_theirs_stage", raw: "100644 object-id 3\tfile.proto"},
		{name: "invalid_stage", raw: "100644 object-id zero\tfile.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			valid := "100644 object-id 0\tvalid.proto\x00"

			entries, err := gitindex.Parse(valid + tt.raw + "\x00")

			require.EqualError(t, err, fmt.Sprintf("invalid Git index entry %q", tt.raw))
			assert.Nil(t, entries)
		})
	}
}

func TestParseNonlocalPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
	}{
		{name: "empty_path", path: ""},
		{name: "absolute_path", path: "/file.proto"},
		{name: "parent_directory", path: ".."},
		{name: "parent_relative_path", path: "../file.proto"},
		{name: "nested_parent_escape", path: "directory/../../file.proto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			valid := "100644 object-id 0\tvalid.proto\x00"
			raw := "100644 object-id 0\t" + tt.path

			entries, err := gitindex.Parse(valid + raw + "\x00")

			require.EqualError(t, err, fmt.Sprintf("invalid tracked file path %q", tt.path))
			assert.Nil(t, entries)
		})
	}
}

func TestEntryIsRegular(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode gitindex.Mode
		want bool
	}{
		{name: "regular_file", mode: "100644", want: true},
		{name: "executable_file", mode: "100755", want: true},
		{name: "symlink", mode: "120000"},
		{name: "gitlink", mode: "160000"},
		{name: "unknown_mode", mode: "100600"},
		{name: "unset_mode"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := gitindex.Entry{Mode: tt.mode}

			assert.Equal(t, tt.want, entry.IsRegular())
		})
	}
}
