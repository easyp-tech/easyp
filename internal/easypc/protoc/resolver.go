package protoc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/protoc/wkt"
)

// ErrFileNotFound indicates that a requested proto file could not be found.
var ErrFileNotFound = errors.New("file not found")

// Resolver abstracts file resolution for the protoc compiler.
type Resolver interface {
	// FindFileByPath locates a file by its path and returns its raw bytes.
	// Returns ErrFileNotFound if the file does not exist.
	FindFileByPath(ctx context.Context, path string) ([]byte, error)
}

// MemResolver resolves proto files directly from an in-memory map.
type MemResolver map[string][]byte

// FindFileByPath implements Resolver.
func (m MemResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	clean := filepath.Clean(filepath.ToSlash(path))
	clean = strings.TrimPrefix(clean, "./")
	if b, ok := m[clean]; ok {
		return b, nil
	}
	if b, ok := m[path]; ok {
		return b, nil
	}
	return nil, ErrFileNotFound
}

// DirResolver resolves files relative to a filesystem directory.
type DirResolver string

// FindFileByPath implements Resolver.
func (d DirResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	full := filepath.Join(string(d), path)
	b, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}
	return b, nil
}

// MultiResolver queries multiple resolvers in order, returning from the first that succeeds.
type MultiResolver []Resolver

// FindFileByPath implements Resolver.
func (mr MultiResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	for _, r := range mr {
		if ctx != nil && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if r == nil {
			continue
		}
		b, err := r.FindFileByPath(ctx, path)
		if err == nil {
			return b, nil
		}
		if !errors.Is(err, ErrFileNotFound) {
			return nil, err
		}
	}
	return nil, ErrFileNotFound
}

// WKTResolver resolves standard Well-Known Type proto definitions from embedded assets.
type WKTResolver struct{}

// FindFileByPath implements Resolver.
func (w WKTResolver) FindFileByPath(ctx context.Context, path string) ([]byte, error) {
	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}
	clean := strings.TrimPrefix(filepath.ToSlash(path), "/")
	b, err := wkt.ReadFile(clean)
	if err != nil {
		return nil, ErrFileNotFound
	}
	return b, nil
}
