// Package compiler adapts easyprotoc to EasyP generation source views.
package compiler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/easyp-tech/easyprotoc/protoc"
)

// Resolver opens a logical proto name within the caller's authorized source view.
type Resolver func(name string) (io.ReadCloser, error)

// Options controls compilation without granting independent filesystem access.
type Options struct {
	Resolve        Resolver
	SourceInfo     bool
	MaxParallelism int
}

func compileNative(ctx context.Context, targets []string, options Options) (*protoc.Result, error) {
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
	resolve := protoc.ResolverFunc(func(ctx context.Context, name string) ([]byte, error) {
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
	result, err := protoc.Compile(ctx, targets,
		protoc.WithResolverOnly(resolve),
		protoc.WithIncludeImports(true),
		protoc.WithIncludeSourceInfo(options.SourceInfo),
		protoc.WithConcurrency(options.MaxParallelism),
		protoc.WithErrorHandler(func(diagnostic protoc.Diagnostic) {
			if !diagnostic.IsWarn {
				diagnostics = append(diagnostics, errors.New(diagnostic.String()))
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("Compile: %w", errors.Join(append(diagnostics, err)...))
	}
	return result, nil
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
