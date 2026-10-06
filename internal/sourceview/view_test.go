package sourceview

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePreservesComponentOrderAndLinkTrace(t *testing.T) {
	t.Parallel()
	view := New(fstest.MapFS{
		"real/nested/file.proto": {Data: []byte("message File {}")},
		"real/sibling.proto":     {Data: []byte("message Sibling {}")},
		"alias":                  {Mode: fs.ModeSymlink, Data: []byte("real/nested")},
		"chain":                  {Mode: fs.ModeSymlink, Data: []byte("alias/file.proto")},
		"chain2":                 {Mode: fs.ModeSymlink, Data: []byte("chain")},
		"relative":               {Mode: fs.ModeSymlink, Data: []byte("alias/../sibling.proto")},
	})
	tests := []struct {
		name  string
		input string
		path  string
		links []string
	}{
		{name: "file chain", input: "chain2", path: "real/nested/file.proto", links: []string{"chain2", "chain", "alias"}},
		{name: "intermediate directory", input: "alias/./file.proto", path: "real/nested/file.proto", links: []string{"alias"}},
		{name: "parent after alias", input: "alias/../sibling.proto", path: "real/sibling.proto", links: []string{"alias"}},
		{name: "parent inside target", input: "relative", path: "real/sibling.proto", links: []string{"relative", "alias"}},
		{name: "root", input: ".", path: "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolved, err := view.Resolve(t.Context(), tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.path, resolved.Path)
			require.NotNil(t, resolved.Info)
			var paths []string
			for _, link := range resolved.Links {
				paths = append(paths, link.Path)
				assert.NotEmpty(t, link.Target)
				assert.Equal(t, fs.ModeSymlink, link.Info.Mode().Type())
			}
			assert.Equal(t, tt.links, paths)
		})
	}
	file, err := view.Open(t.Context(), "chain2")
	require.NoError(t, err)
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	assert.Equal(t, "message File {}", string(data))
}

func TestResolveLinkCyclesUsePhysicalPaths(t *testing.T) {
	t.Parallel()
	view := New(fstest.MapFS{
		"one":                {Mode: fs.ModeSymlink, Data: []byte("two")},
		"two":                {Mode: fs.ModeSymlink, Data: []byte("one")},
		"root":               {Mode: fs.ModeSymlink, Data: []byte("inner/leaf")},
		"inner/leaf":         {Mode: fs.ModeSymlink, Data: []byte("inner/leaf")},
		"inner/inner/leaf":   {Data: []byte("same link contents are not identities")},
		"again":              {Mode: fs.ModeSymlink, Data: []byte(".")},
		"regular/file.proto": {Data: []byte("file")},
		"expanding":          {Mode: fs.ModeSymlink, Data: []byte("expanding/child")},
		"target-parent":      {Mode: fs.ModeSymlink, Data: []byte("regular/../target-parent")},
	})
	for _, input := range []string{"one", "expanding", "target-parent"} {
		_, err := view.Resolve(t.Context(), input)
		require.ErrorIs(t, err, ErrCycle)
	}
	resolved, err := view.Resolve(t.Context(), "root")
	require.NoError(t, err)
	assert.Equal(t, "inner/inner/leaf", resolved.Path)
	require.Len(t, resolved.Links, 2)
	assert.Equal(t, resolved.Links[0].Target, resolved.Links[1].Target)
	resolved, err = view.Resolve(t.Context(), "again/again/regular/file.proto")
	require.NoError(t, err)
	assert.Equal(t, "regular/file.proto", resolved.Path)
}

func TestResolveRejectsUnsafeAndIrregularTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		target string
		want   error
	}{
		{name: "parent escape", target: "../outside", want: ErrOutsideRoot},
		{name: "parent after root alias", target: "root/../outside", want: ErrOutsideRoot},
		{name: "absolute", target: "/outside/secret", want: ErrOutsideRoot},
		{name: "drive absolute", target: "C:/outside", want: ErrUnsupported},
		{name: "drive relative", target: "C:outside", want: ErrUnsupported},
		{name: "UNC", target: "//server/share", want: ErrUnsupported},
		{name: "backslash", target: "dir\\file", want: ErrUnsupported},
		{name: "NUL", target: "dir\x00file", want: ErrUnsupported},
		{name: "dangling", target: "missing", want: fs.ErrNotExist},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			backend := &openedFS{MapFS: fstest.MapFS{
				"pointer": {Mode: fs.ModeSymlink, Data: []byte(tt.target)},
				"root":    {Mode: fs.ModeSymlink, Data: []byte(".")},
			}}
			view := New(backend)
			resolved, err := view.Resolve(t.Context(), "pointer")
			require.ErrorIs(t, err, tt.want)
			require.NotEmpty(t, resolved.Links)
			assert.Equal(t, tt.target, resolved.Links[0].Target)
			_, err = view.Open(t.Context(), "pointer")
			require.ErrorIs(t, err, tt.want)
			assert.Empty(t, backend.opened, "rejected targets must not be opened")
		})
	}
	view := New(fstest.MapFS{
		"gitlink": {Mode: fs.ModeIrregular},
		"fifo":    {Mode: fs.ModeNamedPipe},
		"file":    {Data: []byte("file")},
	})
	for _, input := range []string{"gitlink", "fifo"} {
		resolved, err := view.Resolve(t.Context(), input)
		require.ErrorIs(t, err, ErrUnsupported)
		require.NotNil(t, resolved.Info)
		assert.False(t, resolved.Info.Mode().IsRegular())
	}
	for _, input := range []string{"file/..", "file/.", "file/child"} {
		_, err := view.Resolve(t.Context(), input)
		require.Error(t, err, input)
	}
}

