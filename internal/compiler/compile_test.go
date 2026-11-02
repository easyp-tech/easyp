package compiler

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCompileUsesAuthorizedVirtualSourcesAndTargetOrder(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sources := map[string]string{
		filepath.Join(root, "shared.proto"): `syntax="proto3"; package shared; message Value { string text=1; }`,
		filepath.Join(root, "a.proto"):      `syntax="proto3"; package a; import "shared.proto"; message A { shared.Value value=1; }`,
		filepath.Join(root, "b.proto"):      `syntax="proto3"; package b; message B { string value=1; }`,
	}
	accessor := func(name string) (io.ReadCloser, error) {
		text, ok := sources[name]
		if !ok {
			return nil, fs.ErrNotExist
		}
		return io.NopCloser(strings.NewReader(text)), nil
	}
	files, err := Compile(t.Context(), []string{"b.proto", "a.proto", "b.proto"}, Options{
		Resolve: SourceResolver([]string{root}, accessor, true), SourceInfo: true,
	})
	require.NoError(t, err)
	require.Len(t, files, 3)
	assert.Equal(t, "b.proto", files[0].Path())
	assert.Equal(t, "a.proto", files[1].Path())
	assert.Same(t, files[0], files[2])
	require.Equal(t, 1, files[1].Imports().Len())
	assert.Equal(t, "shared.proto", files[1].Imports().Get(0).Path())
	assert.Equal(t, protoreflect.FullName("shared.Value"), files[1].Messages().Get(0).Fields().Get(0).Message().FullName())
}

func TestCompilePreservesResolverFailure(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		err  error
	}{
		{name: "permission", err: fs.ErrPermission},
		{name: "missing", err: fs.ErrNotExist},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			files, err := Compile(t.Context(), []string{"entry.proto"}, Options{Resolve: func(name string) (io.ReadCloser, error) {
				if name == "entry.proto" {
					return io.NopCloser(strings.NewReader("syntax=\"proto3\";\n\n\nimport \"denied.proto\";\nmessage Entry {}")), nil
				}
				return nil, tt.err
			}})
			require.ErrorIs(t, err, tt.err)
			assert.Nil(t, files)
			assert.Contains(t, err.Error(), "denied.proto")
		})
	}
}

func TestSourceResolverRespectsRootOrderAndTerminalErrors(t *testing.T) {
	t.Parallel()
	denied := errors.New("source denied")
	var opened []string
	resolve := SourceResolver([]string{"first", "second"}, func(name string) (io.ReadCloser, error) {
		opened = append(opened, name)
		return nil, denied
	}, true)
	_, err := resolve("google/protobuf/empty.proto")
	require.ErrorIs(t, err, denied)
	assert.Equal(t, []string{filepath.Join("first", "google/protobuf/empty.proto")}, opened)
}

func TestCompileReportsSourceLocationsAndSyntaxErrors(t *testing.T) {
	t.Parallel()
	resolve := func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("syntax=\"proto3\";\npackage api;\nmessage Item {\n  string value=1;\n}\n")), nil
	}
	files, err := Compile(t.Context(), []string{"api.proto"}, Options{Resolve: resolve, SourceInfo: true})
	require.NoError(t, err)
	field := files[0].Messages().Get(0).Fields().Get(0)
	location := files[0].SourceLocations().ByDescriptor(field)
	assert.Equal(t, 3, location.StartLine)
	assert.Equal(t, 2, location.StartColumn)
	_, err = Compile(t.Context(), []string{"bad.proto"}, Options{Resolve: func(string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("syntax=\"proto3\";\nmessage Item {\n  Missing value=1;\n}\n")), nil
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad.proto:3:3")
}

func TestCompileCanceledDoesNotOpenSources(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := Compile(ctx, []string{"input.proto"}, Options{Resolve: func(string) (io.ReadCloser, error) {
		t.Error("canceled compilation opened a source")
		return nil, fs.ErrNotExist
	}})
	require.ErrorIs(t, err, context.Canceled)
}

type failingSource struct {
	readErr  error
	closeErr error
	closed   bool
}

func (s *failingSource) Read([]byte) (int, error) { return 0, s.readErr }
func (s *failingSource) Close() error {
	s.closed = true
	return s.closeErr
}

func TestCompileClosesSourcesAndPreservesBothErrors(t *testing.T) {
	t.Parallel()
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	source := &failingSource{readErr: readErr, closeErr: closeErr}
	files, err := Compile(t.Context(), []string{"input.proto"}, Options{Resolve: func(string) (io.ReadCloser, error) { return source, nil }})
	require.ErrorIs(t, err, readErr)
	require.ErrorIs(t, err, closeErr)
	assert.True(t, source.closed)
	assert.Nil(t, files)
}
