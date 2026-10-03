package modules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type revisionTestSource struct{ entry v1.LockedModule }

func (s revisionTestSource) Fetch(context.Context, string, string) (Fetched, error) {
	return Fetched{Lock: s.entry}, nil
}

func TestLockedVersionGuard(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "example.com/dep", Version: "v1.0.0", Commit: "old", Hash: "hash-old"}
	tests := []struct {
		name, version, commit, hash string
		reject                      bool
	}{
		{"unchanged", "v1.0.0", "old", "hash-old", false},
		{"commit case", "v1.0.0", "OLD", "hash-old", false},
		{"retagged", "v1.0.0", "new", "hash-new", true},
		{"hash changed", "v1.0.0", "old", "hash-new", true},
		{"new version", "v1.1.0", "new", "hash-new", false},
		{"versionless refresh", "head-commit", "new", "hash-new", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := v1.LockedModule{Source: old.Source, Version: tt.version, Commit: tt.commit, Hash: tt.hash}
			guard := lockedVersionSource{Source: revisionTestSource{entry: entry}, locked: map[string]v1.LockedModule{old.Source: old}}
			got, err := guard.Fetch(t.Context(), old.Source, tt.version)
			if tt.reject {
				require.ErrorIs(t, err, ErrLockedVersionChanged)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, entry, got.Lock)
		})
	}
}
