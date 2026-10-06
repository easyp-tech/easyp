package workspace

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/sourceview"
)

// ReadFile reads workspace metadata at its logical path through a bounded root.
func ReadFile(root, path string) ([]byte, error) {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("Rel: %w", err)
	}
	return sourceview.ReadLocal(context.Background(), root, relative)
}

// ReadFileAt reads local metadata within the discovered workspace boundary.
func ReadFileAt(path string) ([]byte, error) {
	boundary, err := Boundary(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("Boundary: %w", err)
	}
	return ReadFile(boundary, path)
}

// Walk traverses logical workspace paths, including bounded directory aliases.
func Walk(root string, visit func(string, fs.DirEntry, error) error) error {
	return WalkAt(root, root, visit)
}

// WalkAt walks a logical subtree within its broader workspace boundary.
func WalkAt(boundary, root string, visit func(string, fs.DirEntry, error) error) error {
	relative, err := filepath.Rel(boundary, root)
	if err != nil {
		return fmt.Errorf("Rel: %w", err)
	}
	return sourceview.WalkLocal(context.Background(), boundary, relative, func(logical string, resolved sourceview.Resolution, walkErr error) error {
		path := filepath.Join(boundary, filepath.FromSlash(logical))
		if SkipDirectory(root, path) {
			if resolved.Info == nil || !resolved.Info.IsDir() {
				return nil
			}
			return fs.SkipDir
		}
		if walkErr != nil {
			switch filepath.Base(path) {
			case "easyp.yaml", "easyp.gen.yaml", "protobuf.mod", "protobuf.lock", "buf.yaml", "buf.work.yaml", "buf.lock":
			default:
				if filepath.Ext(path) != ".proto" && logical != filepath.ToSlash(relative) {
					return nil
				}
			}
			var entry fs.DirEntry
			if resolved.Info != nil {
				entry = fs.FileInfoToDirEntry(resolved.Info)
			}
			return visit(path, entry, walkErr)
		}
		if resolved.Info.IsDir() && SkipDirectory(physicalRoot(root), filepath.Join(physicalRoot(boundary), filepath.FromSlash(resolved.Path))) {
			return fs.SkipDir
		}
		return visit(path, fs.FileInfoToDirEntry(resolved.Info), nil)
	})
}

func physicalRoot(root string) string {
	canonical, err := filepath.EvalSymlinks(root)
	if err == nil {
		return canonical
	}
	return root
}
