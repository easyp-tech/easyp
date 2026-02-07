package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/go-protoparser/v4/parser"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
)

// Lint lints the proto file.
func (c *Core) Lint(ctx context.Context, fsWalker DirWalker) ([]IssueInfo, error) {
	var res []IssueInfo

	err := fsWalker.WalkDir(func(path string, err error) error {
		switch {
		case err != nil:
			return err
		case ctx.Err() != nil:
			return ctx.Err()
		case path_helpers.IsIgnoredPath(path, c.ignore):
			return nil
		case filepath.Ext(path) != ".proto":
			return nil
		}

		protoInfo, err := c.protoInfoRead(ctx, fsWalker, path)
		if err != nil {
			return fmt.Errorf("c.protoInfoRead: %w", err)
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
		c.logger.Debug(
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

// CheckIsIgnored check if passed breakingCheckRuleName has to be ignored due to ignore command in comments
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
