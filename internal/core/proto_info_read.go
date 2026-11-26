package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/protoast"
	"github.com/easyp-tech/easyp/wellknownimports"
)

func (c *Core) protoInfoRead(ctx context.Context, fs FS, path string) (ProtoInfo, error) {
	src, err := c.readSource(ctx, fs, path)
	if err != nil {
		return ProtoInfo{}, err
	}
	return c.protoInfoFromSource(ctx, fs, path, src)
}

// readSource reads a linted file through the same opener imports use.
func (c *Core) readSource(ctx context.Context, fs FS, path string) ([]byte, error) {
	f, err := c.openDiskSource(fs, path)
	if err != nil {
		return nil, fmt.Errorf("Open: %w", err)
	}
	defer c.close(ctx, f, path)

	src, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Parse: ReadAll: %w", err)
	}
	return src, nil
}

func (c *Core) protoInfoFromSource(ctx context.Context, fs FS, path string, src []byte) (ProtoInfo, error) {
	file, err := protoast.Parse(path, src)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("Parse: %w", err)
	}

	imported, err := c.readFilesFromImport(ctx, fs, file)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %w", err)
	}

	return ProtoInfo{
		Path:          path,
		ImportPath:    c.sourceImportPath(fs, path),
		File:          file,
		ImportedFiles: imported,
	}, nil
}

// readFilesFromImport parses the direct imports of file.
func (c *Core) readFilesFromImport(ctx context.Context, disk FS, file *protoast.File) (map[ImportPath]*protoast.File, error) {
	imported := make(map[ImportPath]*protoast.File, len(file.Imports))
	for _, imp := range file.Imports {
		importPath := ConvertImportPath(imp.Location)
		parsed, err := c.readFileFromImport(ctx, disk, string(importPath))
		if err != nil {
			return nil, fmt.Errorf("readFileFromImport: %w", err)
		}
		imported[importPath] = parsed
	}
	return imported, nil
}

// importCacheKey identifies importName as openImportFile would find it on
// disk: the lint root plus the name, which the lint FS tries first.
func importCacheKey(disk FS, importName string) (string, bool) {
	rooted, ok := disk.(interface{ RootPath() string })
	if !ok {
		return "", false
	}
	return rooted.RootPath() + "\x00" + importName, true
}

// cacheLintedOutline records a linted file for imports of the same name. The
// linted file was opened by openDiskSource on the lint FS, the first place
// openImportFile looks, so an import of that name resolves to the same file.
func (c *Core) cacheLintedOutline(disk FS, path string, file *protoast.File) {
	key, cacheable := importCacheKey(disk, path)
	if !cacheable || c.importCache == nil {
		return
	}
	c.importCache.mu.Lock()
	if _, ok := c.importCache.files[key]; !ok {
		c.importCache.files[key] = file.Outline()
	}
	c.importCache.mu.Unlock()
}

// readFileFromImport parses an imported file once per lint root and keeps its
// outline: lint rules only look up names declared by imported files.
func (c *Core) readFileFromImport(ctx context.Context, disk FS, importName string) (*protoast.File, error) {
	key, cacheable := importCacheKey(disk, importName)
	if cacheable && c.importCache != nil {
		c.importCache.mu.Lock()
		parsed, ok := c.importCache.files[key]
		c.importCache.mu.Unlock()
		if ok {
			return parsed, nil
		}
	}

	f, err := c.openImportFile(disk, importName)
	if err != nil {
		return nil, fmt.Errorf("openImportFile: %w", err)
	}
	defer c.close(ctx, f, importName)

	parsed, err := protoast.Read(importName, f)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: %w", &os.PathError{Op: "parse", Path: importName, Err: err})
	}
	parsed = parsed.Outline()
	if cacheable && c.importCache != nil {
		c.importCache.mu.Lock()
		c.importCache.files[key] = parsed
		c.importCache.mu.Unlock()
	}
	return parsed, nil
}

func (c *Core) openImportFile(disk FS, importName string) (io.ReadCloser, error) {
	if !filepath.IsLocal(importName) {
		return nil, fmt.Errorf("invalid import path %q", importName)
	}
	var f io.ReadCloser
	var err error
	allowed := true
	if rooted, ok := disk.(interface{ RootPath() string }); ok && c.importFileAllowed != nil {
		allowed = c.importFileAllowed(filepath.Join(rooted.RootPath(), importName))
	}
	if allowed {
		f, err = c.openDiskSource(disk, importName)
	} else {
		err = os.ErrNotExist
	}
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Search dependency roots when the current filesystem has no matching file.
	case err != nil:
		return nil, fmt.Errorf("Open: %w", err)
	default:
		return f, nil
	}

	for _, root := range c.importRoots {
		fullPath := filepath.Join(root, importName)
		f, err := c.openSourceFile(fullPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("Open: %w", err)
		}
		return f, nil
	}

	f, err = wellknownimports.Content.Open(importName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &OpenImportFileError{FileName: importName}
		}
		return nil, fmt.Errorf("Open: %w", err)
	}
	return f, nil
}

// sourceImportPath separates the import identity from the user-facing scan path.
func (c *Core) sourceImportPath(disk FS, name string) string {
	rooted, ok := disk.(interface{ RootPath() string })
	if !ok {
		return ""
	}
	full := filepath.Join(rooted.RootPath(), filepath.FromSlash(name))
	var best string
	for _, root := range c.importRoots {
		relative, err := filepath.Rel(root, full)
		if err != nil || !filepath.IsLocal(relative) {
			continue
		}
		if best == "" || len(relative) < len(best) {
			best = relative
		}
	}
	return filepath.ToSlash(best)
}

func (c *Core) openDiskSource(disk FS, name string) (io.ReadCloser, error) {
	if rooted, ok := disk.(interface{ RootPath() string }); ok && c.sourceFileOpen != nil {
		full := filepath.Join(rooted.RootPath(), filepath.FromSlash(name))
		if c.importFileAllowed != nil && !c.importFileAllowed(full) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		file, err := c.sourceFileOpen(full)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			return file, err
		}
	}
	return disk.Open(name)
}