func TestLocalAbsolutePointersStayInsideRoot(t *testing.T) {
	t.Parallel()
	rootPath := t.TempDir()
	canonicalRoot, err := filepath.EvalSymlinks(rootPath)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(rootPath, "real", "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "real", "file.proto"), []byte("bounded"), 0o600))
	require.NoError(t, os.Symlink("real/nested", filepath.Join(rootPath, "dir")))
	absTarget := canonicalRoot + "/dir/../file.proto"
	require.NoError(t, os.Symlink(absTarget, filepath.Join(rootPath, "absolute")))
	require.NoError(t, os.Symlink(canonicalRoot, filepath.Join(rootPath, "root")))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret"), []byte("external"), 0o600))
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret"), filepath.Join(rootPath, "external")))
	require.NoError(t, os.Symlink(canonicalRoot+"/../external", filepath.Join(rootPath, "parent-escape")))
	root, err := os.OpenRoot(rootPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	view, err := NewLocal(root.FS(), canonicalRoot)
	require.NoError(t, err)
	resolved, err := view.Resolve(t.Context(), "absolute")
	require.NoError(t, err)
	assert.Equal(t, "real/file.proto", resolved.Path)
	assert.Equal(t, absTarget, resolved.Links[0].Target)
	file, err := view.Open(t.Context(), "root/absolute")
	require.NoError(t, err)
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	assert.Equal(t, "bounded", string(data))
	for _, input := range []string{"external", "parent-escape"} {
		_, err = view.Open(t.Context(), input)
		require.ErrorIs(t, err, ErrOutsideRoot)
	}
	_, err = New(root.FS()).Resolve(t.Context(), "absolute")
	require.ErrorIs(t, err, ErrOutsideRoot)
	_, err = NewLocal(root.FS(), "relative")
	require.ErrorIs(t, err, fs.ErrInvalid)
	_, err = root.Stat("real/file.proto")
	require.NoError(t, err, "view must not close its borrowed root")
}

func TestLocalVerifiedRootSpellingsPreserveUncleanedPointer(t *testing.T) {
	t.Parallel()
	backend := &openedFS{MapFS: fstest.MapFS{
		"real/nested": {Mode: fs.ModeDir},
		"real/file":   {Data: []byte("bounded")},
		"dir":         {Mode: fs.ModeSymlink, Data: []byte("real/nested")},
		"pointer":     {Mode: fs.ModeSymlink, Data: []byte("/var/source/dir/../file")},
	}}
	view, err := NewLocal(backend, "/private/var/source", "/var/source")
	require.NoError(t, err)
	resolved, err := view.Resolve(t.Context(), "pointer")
	require.NoError(t, err)
	assert.Equal(t, "real/file", resolved.Path)
	assert.Equal(t, "/var/source/dir/../file", resolved.Links[0].Target)
	assert.Empty(t, backend.opened)
	_, err = NewLocal(backend, "/private/var/source", "relative")
	require.ErrorIs(t, err, fs.ErrInvalid)
}

func TestWalkKeepsDistinctAliasesAndSortedLogicalNames(t *testing.T) {
	t.Parallel()
	view := New(unsortedFS{MapFS: fstest.MapFS{
		"real/z.proto": {Data: []byte("z")},
		"real/a.proto": {Data: []byte("a")},
		"alias-one":    {Mode: fs.ModeSymlink, Data: []byte("real")},
		"alias-two":    {Mode: fs.ModeSymlink, Data: []byte("real")},
	}})
	var visited []string
	err := view.Walk(t.Context(), ".", func(logical string, resolved Resolution, err error) error {
		require.NoError(t, err)
		visited = append(visited, logical+"="+resolved.Path)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		".=.",
		"alias-one=real", "alias-one/a.proto=real/a.proto", "alias-one/z.proto=real/z.proto",
		"alias-two=real", "alias-two/a.proto=real/a.proto", "alias-two/z.proto=real/z.proto",
		"real=real", "real/a.proto=real/a.proto", "real/z.proto=real/z.proto",
	}, visited)
	visited = nil
	err = view.Walk(t.Context(), "alias-one", func(logical string, resolved Resolution, err error) error {
		require.NoError(t, err)
		visited = append(visited, logical)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"alias-one", "alias-one/a.proto", "alias-one/z.proto"}, visited)
}

func TestWalkReportsAuxiliaryErrorsAndAncestorDirectoryCycles(t *testing.T) {
	t.Parallel()
	view := New(fstest.MapFS{
		"dir/file.proto": {Data: []byte("file")},
		"dir/up":         {Mode: fs.ModeSymlink, Data: []byte("..")},
		"unused":         {Mode: fs.ModeSymlink, Data: []byte("/external")},
		"skipped/bad":    {Mode: fs.ModeSymlink, Data: []byte("/external")},
	})
	var visited []string
	var failed []string
	err := view.Walk(t.Context(), ".", func(logical string, resolved Resolution, err error) error {
		visited = append(visited, logical)
		if err != nil {
			failed = append(failed, logical)
			require.NotEmpty(t, resolved.Links)
			if logical == "dir/up" {
				require.ErrorIs(t, err, ErrCycle)
			} else {
				require.ErrorIs(t, err, ErrOutsideRoot)
			}
			return nil
		}
		if logical == "skipped" {
			return fs.SkipDir
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{".", "dir", "dir/file.proto", "dir/up", "skipped", "unused"}, visited)
	assert.Equal(t, []string{"dir/up", "unused"}, failed)
	err = view.Walk(t.Context(), ".", func(_ string, _ Resolution, err error) error { return err })
	require.ErrorIs(t, err, ErrCycle)
}

func TestWalkSkipControlsAndContext(t *testing.T) {
	t.Parallel()
	view := New(fstest.MapFS{"a": {Data: []byte("a")}, "b": {Data: []byte("b")}})
	for _, skip := range []error{fs.SkipDir, fs.SkipAll} {
		var visited []string
		err := view.Walk(t.Context(), ".", func(logical string, _ Resolution, _ error) error {
			visited = append(visited, logical)
			if logical == "a" {
				return skip
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, []string{".", "a"}, visited)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := view.Resolve(ctx, "a")
	require.ErrorIs(t, err, context.Canceled)
	_, err = view.Open(ctx, "a")
	require.ErrorIs(t, err, context.Canceled)
	err = view.Walk(ctx, ".", func(_ string, _ Resolution, _ error) error {
		t.Fatal("canceled walk invoked callback")
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	ctx, cancel = context.WithCancel(t.Context())
	err = view.Walk(ctx, ".", func(logical string, _ Resolution, _ error) error {
		if logical == "." {
			cancel()
		}
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestWalkSkipDirOnDirectoryAliasKeepsSiblings(t *testing.T) {
	t.Parallel()
	view := New(fstest.MapFS{
		"a-alias":     {Mode: fs.ModeSymlink, Data: []byte("directory")},
		"b.proto":     {Data: []byte("b")},
		"directory/a": {Data: []byte("a")},
	})
	var visited []string
	err := view.Walk(t.Context(), ".", func(logical string, _ Resolution, err error) error {
		require.NoError(t, err)
		visited = append(visited, logical)
		if logical == "a-alias" {
			return fs.SkipDir
		}
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{".", "a-alias", "b.proto", "directory", "directory/a"}, visited)
}

func TestCancellationDuringFilesystemOperationCannotBeIgnored(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	backend := cancelFS{MapFS: fstest.MapFS{"file": {Data: []byte("file")}}, cancel: cancel}
	_, err := New(backend).Resolve(ctx, "file")
	require.ErrorIs(t, err, context.Canceled)
	ctx, cancel = context.WithCancel(t.Context())
	backend.cancel = cancel
	err = New(backend).Walk(ctx, "file", func(_ string, _ Resolution, _ error) error {
		t.Fatal("walk must propagate cancellation before callback")
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
}

func TestOpenRejectsReplacementAndClosesOwnedFiles(t *testing.T) {
	t.Parallel()
	backend := &replacedFS{MapFS: fstest.MapFS{"file": {Data: []byte("before")}}}
	_, err := New(backend).Open(t.Context(), "file")
	require.ErrorIs(t, err, ErrChanged)
	assert.True(t, backend.closed)
	view := New(fstest.MapFS{"directory": {Mode: fs.ModeDir}})
	_, err = view.Open(t.Context(), "directory")
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestOpenRejectsNativeIdentityReplacementWithMatchingMetadata(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	stamp := time.Unix(123456789, 0)
	for name, data := range map[string]string{"file": "before", "replacement": "after!"} {
		filename := filepath.Join(directory, name)
		require.NoError(t, os.WriteFile(filename, []byte(data), 0o600))
		require.NoError(t, os.Chtimes(filename, stamp, stamp))
	}
	root, err := os.OpenRoot(directory)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	backend := &nativeReplacementFS{ReadLinkFS: root.FS().(fs.ReadLinkFS), directory: directory}
	_, err = New(backend).Open(t.Context(), "file")
	require.ErrorIs(t, err, ErrChanged)
	require.NotNil(t, backend.opened)
	assert.True(t, backend.opened.closed)
}

func TestOpenStatFailureJoinsCloseError(t *testing.T) {
	t.Parallel()
	statErr := errors.New("stat failure")
	closeErr := errors.New("close failure")
	backend := &failedStatFS{MapFS: fstest.MapFS{"file": {Data: []byte("file")}}, statErr: statErr, closeErr: closeErr}
	_, err := New(backend).Open(t.Context(), "file")
	require.ErrorIs(t, err, statErr)
	require.ErrorIs(t, err, closeErr)
	assert.True(t, backend.closed)
}

type openedFS struct {
	fstest.MapFS
	opened []string
}

func (f *openedFS) Open(name string) (fs.File, error) {
	f.opened = append(f.opened, name)
	return f.MapFS.Open(name)
}

type unsortedFS struct{ fstest.MapFS }

func (f unsortedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := f.MapFS.ReadDir(name)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries, nil
}

type cancelFS struct {
	fstest.MapFS
	cancel context.CancelFunc
}

func (f cancelFS) Lstat(name string) (fs.FileInfo, error) {
	info, err := f.MapFS.Lstat(name)
	if name == "file" {
		f.cancel()
	}
	return info, err
}

type replacedFS struct {
	fstest.MapFS
	closed bool
}

func (f *replacedFS) Open(_ string) (fs.File, error) {
	return &replacedFile{Reader: strings.NewReader("after"), owner: f}, nil
}

type replacedFile struct {
	*strings.Reader
	owner *replacedFS
}

func (f *replacedFile) Close() error {
	f.owner.closed = true
	return nil
}

func (f *replacedFile) Stat() (fs.FileInfo, error) {
	return replacementInfo{}, nil
}

type replacementInfo struct{}

func (replacementInfo) Name() string       { return "file" }
func (replacementInfo) Size() int64        { return 5 }
func (replacementInfo) Mode() fs.FileMode  { return 0 }
func (replacementInfo) ModTime() time.Time { return time.Time{} }
func (replacementInfo) IsDir() bool        { return false }
func (replacementInfo) Sys() any           { return nil }

type nativeReplacementFS struct {
	fs.ReadLinkFS
	directory string
	opened    *ownedFile
}

func (f *nativeReplacementFS) Open(name string) (fs.File, error) {
	err := os.Rename(filepath.Join(f.directory, "replacement"), filepath.Join(f.directory, name))
	if err != nil {
		return nil, err
	}
	file, err := f.ReadLinkFS.Open(name)
	if err != nil {
		return nil, err
	}
	f.opened = &ownedFile{File: file}
	return f.opened, nil
}

type ownedFile struct {
	fs.File
	closed bool
}

func (f *ownedFile) Close() error {
	f.closed = true
	return f.File.Close()
}

type failedStatFS struct {
	fstest.MapFS
	statErr  error
	closeErr error
	closed   bool
}

func (f *failedStatFS) Open(_ string) (fs.File, error) { return failedStatFile{owner: f}, nil }

type failedStatFile struct{ owner *failedStatFS }

func (f failedStatFile) Stat() (fs.FileInfo, error) { return nil, f.owner.statErr }
func (f failedStatFile) Read(_ []byte) (int, error) { return 0, io.EOF }
func (f failedStatFile) Close() error {
	f.owner.closed = true
	return f.owner.closeErr
}
