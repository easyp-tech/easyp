package lint

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
)

// Lint lints the proto file.
func (c *Lint) Lint(ctx context.Context, disk fs.FS) error {
	var res []error

	err := fs.WalkDir(disk, ".", func(path string, d fs.DirEntry, err error) error {
		slog.Info("path", "path", path, "d", d)

		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case d.IsDir():
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		}

		path = filepath.Join(c.rootPath, path)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("os.Open: %w", err)
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
