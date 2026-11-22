// Package compilerbackend wraps protobuf compilation behind a single backend boundary.
package compilerbackend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/easyp-tech/easyprotoc/protoc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ErrNoInputFiles reports a compilation request with no target files.
var ErrNoInputFiles = errors.New("no input files provided")

// Request configures one compilation run. Each import is looked up as
// Accessor(join(importPath, name)) in ImportPaths order, or as Accessor(name)
// when ImportPaths is empty.
type Request struct {
	Files           []string
	ImportPaths     []string
	Accessor        func(string) (io.ReadCloser, error)
	StandardImports bool
	SourceInfo      bool
	MaxParallelism  int
}

// Result holds compiled target descriptors in request order and
// FileDescriptorProtos for every compiled target and transitively imported
// file, keyed by file path.
type Result struct {
	Files  []protoreflect.FileDescriptor
	Protos map[string]*descriptorpb.FileDescriptorProto
}

// Compile compiles the requested proto files and returns linked descriptors
// alongside their descriptor protos.
func Compile(ctx context.Context, req Request) (*Result, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if len(req.Files) == 0 {
		return nil, ErrNoInputFiles
	}
	return compileEasyprotoc(ctx, req)
}

func compileEasyprotoc(ctx context.Context, req Request) (*Result, error) {
	open := req.Accessor
	if open == nil {
		open = func(name string) (io.ReadCloser, error) {
			return os.Open(name)
		}
	}

	var (
		diagnosticsMu sync.Mutex
		diagnostics   []string
	)
	opts := []protoc.Option{
		protoc.WithResolver(accessorResolver{
			importPaths: req.ImportPaths,
			open:        open,
		}),
		protoc.WithStandardImports(req.StandardImports),
		protoc.WithIncludeImports(true),
		protoc.WithIncludeSourceInfo(req.SourceInfo),
		protoc.WithErrorHandler(func(diag protoc.Diagnostic) {
			if diag.IsWarn {
				return
			}
			diagnosticsMu.Lock()
			diagnostics = append(diagnostics, diag.String())
			diagnosticsMu.Unlock()
		}),
	}
	if req.MaxParallelism > 0 {
		opts = append(opts, protoc.WithConcurrency(req.MaxParallelism))
	}

	compiled, err := protoc.Compile(ctx, req.Files, opts...)
	if err != nil {
		if len(diagnostics) > 0 {
			return nil, fmt.Errorf("%s: %w", strings.Join(diagnostics, "\n"), err)
		}
		return nil, err
	}

	compiledFiles, err := compiled.Files()
	if err != nil {
		return nil, fmt.Errorf("Files: %w", err)
	}
	byPath := make(map[string]protoreflect.FileDescriptor, len(compiledFiles))
	for _, fd := range compiledFiles {
		byPath[fd.Path()] = fd
	}

	files := make([]protoreflect.FileDescriptor, len(req.Files))
	for i, name := range req.Files {
		fd, ok := byPath[name]
		if !ok {
			clean := strings.TrimPrefix(filepath.Clean(filepath.ToSlash(name)), "./")
			fd, ok = byPath[clean]
		}
		if !ok {
			return nil, fmt.Errorf("compiled target %q not found in descriptor set", name)
		}
		files[i] = fd
	}

	ds := compiled.DescriptorSet()
	protos := make(map[string]*descriptorpb.FileDescriptorProto, len(ds.GetFile()))
	for _, fdp := range ds.GetFile() {
		protos[fdp.GetName()] = fdp
	}

	return &Result{
		Files:  files,
		Protos: protos,
	}, nil
}

type accessorResolver struct {
	importPaths []string
	open        func(string) (io.ReadCloser, error)
}

func (r accessorResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if len(r.importPaths) == 0 {
		return r.read(path)
	}
	native := filepath.FromSlash(path)
	for _, importPath := range r.importPaths {
		data, err := r.read(filepath.Join(importPath, native))
		if errors.Is(err, protoc.ErrFileNotFound) {
			continue
		}
		return data, err
	}
	return nil, protoc.ErrFileNotFound
}

func (r accessorResolver) read(name string) ([]byte, error) {
	reader, err := r.open(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, fs.ErrNotExist) {
			return nil, protoc.ErrFileNotFound
		}
		return nil, err
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, fmt.Errorf("ReadAll: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("Close: %w", closeErr)
	}
	return data, nil
}

// ReaderAccessor adapts a lookup function returning plain readers.
func ReaderAccessor(open func(string) (io.Reader, error)) func(string) (io.ReadCloser, error) {
	return func(name string) (io.ReadCloser, error) {
		reader, err := open(name)
		if err != nil {
			return nil, err
		}
		if closer, ok := reader.(io.ReadCloser); ok {
			return closer, nil
		}
		return io.NopCloser(reader), nil
	}
}

// BytesAccessor adapts a lookup function returning raw file contents.
func BytesAccessor(read func(string) ([]byte, error)) func(string) (io.ReadCloser, error) {
	return func(name string) (io.ReadCloser, error) {
		data, err := read(name)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}
