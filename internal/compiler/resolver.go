package compiler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	native "github.com/easyp-tech/easyp-test-compiler/protoc"
)

// SourceResolver searches ordered roots through accessor. Only missing candidates
// advance to the next root; authorization and source-integrity errors terminate.
// standardImports explicitly permits embedded standard protobuf sources afterward.
func SourceResolver(importPaths []string, accessor Resolver, standardImports bool) Resolver {
	roots := append([]string(nil), importPaths...)
	if len(roots) == 0 {
		roots = []string{"."}
	}
	if accessor == nil {
		accessor = func(name string) (io.ReadCloser, error) { return os.Open(name) }
	}
	resolve := func(name string) (io.ReadCloser, error) {
		if !validImportName(name) {
			return nil, &os.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
		}
		for _, root := range roots {
			source, err := accessor(filepath.Join(root, filepath.FromSlash(name)))
			if errors.Is(err, fs.ErrNotExist) || errors.Is(err, native.ErrFileNotFound) {
				continue
			}
			return source, err
		}
		return nil, &os.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	if standardImports {
		return StandardImports(resolve)
	}
	return resolve
}

// StandardImports permits embedded well-known sources only after resolve reports
// a missing name. It never overrides a terminal source authorization decision.
func StandardImports(resolve Resolver) Resolver {
	return func(name string) (io.ReadCloser, error) {
		source, err := resolve(name)
		if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, native.ErrFileNotFound) {
			return source, err
		}
		data, builtinErr := (native.WKTResolver{}).FindFileByPath(context.Background(), name)
		if builtinErr != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}
