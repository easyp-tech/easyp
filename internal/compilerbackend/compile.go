// Package compilerbackend compiles protos for every CLI stage with easyp-protoc.
package compilerbackend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	easyp "github.com/easyp-tech/easyp/internal/easypc/protoc"
	"github.com/easyp-tech/easyp/internal/parallel"
)

// Request mirrors protocompile.SourceResolver semantics: each import is looked
// up as Accessor(join(importPath, name)) in ImportPaths order, or as
// Accessor(name) when ImportPaths is empty.
type Request struct {
	Files           []string
	ImportPaths     []string
	Accessor        func(string) (io.ReadCloser, error)
	StandardImports bool
	SourceInfo      bool
	MaxParallelism  int
	// DescriptorsOnly skips building the protoreflect registry: Result.Files
	// stays nil and the caller validates Protos itself when it needs to.
	DescriptorsOnly bool
}

// Result holds compiled targets in request order and their FileDescriptorProtos
// keyed by file path.
type Result struct {
	Files  []protoreflect.FileDescriptor
	Protos map[string]*descriptorpb.FileDescriptorProto
}

// Compile compiles the request with easyp-protoc.
func Compile(ctx context.Context, req Request) (*Result, error) {
	if os.Getenv("EASYP_COMPILER_TIMING") != "" {
		start := time.Now()
		defer func() {
			fmt.Fprintf(os.Stderr, "compilerbackend files=%d took=%s\n", len(req.Files), time.Since(start))
		}()
	}
	return compileEasyp(ctx, req)
}

func compileEasyp(ctx context.Context, req Request) (*Result, error) {
	var (
		mu    sync.Mutex
		diags []string
	)
	opts := []easyp.Option{
		easyp.WithResolver(accessorResolver{importPaths: req.ImportPaths, accessor: req.Accessor}),
		easyp.WithStandardImports(req.StandardImports),
		easyp.WithIncludeImports(true),
		easyp.WithIncludeSourceInfo(req.SourceInfo),
		easyp.WithErrorHandler(func(d easyp.Diagnostic) {
			if d.IsWarn {
				return
			}
			mu.Lock()
			diags = append(diags, d.String())
			mu.Unlock()
		}),
	}
	if req.MaxParallelism > 0 {
		opts = append(opts, easyp.WithConcurrency(req.MaxParallelism))
	}
	mark := phaseTimer()
	compiled, err := easyp.Compile(ctx, req.Files, opts...)
	mark("easyp.Compile")
	if err != nil {
		if len(diags) > 0 {
			return nil, fmt.Errorf("%s: %w", strings.Join(diags, "\n"), err)
		}
		return nil, err
	}
	set, err := decodeDescriptorSet(ctx, compiled.WireBytes())
	mark("DescriptorSet")
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("DescriptorSet: easyp-protoc returned undecodable output")
	}
	protos := make(map[string]*descriptorpb.FileDescriptorProto, len(set.GetFile()))
	for _, file := range set.GetFile() {
		protos[file.GetName()] = file
	}
	if req.DescriptorsOnly {
		for _, name := range req.Files {
			if protos[name] == nil {
				return nil, fmt.Errorf("FindFileByPath: %q not found", name)
			}
		}
		return &Result{Protos: protos}, nil
	}
	registry, err := protodesc.NewFiles(set)
	mark("protodesc.NewFiles")
	if err != nil {
		return nil, fmt.Errorf("NewFiles: %w", err)
	}
	files := make([]protoreflect.FileDescriptor, 0, len(req.Files))
	for _, name := range req.Files {
		file, err := registry.FindFileByPath(name)
		if err != nil {
			return nil, fmt.Errorf("FindFileByPath: %w", err)
		}
		files = append(files, file)
	}
	return &Result{Files: files, Protos: protos}, nil
}

// decodeDescriptorSet unmarshals a wire FileDescriptorSet with its files
// decoded in parallel. A set holding anything but files is decoded whole.
func decodeDescriptorSet(ctx context.Context, wire []byte) (*descriptorpb.FileDescriptorSet, error) {
	var records [][]byte
	for rest := wire; len(rest) > 0; {
		number, kind, n := protowire.ConsumeTag(rest)
		if n < 0 || number != 1 || kind != protowire.BytesType {
			records = nil
			break
		}
		record, m := protowire.ConsumeBytes(rest[n:])
		if m < 0 {
			records = nil
			break
		}
		records = append(records, record)
		rest = rest[n+m:]
	}
	if records == nil && len(wire) > 0 {
		var set descriptorpb.FileDescriptorSet
		if err := proto.Unmarshal(wire, &set); err != nil {
			return nil, err
		}
		return &set, nil
	}
	set := &descriptorpb.FileDescriptorSet{File: make([]*descriptorpb.FileDescriptorProto, len(records))}
	err := parallel.Run(ctx, len(records), 0, func(index int) error {
		file := &descriptorpb.FileDescriptorProto{}
		set.File[index] = file
		return proto.Unmarshal(records[index], file)
	})
	if err != nil {
		return nil, err
	}
	return set, nil
}

// phaseTimer prints the time since its previous mark when EASYP_COMPILER_TIMING is set.
func phaseTimer() func(string) {
	if os.Getenv("EASYP_COMPILER_TIMING") == "" {
		return func(string) {}
	}
	last := time.Now()
	return func(phase string) {
		now := time.Now()
		fmt.Fprintf(os.Stderr, "compilerbackend   %s took=%s\n", phase, now.Sub(last))
		last = now
	}
}

type accessorResolver struct {
	importPaths []string
	accessor    func(string) (io.ReadCloser, error)
}

// FindFileByPath implements easyp.Resolver with protocompile.SourceResolver lookup order.
func (r accessorResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	err := ctx.Err()
	if err != nil {
		return nil, err
	}
	if len(r.importPaths) == 0 {
		return r.read(path)
	}
	for _, importPath := range r.importPaths {
		data, err := r.read(filepath.Join(importPath, path))
		if errors.Is(err, easyp.ErrFileNotFound) {
			continue
		}
		return data, err
	}
	return nil, easyp.ErrFileNotFound
}

// traceAccessor logs every resolver read when EASYP_COMPILER_TRACE is set.
var traceAccessor = func() func(string, []byte, error) {
	if os.Getenv("EASYP_COMPILER_TRACE") == "" {
		return nil
	}
	return func(name string, data []byte, err error) {
		fmt.Fprintf(os.Stderr, "compilerbackend read %q err=%v bytes=%d\n", name, err, len(data))
	}
}()

func (r accessorResolver) read(name string) (data []byte, err error) {
	if traceAccessor != nil {
		defer func() { traceAccessor(name, data, err) }()
	}
	reader, err := r.accessor(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, easyp.ErrFileNotFound
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

// ReaderAccessor adapts a resolver that returns plain readers.
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

// BytesAccessor adapts a resolver that returns file contents.
func BytesAccessor(read func(string) ([]byte, error)) func(string) (io.ReadCloser, error) {
	return func(name string) (io.ReadCloser, error) {
		data, err := read(name)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}
