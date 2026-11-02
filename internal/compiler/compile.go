// Package compiler adapts the native protobuf compiler to EasyP source views.
package compiler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	native "github.com/easyp-tech/easyp-test-compiler/protoc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Resolver opens a logical proto name within the caller's authorized source view.
type Resolver func(name string) (io.ReadCloser, error)

// Options controls compilation without granting independent filesystem access.
type Options struct {
	Resolve        Resolver
	SourceInfo     bool
	MaxParallelism int
}

// Compile returns linked targets in request order, retaining their import closure.
// Resolver decisions are authoritative; compilation never opens another source.
func Compile(ctx context.Context, targets []string, options Options) ([]protoreflect.FileDescriptor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, nil
	}
	if options.Resolve == nil {
		return nil, fmt.Errorf("source resolver is required")
	}
	var diagnostics []error
	resolve := native.ResolverFunc(func(ctx context.Context, name string) ([]byte, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !validImportName(name) {
			return nil, fmt.Errorf("invalid proto import path %q", name)
		}
		source, err := options.Resolve(name)
		if err != nil {
			return nil, err
		}
		if source == nil {
			return nil, fmt.Errorf("resolver returned no source for %q", name)
		}
		data, readErr := io.ReadAll(source)
		closeErr := source.Close()
		if readErr != nil || closeErr != nil {
			return nil, errors.Join(wrapError("ReadAll", readErr), wrapError("Close", closeErr))
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return data, nil
	})
	result, err := native.Compile(ctx, targets,
		native.WithResolverOnly(resolve),
		native.WithIncludeImports(true),
		native.WithIncludeSourceInfo(options.SourceInfo),
		native.WithConcurrency(options.MaxParallelism),
		native.WithErrorHandler(func(diagnostic native.Diagnostic) {
			if !diagnostic.IsWarn {
				diagnostics = append(diagnostics, errors.New(diagnostic.String()))
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("Compile: %w", errors.Join(append(diagnostics, err)...))
	}
	files, err := result.FilesWithError()
	if err != nil {
		return nil, fmt.Errorf("FilesWithError: %w", err)
	}
	byName := make(map[string]protoreflect.FileDescriptor, len(files))
	for _, file := range files {
		byName[file.Path()] = file
	}
	targetFiles := make([]protoreflect.FileDescriptor, 0, len(targets))
	for _, target := range targets {
		name := filepath.ToSlash(filepath.Clean(target))
		file := byName[name]
		if file == nil {
			return nil, fmt.Errorf("compiled target %q has no descriptor", target)
		}
		targetFiles = append(targetFiles, file)
	}
	return targetFiles, nil
}

func validImportName(name string) bool {
	return name != "." && fs.ValidPath(name) && !strings.ContainsAny(name, "\\:\x00")
}

func wrapError(callee string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", callee, err)
}
