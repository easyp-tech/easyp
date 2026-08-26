package modfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "empty_file",
			input: "",
			want:  nil,
		},
		{
			name:  "comments_only",
			input: "# nothing\n// still nothing\n",
			want:  nil,
		},
		{
			name: "basic_direct_block",
			input: `direct (
	github.com/googleapis/googleapis@common-protos-1_3_1
	github.com/bufbuild/protoc-gen-validate
)
`,
			want: []string{
				"github.com/googleapis/googleapis@common-protos-1_3_1",
				"github.com/bufbuild/protoc-gen-validate",
			},
		},
		{
			name: "compact_paren",
			input: `direct(
	github.com/foo/bar@v1.0.0
)
`,
			want: []string{"github.com/foo/bar@v1.0.0"},
		},
		{
			name: "inline_comments",
			input: `direct (
	github.com/foo/bar@v1.0.0 # pin
	github.com/baz/qux // latest
)
`,
			want: []string{
				"github.com/foo/bar@v1.0.0",
				"github.com/baz/qux",
			},
		},
		{
			name: "empty_direct_block",
			input: `direct (
)
`,
			want: nil,
		},
		{
			name: "garbage_outside",
			input: `module example.com/foo
direct (
	github.com/foo/bar@v1
)
`,
			wantErr: true,
		},
		{
			name: "unclosed_block",
			input: `direct (
	github.com/foo/bar@v1
`,
			wantErr: true,
		},
		{
			name: "duplicate_module",
			input: `direct (
	github.com/foo/bar@v1
	github.com/foo/bar@v2
)
`,
			wantErr: true,
		},
		{
			name: "multiple_fields_on_line",
			input: `direct (
	github.com/foo/bar@v1 extra
)
`,
			wantErr: true,
		},
		{
			name: "multiple_direct_blocks",
			input: `direct (
	github.com/foo/bar@v1
)
direct (
	github.com/baz/qux@v1
)
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse([]byte(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_RoundTrip(t *testing.T) {
	t.Parallel()

	deps := []string{
		"github.com/googleapis/googleapis@common-protos-1_3_1",
		"github.com/bufbuild/protoc-gen-validate",
	}

	got, err := Parse(Format(deps))
	require.NoError(t, err)
	require.Equal(t, deps, got)
}
