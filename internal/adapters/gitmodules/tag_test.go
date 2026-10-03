package gitmodules

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		tag       string
		subdir    string
		annotated bool
	}{
		{name: "lightweight_tag", tag: "common-protos-1_3_1"},
		{name: "annotated_tag", tag: "common-protos-1_3_1", annotated: true},
		{name: "tag_with_slash", tag: "schemas/stable"},
		{name: "nested_module_tag", tag: "schemas-release", subdir: "foo"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote, commit := diagnosticRepository(t)
			tag := tt.tag
			if tt.subdir != "" {
				tag = tt.subdir + "/" + tag
			}
			if tt.annotated {
				runTestGit(t, remote, "-c", "user.name=Test", "-c", "user.email=test@example.test", "tag", "-am", "release", tag)
			} else {
				runTestGit(t, remote, "tag", tag)
			}

			resolved, err := New(t.TempDir()).ResolveTag(t.Context(), filepath.Join(remote, tt.subdir), tt.tag)

			require.NoError(t, err)
			assert.Equal(t, commit, resolved)
		})
	}
}

func TestResolveTagErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tag     string
		offline bool
		cancel  bool
		wantErr string
	}{
		{name: "missing_tag", tag: "absent", wantErr: "was not found"},
		{name: "branch_is_not_a_tag", tag: "HEAD", wantErr: "was not found"},
		{name: "repository_access", tag: "v1.0.0", offline: true, wantErr: "could not query Git tags"},
		{name: "invalid_tag", tag: "../unsafe", wantErr: "check-ref-format"},
		{name: "cancelled_context", tag: "v1.0.0", cancel: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			remote, _ := diagnosticRepository(t)
			if tt.offline {
				require.NoError(t, os.Rename(remote, remote+"-offline"))
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if tt.cancel {
				cancel()
			}

			commit, err := New(t.TempDir()).ResolveTag(ctx, remote, tt.tag)

			assert.Empty(t, commit)
			if tt.cancel {
				require.ErrorIs(t, err, context.Canceled)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
			if tt.offline {
				assert.False(t, strings.Contains(err.Error(), "was not found"))
			}
		})
	}
}
