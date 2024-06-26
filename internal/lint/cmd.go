package lint

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

// Lint lints the proto file.
func (c *Lint) Lint(ctx context.Context, disk fs.FS) error {
	var res []error

	err := fs.WalkDir(disk, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			if slices.Contains(c.ignoreDirs, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		}

		f, err := disk.Open(path)
		if err != nil {
			return fmt.Errorf("disk.Open: %w", err)
		}
		defer func() {
			err := f.Close()
			if err != nil {
				// TODO: Handle error
			}
		}()

		proto, err := readProtoFile(f)
		if err != nil {
			return fmt.Errorf("readProtoFile: %w", err)
		}

		protoFilesFromImport, err := c.readFilesFromImport(ctx, disk, proto)
		if err != nil {
			return fmt.Errorf("readFilesFromImport: %w", err)
		}

		protoInfo := ProtoInfo{
			Path:                 path,
			Info:                 proto,
			ProtoFilesFromImport: protoFilesFromImport,
		}

		for i := range c.rules {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			results := c.rules[i].Validate(protoInfo)
			for _, result := range results {
				res = append(res, fmt.Errorf("%s:%w", path, result))
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("fs.WalkDir: %w", err)
	}

	return errors.Join(res...)
}

// readFilesFromImport reads all files that imported from scanning file
func (c *Lint) readFilesFromImport(
	ctx context.Context, disk fs.FS, scanProto *unordered.Proto,
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

func (c *Lint) readFileFromImport(ctx context.Context, disk fs.FS, importName string) (*unordered.Proto, error) {
	// first try to read it locally
	f, err := disk.Open(importName)
	if err == nil {
		// locally import
		defer func() {
			_ = f.Close()
		}()

		proto, err := readProtoFile(f)
		if err != nil {
			return nil, fmt.Errorf("readProtoFile: %w", err)
		}
		return proto, nil
	}

	for _, dep := range c.deps {
		modulePath, err := c.moduleReflect.GetModulePath(ctx, dep)
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
		defer func() {
			_ = f.Close()
		}()

		proto, err := readProtoFile(f)
		if err != nil {
			return nil, fmt.Errorf("readProtoFile: %w", err)
		}

		return proto, nil
	}

	// try to reda from standrt
	stdPath := "/usr/include"
	fullPath := filepath.Join(stdPath, importName)
	f, err = os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}

		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	proto, err := readProtoFile(f)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: %w", err)
	}

	//return nil, fmt.Errorf("file %s not found", importName)
	return proto, nil
}

func readProtoFile(f fs.File) (*unordered.Proto, error) {
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
