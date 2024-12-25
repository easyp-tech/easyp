package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

func (c *Core) BreakingCheck(ctx context.Context, workingDir, path string) ([]IssueInfo, error) {
	return nil, nil
}

func (c *Core) readCurrentProtoFiles(ctx context.Context, workingDir, path string) ([]ProtoInfo, error) {
	protoFiles := make([]ProtoInfo, 0)

	fsWalker := fs.NewFSWalker(os.DirFS(workingDir), path)

	err := fsWalker.WalkDir(func(path string, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case filepath.Ext(path) == ".proto":
			return nil
		}

		protoInfo, err := c.protoInfoRead(ctx, fsWalker, path)
		if err != nil {
			return fmt.Errorf("c.protoInfoRead: %w", err)
		}

		protoFiles = append(protoFiles, protoInfo)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs.WalkDir: %w", err)
	}

	return protoFiles, nil
}
