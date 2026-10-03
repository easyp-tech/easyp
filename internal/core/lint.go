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
	c.logger.Info(ctx, "starting lint")

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

		suppressions, err := c.commentSuppressions(protoInfo)
		if err != nil {
			return fmt.Errorf("commentSuppressions: %w", err)
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
				if c.allowCommentIgnores && suppressions.contains(result.RuleName, result.Position.Line) {
					continue
				}
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

	c.logger.Info(ctx, "lint completed", slog.Int("issues", len(res)))

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
		c.logger.Warn(
			ctx,
			"failed to close file",
			slog.Any("error", err),
			slog.String("path", path),
		)
	}
}

// CheckIsIgnored recognizes exact rule tokens in declaration comments.
// Per-policy enablement and block scopes belong to commentSuppressions.
func CheckIsIgnored(comments []*parser.Comment, ruleName string) bool {
	for _, comment := range comments {
		for _, line := range comment.Lines() {
			action, rest, ok := parseDirectivePrefix(strings.TrimSpace(line))
			if !ok || action == "enable" {
				continue
			}
			for _, name := range strings.FieldsFunc(rest, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }) {
				if name == ruleName {
					return true
				}
			}
		}
	}
	return false
}
