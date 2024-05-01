package lint

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

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

		got, err := protoparser.Parse(f)
		if err != nil {
			return fmt.Errorf("protoparser.Parse: %w", err)
		}

		proto, err := unordered.InterpretProto(got)
		if err != nil {
			return fmt.Errorf("unordered.InterpretProto: %w", err)
		}

		// TODO:  read all imported proto files
		c.readImportedFiles(ctx, disk, proto)

		for i := range c.rules {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			results := c.rules[i].Validate(ProtoInfo{
				Path: path,
				Info: proto,
			})
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

// readImportedFiles reads all files that imported from scanning file
func (c *Lint) readImportedFiles(ctx context.Context, disk fs.FS, scanProto *unordered.Proto) error {
	for _, imp := range scanProto.ProtoBody.Imports {
		importedFile, err := c.readImportedFile(ctx, disk, strings.Trim(imp.Location, "\""))
		if err != nil {
			return fmt.Errorf("readImportedFile: %w", err)
		}

		_ = importedFile
	}

	return nil
}

func (c *Lint) readImportedFile(ctx context.Context, disk fs.FS, importName string) (*unordered.Proto, error) {
	// first try to read it locally
	f, err := disk.Open(importName)
	if err != nil {
		// not locally import
		modulePath, err := c.moduleReflect.GetModulePath(ctx, importName)
		_ = modulePath
		_ = err
	}

	_ = f
	_ = err

	return nil, nil
}
