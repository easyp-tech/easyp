package gitmodules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestTrackedV1Files(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		regularName   string
		linkName      string
		materialize   bool
		replaceByLink bool
		wantErr       string
	}{
		{name: "regular_files_and_gitlinks", regularName: "file.proto"},
		{name: "spaces_tabs_and_newlines_in_names", regularName: "directory name/file\twith\nnewline.proto"},
		{name: "materialized_git_symlink", regularName: "file.proto", materialize: true},
		{name: "materialized_config_symlink", regularName: "file.proto", linkName: "buf.yaml", materialize: true, wantErr: "non-regular dependency config"},
		{name: "materialized_buf_lock_symlink", regularName: "file.proto", linkName: "buf.lock", materialize: true, wantErr: "non-regular dependency config"},
		{name: "regular_git_file_replaced_by_symlink", regularName: "file.proto", replaceByLink: true, wantErr: "non-regular"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repository := t.TempDir()
			file := filepath.Join(repository, tt.regularName)
			require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
			require.NoError(t, os.WriteFile(file, []byte("syntax = \"proto3\";\n"), 0o644))
			linkName := tt.linkName
			if linkName == "" {
				linkName = "ignored"
			}
			link := filepath.Join(repository, linkName)
			require.NoError(t, os.Symlink(tt.regularName, link))
			runTestGit(t, repository, "init", "-q")
			runTestGit(t, repository, "add", ".")
			runTestGit(t, repository, "-c", "user.name=Test", "-c", "user.email=test@example.test", "commit", "-qm", "initial")
			commit := runTestGit(t, repository, "rev-parse", "HEAD")
			runTestGit(t, repository, "update-index", "--add", "--cacheinfo", "160000", commit[:len(commit)-1], "submodule")
			if tt.materialize {
				require.NoError(t, os.Remove(link))
				require.NoError(t, os.WriteFile(link, []byte(tt.regularName), 0o644))
			}
			if tt.replaceByLink {
				require.NoError(t, os.Remove(file))
				require.NoError(t, os.Symlink("ignored", file))
			}

			files, err := trackedV1Files(t.Context(), repository)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{tt.regularName}, files)
		})
	}
}

func TestSelectV1ProtoFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		filters []v1.ProtoFileFilter
		want    []string
	}{
		{
			name: "no_filters",
			want: []string{"proto/api.proto", "proto/api/public.proto", "proto/api/test/hidden.proto", "proto/apiv2/other.proto", "other/keep.proto", "proto/api/test/README"},
		},
		{
			name:    "includes_keep_paths_and_non_proto_files",
			filters: []v1.ProtoFileFilter{{Root: "proto", Includes: []string{"proto/api"}, Excludes: []string{"proto/api/test"}}},
			want:    []string{"proto/api/public.proto", "other/keep.proto", "proto/api/test/README"},
		},
		{
			name:    "exclude_uses_directory_boundary",
			filters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/api"}}},
			want:    []string{"proto/api.proto", "proto/apiv2/other.proto", "other/keep.proto", "proto/api/test/README"},
		},
		{
			name: "multiple_modules_share_a_root",
			filters: []v1.ProtoFileFilter{
				{Root: "proto", Includes: []string{"proto/api"}, Excludes: []string{"proto/api/test"}},
				{Root: "proto", Includes: []string{"proto/apiv2"}},
			},
			want: []string{"proto/api/public.proto", "proto/apiv2/other.proto", "other/keep.proto", "proto/api/test/README"},
		},
		{
			name:    "unfiltered_overlapping_root_remains_available",
			filters: []v1.ProtoFileFilter{{Root: "proto", Excludes: []string{"proto/api"}}, {Root: "proto/api"}},
			want:    []string{"proto/api.proto", "proto/api/public.proto", "proto/api/test/hidden.proto", "proto/apiv2/other.proto", "other/keep.proto", "proto/api/test/README"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files := []string{"proto/api.proto", "proto/api/public.proto", "proto/api/test/hidden.proto", "proto/apiv2/other.proto", "other/keep.proto", "proto/api/test/README"}
			original := append([]string(nil), files...)

			selected := selectV1ProtoFiles(files, tt.filters)

			assert.Equal(t, tt.want, selected)
			assert.Equal(t, original, files, "source selection must not change the tracked file list")
		})
	}
}
