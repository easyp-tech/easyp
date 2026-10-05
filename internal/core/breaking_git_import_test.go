package core_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	gitadapter "github.com/easyp-tech/easyp/internal/adapters/go_git"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingCheckResolvesWellKnownImportFromGitBaseline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		importName string
		fieldType  string
	}{
		{name: "timestamp", importName: "google/protobuf/timestamp.proto", fieldType: "google.protobuf.Timestamp"},
		{name: "duration", importName: "google/protobuf/duration.proto", fieldType: "google.protobuf.Duration"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			path := filepath.Join(root, "proto", "event.proto")
			require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
			contents := "syntax = \"proto3\";\npackage event.v1;\nimport \"" + tt.importName + "\";\nmessage Event { " + tt.fieldType + " created_at = 1; }\n"
			require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
			repository, err := git.PlainInit(root, false)
			require.NoError(t, err)
			worktree, err := repository.Worktree()
			require.NoError(t, err)
			_, err = worktree.Add("proto/event.proto")
			require.NoError(t, err)
			commit, err := worktree.Commit("baseline", &git.CommitOptions{Author: &object.Signature{Name: "EasyP Test", Email: "test@example.com", When: time.Now()}})
			require.NoError(t, err)
			require.NoError(t, repository.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), commit)))
			checker := core.New(core.Options{
				Logger:                  logger.NewNop(),
				CurrentProjectGitWalker: gitadapter.New(),
				BreakingCheckConfig:     core.BreakingCheckConfig{AgainstGitRef: "main"},
			})

			issues, err := checker.BreakingCheck(t.Context(), root, filepath.Join(root, "proto"), ".")

			require.NoError(t, err)
			require.Empty(t, issues)
		})
	}
}
