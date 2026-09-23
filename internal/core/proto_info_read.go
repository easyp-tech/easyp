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

func (c *Core) protoInfoRead(ctx context.Context, fs FS, path string) (ProtoInfo, error) {
	f, err := fs.Open(path)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("Open: %w", err)
	}
	defer c.close(ctx, f, path)

	protoFile, err := readProtoFile(f)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("readProtoFile: %w", err)
	}

	protoFilesFromImport, err := c.readFilesFromImport(ctx, fs, protoFile)
	if err != nil {
		return ProtoInfo{}, fmt.Errorf("readFilesFromImport: %w", err)
	}

	protoInfo := ProtoInfo{
		Path:                 path,
		Info:                 protoFile,
		ProtoFilesFromImport: protoFilesFromImport,
	}
	return protoInfo, nil
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
	f, err := disk.Open(importName)
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
		f, err := os.Open(fullPath)
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
