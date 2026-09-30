package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	disk "github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestBreakingProfilesParentCatalogDescribesActualFindings(t *testing.T) {
	t.Parallel()
	for _, profile := range config.BreakingProfiles() {
		t.Run(profile.Name, func(t *testing.T) {
			t.Parallel()
			before, after := t.TempDir(), t.TempDir()
			text := "syntax = \"proto3\"; package sample.v1; message Item { oneof choice { string id = 1; } } enum State { ZERO = 0; } service API { rpc Get(Item) returns(Item); }"
			require.NoError(t, os.WriteFile(filepath.Join(before, "old.proto"), []byte(text), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(after, "new.proto"), []byte(text), 0o600))
			app := New(Options{Logger: logger.NewNop(), BreakingCheckConfig: BreakingCheckConfig{Categories: []string{profile.Name}}})
			findings, err := app.CompareBreaking(t.Context(), disk.NewFSWalker(after, "."), disk.NewFSWalker(before, "."), nil)
			require.NoError(t, err)
			for _, issue := range findings {
				assert.Contains(t, profile.Rules, issue.RuleName, "emitted rule missing from profile catalog")
			}
			require.NoError(t, os.WriteFile(filepath.Join(after, "new.proto"), []byte("syntax = \"proto3\"; package sample.v1; message Item {}"), 0o600))
			findings, err = app.CompareBreaking(t.Context(), disk.NewFSWalker(after, "."), disk.NewFSWalker(before, "."), nil)
			require.NoError(t, err)
			require.NotEmpty(t, findings)
			for _, issue := range findings {
				assert.Contains(t, profile.Rules, issue.RuleName, "emitted rule missing from profile catalog")
			}
		})
	}
}

func TestBreakingProfilesParentCatalogOwnsReturnedSlices(t *testing.T) {
	t.Parallel()
	before := config.BreakingProfiles()
	modified := config.BreakingProfiles()
	for i := range modified {
		modified[i].Name = "modified"
		for j := range modified[i].Rules {
			modified[i].Rules[j] = "modified"
		}
	}
	assert.Equal(t, before, config.BreakingProfiles())
}

func TestBreakingProfilesParentMovedFileOptions(t *testing.T) {
	t.Parallel()
	for _, profile := range []string{"FILE", "PACKAGE", "WIRE_JSON", "WIRE"} {
		t.Run(profile, func(t *testing.T) {
			t.Parallel()
			before, after := t.TempDir(), t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(before, "old.proto"), []byte("syntax = \"proto3\"; package api.v1; option go_package = \"example.test/before\"; message Item {}"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(after, "new.proto"), []byte("syntax = \"proto3\"; package api.v1; option go_package = \"example.test/after\"; message Item {}"), 0o600))
			app := New(Options{Logger: logger.NewNop(), BreakingCheckConfig: BreakingCheckConfig{Categories: []string{profile}}})
			findings, err := app.CompareBreaking(t.Context(), disk.NewFSWalker(after, "."), disk.NewFSWalker(before, "."), nil)
			require.NoError(t, err)
			found := false
			for _, issue := range findings {
				if issue.RuleName == "FILE_SAME_GO_PACKAGE" {
					found = true
				}
			}
			assert.Equal(t, profile == "FILE" || profile == "PACKAGE", found, "moving declarations must not conceal changed generated Go package")
		})
	}
}
