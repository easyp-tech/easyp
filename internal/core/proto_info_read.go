package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/wellknownimports"
)

// breakingProtoInfoRead is the breaking-owned read-model constructor (D-06):
// the breaking path's own twin of lintProtoInfoRead. Behaviorally identical in
// Phase 1; only the read-model type differs.
func (c *Core) breakingProtoInfoRead(ctx context.Context, fs FS, path string) (BreakingProtoInfo, error) {
	f, err := c.openDiskSource(fs, path)
	if err != nil {
		return BreakingProtoInfo{}, fmt.Errorf("Open: %w", err)
	}
	defer c.close(ctx, f, path)

	protoFile, comments, err := readProtoFileWithDirectives(f)
	if err != nil {
		return BreakingProtoInfo{}, fmt.Errorf("readProtoFileWithDirectives: %w", err)
	}

	protoFilesFromImport, err := c.readFilesFromImport(ctx, fs, protoFile)
	if err != nil {
		return BreakingProtoInfo{}, fmt.Errorf("readFilesFromImport: %w", err)
	}

	breakingInfo := BreakingProtoInfo{
		directiveComments:    comments,
		Path:                 path,
		Info:                 protoFile,
		ProtoFilesFromImport: protoFilesFromImport,
	}
	breakingInfo.ImportPath = c.sourceImportPath(fs, path)
	return breakingInfo, nil
}

// BreakingProtoInfo is the breaking-owned read model for one proto file (D-06).
type BreakingProtoInfo struct {
	directiveComments    []directiveComment
	Path                 string
	ImportPath           string
	Info                 *unordered.Proto
	ProtoFilesFromImport map[ImportPath]*unordered.Proto
}

func readProtoFile(f io.Reader) (*unordered.Proto, error) {
	got, err := protoparser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("Parse: %w", err)
	}

	proto, err := unordered.InterpretProto(got)
	if err != nil {
		return nil, fmt.Errorf("InterpretProto: %w", err)
	}

	return proto, nil
}

// readFilesFromImport reads all files that imported from scanning file
func (c *Core) readFilesFromImport(
	ctx context.Context, disk FS, scanProto *unordered.Proto,
) (map[ImportPath]*unordered.Proto, error) {
	protoFilesFromImport := make(map[ImportPath]*unordered.Proto, len(scanProto.ProtoBody.Imports))

	for _, imp := range scanProto.ProtoBody.Imports {
		importPath := ConvertImportPath(imp.Location)
		fileFromImport, err := c.readFileFromImport(ctx, disk, string(importPath))
		if err != nil {
			return nil, fmt.Errorf("readFileFromImport: %w", err)
		}

		protoFilesFromImport[importPath] = fileFromImport
	}

	return protoFilesFromImport, nil
}

func (c *Core) readFileFromImport(ctx context.Context, disk FS, importName string) (*unordered.Proto, error) {
	f, err := c.openImportFile(disk, importName)
	if err != nil {
		return nil, fmt.Errorf("openImportFile: %w", err)
	}
	defer c.close(ctx, f, importName)

	proto, err := readProtoFile(f)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: %w", &os.PathError{Op: "parse", Path: importName, Err: err})
	}
	return proto, nil
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
