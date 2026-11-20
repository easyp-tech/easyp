package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/easyp-tech/easyp/internal/parallel"
	"github.com/easyp-tech/easyp/internal/protoast"
)

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

// protoInfoFromSource builds the model of a file parsed on its own, without
// the linked set. It is the fallback of protoInfoFromLink.
func (c *Core) protoInfoFromSource(ctx context.Context, fs FS, path string, src []byte) (ProtoInfo, error) {
	file, err := protoast.Parse(path, src)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("Parse: %w", err)
	}

	imported, err := c.linkedFilesFromImport(ctx, fs, file)
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

// linkedFilesFromImport parses the direct imports of file. The per-import wrap
// text pins the etalon's user-visible error chain (see the rename map in
// lint_model.go).
func (c *Core) linkedFilesFromImport(ctx context.Context, disk FS, file *protoast.File) (map[ImportPath]*protoast.File, error) {
	imported := make(map[ImportPath]*protoast.File, len(file.Imports))
	for _, imp := range file.Imports {
		importPath := ConvertImportPath(imp.Location)
		parsed, err := c.linkedFileFromImport(ctx, disk, string(importPath))
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

// linkedFileFromImport parses an imported file once per lint root and keeps
// its outline: lint rules only look up names declared by imported files.
func (c *Core) linkedFileFromImport(ctx context.Context, disk FS, importName string) (*protoast.File, error) {
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

// parseOnTheirOwn parses each linted file without linking, recording syntax
// errors per file and registering outlines for imports of linted files.
func (c *Core) parseOnTheirOwn(ctx context.Context, disk FS, paths []string, sources [][]byte, errs []error) error {
	return parallel.Run(ctx, len(paths), 0, func(i int) error {
		file, err := protoast.Parse(paths[i], sources[i])
		if err != nil {
			errs[i] = fmt.Errorf("Parse: %w", err)
			return nil
		}
		c.cacheLintedOutline(disk, paths[i], file)
		return nil
	})
}

// protoInfoFromLink builds a linted file's model from its linked tree, then
// drops the tree; its outline stays available to files importing it.
func (c *Core) protoInfoFromLink(disk FS, linked *linkedFiles, path string) (ProtoInfo, error) {
	file, ok := linked.linked.File(path)
	if !ok || file.Syntax == nil {
		return ProtoInfo{}, fmt.Errorf("Parse: %s was not linked", path)
	}
	model := protoast.Build(file.Syntax)
	linked.outline(path)
	file.Syntax = nil

	imported := make(map[ImportPath]*protoast.File, len(model.Imports))
	for _, imp := range model.Imports {
		outline, ok := linked.outline(imp.Path)
		if !ok {
			return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %s was not linked", imp.Path)
		}
		imported[ConvertImportPath(imp.Location)] = outline
	}
	return ProtoInfo{
		Path:          path,
		ImportPath:    c.sourceImportPath(disk, path),
		File:          model,
		ImportedFiles: imported,
		Resolved:      resolution(linked.linked, path),
	}, nil
}

func (c *Core) close(ctx context.Context, f io.Closer, path string) {
	err := f.Close()
	if err != nil {
		c.logger.Warn(
			ctx,
			"failed to close file",
			slog.Any("error", err),
			slog.String("path", path),
		)
	}
}
