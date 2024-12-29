package core

import (
	"context"
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
		return ProtoInfo{}, fmt.Errorf("fs.Open: %w", err)
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
		return nil, fmt.Errorf("protoparser.Parse: %w", err)
	}

	proto, err := unordered.InterpretProto(got)
	if err != nil {
		return nil, fmt.Errorf("unordered.InterpretProto: %w", err)
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
	// first try to read it locally
	f, err := disk.Open(importName)
	if err == nil {
		// locally import
		defer c.close(ctx, f, importName)

		proto, err := readProtoFile(f)
		if err != nil {
			return nil, fmt.Errorf("readProtoFile: %w, path: %s", err, importName)
		}
		return proto, nil
	}

	for _, dep := range c.deps {
		modulePath, err := c.getModulePath(ctx, dep)
		if err != nil {
			return nil, fmt.Errorf("c.moduleReflect.GetModulePath: %w", err)
		}

		fullPath := filepath.Join(modulePath, importName)
		f, err = os.Open(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("os.Open: %w", err)
		}
		defer c.close(ctx, f, fullPath)

		proto, err := readProtoFile(f)
		if err != nil {
			return nil, fmt.Errorf("readProtoFile: %w, path: %s", err, importName)
		}

		return proto, nil
	}

	f, err = wellknownimports.Content.Open(importName)
	if err != nil {
		if os.IsNotExist(err) {
			//return nil, ErrOpenImportFile
			return nil, &OpenImportFileError{FileName: importName}
		}

		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer c.close(ctx, f, importName)

	proto, err := readProtoFile(f)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: %w, path: %s", err, importName)
	}

	return proto, nil
}
