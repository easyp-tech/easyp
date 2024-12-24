package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"

	wfs "github.com/easyp-tech/easyp/internal/fs"
	"github.com/easyp-tech/easyp/wellknownimports"
)

// Lint lints the proto file.
func (c *Core) Lint(ctx context.Context, fsWalker wfs.DirWalker) ([]IssueInfo, error) {
	var res []IssueInfo

	err := fsWalker.WalkDir(func(path string, fs wfs.FS, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case slices.Contains(c.ignore, path):
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		}

		f, err := fs.Open(path)
		if err != nil {
			return fmt.Errorf("disk.Open: %w", err)
		}
		defer c.close(ctx, f, path)

		proto, err := readProtoFile(f)
		if err != nil {
			return fmt.Errorf("readProtoFile: %w: path: %s", err, path)
		}

		protoFilesFromImport, err := c.readFilesFromImport(ctx, fs, proto)
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

			if c.shouldIgnore(c.rules[i], path) {
				continue
			}

			results, err := c.rules[i].Validate(protoInfo)
			if err != nil {
				return fmt.Errorf("rule.Validate: %w", err)
			}

			for _, result := range results {
				res = append(res, IssueInfo{
					Issue: result,
					Path:  path,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs.WalkDir: %w", err)
	}

	return res, nil
}

// readFilesFromImport reads all files that imported from scanning file
func (c *Core) readFilesFromImport(
	ctx context.Context, disk wfs.FS, scanProto *unordered.Proto,
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

func (c *Core) readFileFromImport(ctx context.Context, disk wfs.FS, importName string) (*unordered.Proto, error) {
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
		defer c.close(ctx, f, fullPath) // TODO: fix it

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

func (c *Core) shouldIgnore(rule Rule, path string) bool {
	ruleName := GetRuleName(rule)
	ignoreFilesOrDirs := c.ignoreOnly[ruleName]

	for _, fileOrDir := range ignoreFilesOrDirs {
		switch {
		case fileOrDir == path:
			return true
		case strings.HasPrefix(path, fileOrDir):
			return true
		}
	}

	return false
}

func (c *Core) close(ctx context.Context, f io.Closer, path string) {
	err := f.Close()
	if err != nil {
		c.logger.DebugContext(
			ctx,
			"incorrect closing",
			slog.String(
				"err",
				err.Error(),
			),
			slog.String(
				"path",
				path,
			),
		)
	}
}

const (
	// for backward compatibility with buf
	bufLintIgnorePrefix = "buf:lint:ignore "
	lintIgnorePrefix    = "nolint:"
)

// NOTE: Try to not use global var
var allowCommentIgnores = true

// CheckIsIgnored check if passed ruleName has to be ignored due to ignore command in comments
func CheckIsIgnored(comments []*parser.Comment, ruleName string) bool {
	if !allowCommentIgnores {
		return false
	}

	if len(comments) == 0 {
		return false
	}

	bufIgnore := bufLintIgnorePrefix + ruleName
	easypIgnore := lintIgnorePrefix + ruleName

	for _, comment := range comments {
		if strings.Contains(comment.Raw, bufIgnore) {
			return true
		}
		if strings.Contains(comment.Raw, easypIgnore) {
			return true
		}
	}

	return false
}

func SetAllowCommentIgnores(val bool) {
	allowCommentIgnores = val
}
