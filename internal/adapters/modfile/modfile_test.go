package modfile

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/core/models"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    File
		wantErr error
	}{
		{
			name:  "empty_file",
			input: "",
			want:  File{},
		},
		{
			name:  "comments_only",
			input: "# nothing\n// still nothing\n",
			want:  File{},
		},
		{
			name: "basic_direct_block",
			input: `direct (
	github.com/googleapis/googleapis@common-protos-1_3_1
	github.com/bufbuild/protoc-gen-validate
)
`,
			want: File{
				Direct: []string{
					"github.com/googleapis/googleapis@common-protos-1_3_1",
					"github.com/bufbuild/protoc-gen-validate",
				},
			},
		},
		{
			name: "compact_paren",
			input: `direct(
	github.com/foo/bar@v1.0.0
)
`,
			want: File{Direct: []string{"github.com/foo/bar@v1.0.0"}},
		},
		{
			name: "inline_comments",
			input: `direct (
	github.com/foo/bar@v1.0.0 # pin
	github.com/baz/qux // latest
)
`,
			want: File{
				Direct: []string{
					"github.com/foo/bar@v1.0.0",
					"github.com/baz/qux",
				},
			},
		},
		{
			name: "empty_direct_block",
			input: `direct (
)
`,
			want: File{},
		},
		{
			name: "garbage_outside",
			input: `module example.com/foo
direct (
	github.com/foo/bar@v1
)
`,
			wantErr: errGarbageOutside,
		},
		{
			name: "unclosed_block",
			input: `direct (
	github.com/foo/bar@v1
`,
			wantErr: errUnclosedDirect,
		},
		{
			name: "duplicate_module",
			input: `direct (
	github.com/foo/bar@v1
	github.com/foo/bar@v2
)
`,
			wantErr: errDuplicateModule,
		},
		{
			name: "multiple_fields_on_line",
			input: `direct (
	github.com/foo/bar@v1 extra
)
`,
			wantErr: errUnexpectedToken,
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
			wantErr: errMultipleDirect,
		},
		{
			name: "single_line_replace",
			input: `direct (
	github.com/acme/weather@v1.2
)
replace github.com/acme/weather@v1.2 => /home/project
`,
			want: File{
				Direct: []string{"github.com/acme/weather@v1.2"},
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/project",
					},
				},
			},
		},
		{
			name: "replace_block",
			input: `direct (
	github.com/acme/weather@v1.2
)
replace (
	github.com/acme/weather@v1.2 => ../weather
)
`,
			want: File{
				Direct: []string{"github.com/acme/weather@v1.2"},
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "../weather",
					},
				},
			},
		},
		{
			name: "replace_without_direct",
			input: `replace github.com/acme/weather@v1.2 => /home/project
`,
			want: File{
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/project",
					},
				},
			},
		},
		{
			name: "replace_before_direct",
			input: `replace github.com/acme/weather@v1.2 => /home/project
direct (
	github.com/acme/weather@v1.2
)
`,
			want: File{
				Direct: []string{"github.com/acme/weather@v1.2"},
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/project",
					},
				},
			},
		},
		{
			name: "replace_inline_comment",
			input: `replace github.com/acme/weather@v1.2 => /home/project # local checkout
`,
			want: File{
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/project",
					},
				},
			},
		},
		{
			name: "replace_compact_arrow",
			input: `replace github.com/acme/weather@v1.2=>/home/project
`,
			want: File{
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/project",
					},
				},
			},
		},
		{
			name: "replace_same_module_different_versions",
			input: `replace (
	github.com/acme/weather@v1.2 => /home/weather-v1
	github.com/acme/weather@v2.0 => /home/weather-v2
)
`,
			want: File{
				Replace: []Replace{
					{
						Module: models.NewModule("github.com/acme/weather@v1.2"),
						Path:   "/home/weather-v1",
					},
					{
						Module: models.NewModule("github.com/acme/weather@v2.0"),
						Path:   "/home/weather-v2",
					},
				},
			},
		},
		{
			name: "duplicate_replace",
			input: `replace github.com/acme/weather@v1.2 => /a
replace github.com/acme/weather@v1.2 => /b
`,
			wantErr: errDuplicateReplace,
		},
		{
			name: "replace_version_omitted",
			input: `replace github.com/acme/weather => /home/project
`,
			wantErr: errReplaceVersionRequired,
		},
		{
			name: "replace_empty_path",
			input: `replace github.com/acme/weather@v1.2 =>
`,
			wantErr: errEmptyReplacePath,
		},
		{
			name: "unclosed_replace_block",
			input: `replace (
	github.com/acme/weather@v1.2 => /home/project
`,
			wantErr: errUnclosedReplace,
		},
		{
			name: "replace_missing_arrow",
			input: `replace github.com/acme/weather@v1.2 /home/project
`,
			wantErr: errUnexpectedToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse([]byte(tt.input))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormat_RoundTrip(t *testing.T) {
	t.Parallel()

	file := File{
		Direct: []string{
			"github.com/googleapis/googleapis@common-protos-1_3_1",
			"github.com/bufbuild/protoc-gen-validate",
		},
	}

	got, err := Parse(Format(file))
	require.NoError(t, err)
	require.Equal(t, file, got)
}

func TestFormat_ReplaceRoundTrip(t *testing.T) {
	t.Parallel()

	file := File{
		Direct: []string{"github.com/acme/weather@v1.2"},
		Replace: []Replace{
			{
				Module: models.NewModule("github.com/acme/weather@v1.2"),
				Path:   "/home/project",
			},
		},
	}

	got, err := Parse(Format(file))
	require.NoError(t, err)
	require.Equal(t, file, got)
}
