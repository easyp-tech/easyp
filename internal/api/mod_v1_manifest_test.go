package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestWriteV1ResolvedFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		updated       string
		lockDirectory bool
		wantManifest  string
		wantErr       bool
	}{
		{
			name:         "write manifest and lock",
			updated:      "module example.com/root\nrequire example.com/dep\n",
			wantManifest: "module example.com/root\nrequire example.com/dep\n",
		},
		{
			name:          "restore changed manifest when lock cannot replace",
			updated:       "module example.com/root\nrequire example.com/dep\n",
			lockDirectory: true,
			wantManifest:  "module example.com/root\n",
			wantErr:       true,
		},
		{
			name:          "preserve unchanged manifest when lock cannot replace",
			updated:       "module example.com/root\n",
			lockDirectory: true,
			wantManifest:  "module example.com/root\n",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			original := []byte("module example.com/root\n")
			require.NoError(t, os.WriteFile(filepath.Join(root, v1.ModuleFile), original, 0o600))
			if tt.lockDirectory {
				require.NoError(t, os.Mkdir(filepath.Join(root, v1.LockFile), 0o755))
			}

			err := writeV1ResolvedFiles(root, original, []byte(tt.updated), v1.Lock{Version: 1})

			if tt.wantErr {
				var pathErr *os.LinkError
				require.ErrorAs(t, err, &pathErr)
				assert.Equal(t, "rename", pathErr.Op)
				assert.Equal(t, filepath.Join(root, v1.LockFile), pathErr.New)
			} else {
				require.NoError(t, err)
				lock, err := readV1Lock(filepath.Join(root, v1.LockFile))
				require.NoError(t, err)
				assert.Equal(t, 1, lock.Version)
			}
			current, err := os.ReadFile(filepath.Join(root, v1.ModuleFile))
			require.NoError(t, err)
			assert.Equal(t, tt.wantManifest, string(current))
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			assert.Len(t, entries, 2)
		})
	}
}
