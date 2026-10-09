package sourceview

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectedWalkDoesNotResolveUnselectedFiles(t *testing.T) {
	t.Parallel()
	view := New(restrictedWalkFS{FS: fstest.MapFS{
		"api.proto": {Data: []byte("contract")},
		"README.md": {Data: []byte("unrelated")},
	}})
	var names []string
	err := view.WalkSelected(t.Context(), ".", func(name string, entry fs.DirEntry) bool {
		return entry.IsDir() || name == "api.proto"
	}, func(name string, _ Resolution, err error) error {
		names = append(names, name)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, []string{".", "api.proto"}, names)
}

type restrictedWalkFS struct{ fs.FS }

func (f restrictedWalkFS) Open(name string) (fs.File, error) {
	if name == "README.md" {
		return nil, fs.ErrPermission
	}
	return f.FS.Open(name)
}
