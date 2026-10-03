package v1

import (
	"strings"
	"testing"
)

func TestLockRejectsCommitVersionThatDiffersFromPinnedCommit(t *testing.T) {
	commit := strings.Repeat("a", 40)
	lock := Lock{Version: 1, Modules: []LockedModule{{
		Source: "example.com/foo", Version: commit, Commit: commit,
		Hash: "h1:" + strings.Repeat("A", 43) + "=",
	}}}
	if err := lock.Validate(); err != nil {
		t.Fatal(err)
	}
	lock.Modules[0].Commit = strings.Repeat("b", 40)
	if err := lock.Validate(); err == nil || !strings.Contains(err.Error(), "differs from commit") {
		t.Fatalf("expected mismatched pinned commit error, got %v", err)
	}
}
