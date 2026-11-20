// Package compilerbackend wraps protobuf compilation behind a single backend boundary.
package compilerbackend

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/wellknownimports"
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
		err := ctx.Err()
		if err != nil {
			return nil, err
		}
	}
	if len(req.Files) == 0 {
		return nil, ErrNoInputFiles
	}
	return compileProtocompile(ctx, req)
}

func compileProtocompile(ctx context.Context, req Request) (*Result, error) {
	var resolver protocompile.Resolver = &protocompile.SourceResolver{
		ImportPaths: req.ImportPaths,
		Accessor:    req.Accessor,
	}
	if req.StandardImports {
		resolver = wellknownimports.WithStandardImports(resolver)
	}
	compiler := protocompile.Compiler{
		Resolver:       resolver,
		MaxParallelism: req.MaxParallelism,
	}
	if req.SourceInfo {
		compiler.SourceInfoMode = protocompile.SourceInfoStandard
	}
	compiled, err := compiler.Compile(ctx, req.Files...)
	if err != nil {
		return nil, err
	}

	files := make([]protoreflect.FileDescriptor, len(compiled))
	protos := make(map[string]*descriptorpb.FileDescriptorProto, len(compiled))
	var collect func(protoreflect.FileDescriptor)
	collect = func(file protoreflect.FileDescriptor) {
		name := file.Path()
		if _, exists := protos[name]; exists {
			return
		}
		protos[name] = protoutil.ProtoFromFileDescriptor(file)
		for i := range file.Imports().Len() {
			collect(file.Imports().Get(i).FileDescriptor)
		}
	}
	for i, file := range compiled {
		files[i] = file
		collect(file)
	}
	return &Result{
		Files:  files,
		Protos: protos,
	}, nil
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
