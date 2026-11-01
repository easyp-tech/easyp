package core

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// ImmutableData represents an immutable byte slice.
type ImmutableData struct {
	data []byte
}

func newImmutableData(data []byte) *ImmutableData {
	return &ImmutableData{data: data}
}

func (i *ImmutableData) Data() []byte {
	return i.data
}

// GenerateBucket is a thread-safe in-memory bucket for storing generated files.
// It provides methods for adding, retrieving, and removing files.
type GenerateBucket struct {
	filesToWrite map[string]*ImmutableData
	outputPaths  map[string][]string
	lock         sync.RWMutex
}

func NewGenerateBucket() *GenerateBucket {
	return &GenerateBucket{
		filesToWrite: make(map[string]*ImmutableData),
		outputPaths:  make(map[string][]string),
	}
}

func (b *GenerateBucket) PutFile(_ context.Context, path string, data []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.filesToWrite[path] = newImmutableData(data)
}

func (b *GenerateBucket) GetFile(_ context.Context, path string) (*ImmutableData, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	file, ok := b.filesToWrite[path]
	return file, ok
}

// recordOutputPath retains response filenames in their output directory across
// plugins and plans. Several producers may share a filename but relocate it to
// distinct packages, so that filename becomes ambiguous for later insertions.
func (b *GenerateBucket) recordOutputPath(originalPath, outputPath string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !slices.Contains(b.outputPaths[originalPath], outputPath) {
		b.outputPaths[originalPath] = append(b.outputPaths[originalPath], outputPath)
	}
}

func (b *GenerateBucket) insertionOutputPath(originalPath string) (string, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	paths := b.outputPaths[originalPath]
	switch len(paths) {
	case 0:
		return originalPath, nil
	case 1:
		return paths[0], nil
	default:
		return "", fmt.Errorf("ambiguous insertion target %q maps to multiple generated output paths %q", originalPath, paths)
	}
}

func (b *GenerateBucket) DumpToFs(ctx context.Context) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	paths := slices.Sorted(maps.Keys(b.filesToWrite))

	// Validate the complete generated Go layout before writing any file.
	packages := make(map[string]string)
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		file := b.filesToWrite[path]
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, file.Data(), parser.PackageClauseOnly)
		if err != nil {
			return fmt.Errorf("invalid generated Go package in %s: %w", path, err)
		}
		key := filepath.Dir(path)
		if strings.HasSuffix(path, "_test.go") && strings.HasSuffix(parsed.Name.Name, "_test") {
			continue
		}
		if previous, ok := packages[key]; ok && previous != parsed.Name.Name {
			return fmt.Errorf("conflicting Go packages %q and %q in %s; use distinct out directories and matching go_package import paths", previous, parsed.Name.Name, key)
		}
		packages[key] = parsed.Name.Name
	}
	// Apply the completed bucket in path order so write failures are deterministic too.
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		file := b.filesToWrite[path]
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("os.MkdirAll %s: %w", dir, err)
		}

		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("os.Create %s: %w", path, err)
		}

		_, err = f.Write(file.Data())
		if err != nil {
			return fmt.Errorf("f.Write %s: %w", path, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("f.Close %s: %w", path, err)
		}
	}

	return nil
}
