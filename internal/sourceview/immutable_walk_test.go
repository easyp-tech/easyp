package sourceview

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImmutableWalkReusesParentResolution(t *testing.T) {
	t.Parallel()
	for _, depth := range []int{1, 8, 24} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			t.Parallel()
			backend := &countedWalkFS{MapFS: make(fstest.MapFS)}
			prefix := strings.Repeat("nested/", depth)
			const count = 100
			for index := range count {
				backend.MapFS[fmt.Sprintf("%sfile-%d.proto", prefix, index)] = &fstest.MapFile{Data: []byte("contract")}
			}
			visited := 0
			err := NewImmutable(backend).Walk(t.Context(), ".", func(_ string, _ Resolution, err error) error {
				visited++
				return err
			})
			require.NoError(t, err)
			assert.Equal(t, count+depth+1, visited)
			assert.Equal(t, 1, backend.lstats, "only the walk root requires path resolution")
			assert.Equal(t, count+depth, backend.infos, "each entry supplies its own metadata once")
		})
	}
}

func TestImmutableWalkPreservesInheritedAliases(t *testing.T) {
	t.Parallel()
	view := NewImmutable(fstest.MapFS{
		"real/nested/a.proto": {Data: []byte("a")},
		"real/nested/b.proto": {Data: []byte("b")},
		"alias":               {Mode: fs.ModeSymlink, Data: []byte("real")},
		"chain":               {Mode: fs.ModeSymlink, Data: []byte("alias")},
	})
	var files []string
	err := view.Walk(t.Context(), "chain", func(logical string, resolved Resolution, err error) error {
		require.NoError(t, err)
		require.Len(t, resolved.Links, 2)
		assert.Equal(t, "chain", resolved.Links[0].Path)
		assert.Equal(t, "alias", resolved.Links[1].Path)
		if !resolved.Info.IsDir() {
			files = append(files, logical+"="+resolved.Path)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"chain/nested/a.proto=real/nested/a.proto", "chain/nested/b.proto=real/nested/b.proto"}, files)
}

func TestMutableWalkObservesSourceChanges(t *testing.T) {
	t.Parallel()
	backend := fstest.MapFS{"directory/first.proto": {Data: []byte("first")}}
	view := New(backend)
	var first []string
	err := view.Walk(t.Context(), ".", func(logical string, _ Resolution, err error) error {
		first = append(first, logical)
		return err
	})
	require.NoError(t, err)
	delete(backend, "directory/first.proto")
	backend["directory/second.proto"] = &fstest.MapFile{Data: []byte("second")}
	var second []string
	err = view.Walk(t.Context(), ".", func(logical string, _ Resolution, err error) error {
		second = append(second, logical)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, []string{".", "directory", "directory/first.proto"}, first)
	assert.Equal(t, []string{".", "directory", "directory/second.proto"}, second)
}

func TestImmutableWalkRejectsUnsupportedEntries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		file string
		mode fs.FileMode
	}{
		{name: "drive path", file: "C:outside.proto"},
		{name: "backslash", file: "unsafe\\outside.proto"},
		{name: "named pipe", file: "fifo.proto", mode: fs.ModeNamedPipe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			view := NewImmutable(fstest.MapFS{tt.file: {Mode: tt.mode}})
			err := view.Walk(t.Context(), ".", func(_ string, _ Resolution, err error) error { return err })
			require.ErrorIs(t, err, ErrUnsupported)
		})
	}
}

type countedWalkFS struct {
	fstest.MapFS
	lstats int
	infos  int
}

func (f *countedWalkFS) Lstat(name string) (fs.FileInfo, error) {
	f.lstats++
	return f.MapFS.Lstat(name)
}

func (f *countedWalkFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := f.MapFS.ReadDir(name)
	if err != nil {
		return nil, err
	}
	for index, entry := range entries {
		entries[index] = countedWalkEntry{DirEntry: entry, owner: f}
	}
	return entries, nil
}

type countedWalkEntry struct {
	fs.DirEntry
	owner *countedWalkFS
}

func (entry countedWalkEntry) Info() (fs.FileInfo, error) {
	entry.owner.infos++
	return entry.DirEntry.Info()
}
