package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/parallel"
	"github.com/easyp-tech/easyp/internal/protoast"
)

// Lint lints the proto file.
// lintWindow bounds how many parsed files are held at once: files are read
// and parsed in parallel a window at a time, then checked in walk order.
const lintWindow = 256

func (c *Core) Lint(ctx context.Context, fsWalker DirWalker) ([]IssueInfo, error) {
	c.logger.Info(ctx, "starting lint")

	var paths []string
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
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs.WalkDir: %w", err)
	}

	if c.importCache == nil {
		c.importCache = &parsedImports{files: map[string]*protoast.File{}}
	}

	// Pass 1 reads every linted file once. Errors are kept per file and
	// reported in walk order below, as a sequential run would.
	sources := make([][]byte, len(paths))
	readErrs := make([]error, len(paths))
	err = parallel.Run(ctx, len(paths), 0, func(i int) error {
		sources[i], readErrs[i] = c.readSource(ctx, fsWalker, paths[i])
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fs.WalkDir: %w", err)
	}

	// Linking parses every file once; models are built from its syntax trees.
	// When linking fails, files are parsed on their own so that syntax errors
	// and missing imports surface as before.
	var linked *linkedFiles
	if !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) && len(paths) > 0 {
		var compileIssues []IssueInfo
		linked, compileIssues = c.linkSources(ctx, fsWalker, paths, sources)
		if linked == nil {
			if err := c.parseOnTheirOwn(ctx, fsWalker, paths, sources, readErrs); err != nil {
				return nil, fmt.Errorf("fs.WalkDir: %w", err)
			}
			if len(compileIssues) > 0 && !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) {
				// Like buf, lint rules do not run on files that do not compile.
				c.logger.Info(ctx, "lint completed", slog.Int("issues", len(compileIssues)))
				return compileIssues, nil
			}
		}
	}

	var res []IssueInfo
	suppressed := make(map[string]suppressionRanges, len(paths))
	infos := make([]ProtoInfo, min(lintWindow, len(paths)))
	infoErrs := make([]error, len(infos))
	checked := make([]checkedFile, len(infos))
	for start := 0; start < len(paths); start += lintWindow {
		window := paths[start:min(start+lintWindow, len(paths))]
		// Pass 2 builds models and resolves imports a window at a time. Rules keep
		// cross-file state, so they run below in the original order.
		err := parallel.Run(ctx, len(window), 0, func(i int) error {
			infos[i], infoErrs[i] = ProtoInfo{}, nil
			if readErrs[start+i] != nil {
				infoErrs[i] = readErrs[start+i]
				return nil
			}
			if linked != nil {
				infos[i], infoErrs[i] = c.protoInfoFromLinked(fsWalker, linked, window[i])
			} else {
				infos[i], infoErrs[i] = c.protoInfoFromSource(ctx, fsWalker, window[i], sources[start+i])
			}
			sources[start+i] = nil
			if infoErrs[i] == nil {
				checked[i] = c.checkFile(ctx, window[i], infos[i])
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("fs.WalkDir: %w", err)
		}
		for i, path := range window {
			if infoErrs[i] != nil {
				return nil, fmt.Errorf("fs.WalkDir: c.protoInfoRead: %w", infoErrs[i])
			}
			if checked[i].err != nil {
				return nil, fmt.Errorf("fs.WalkDir: %w", checked[i].err)
			}
			if err := c.collectFile(infos[i]); err != nil {
				return nil, fmt.Errorf("fs.WalkDir: %w", err)
			}
			suppressed[path] = checked[i].suppressions
			res = append(res, checked[i].issues...)
			infos[i], checked[i] = ProtoInfo{}, checkedFile{}
		}
	}

	for _, rule := range c.rules {
		moduleRule, ok := rule.(ModuleRule)
		if !ok {
			continue
		}
		for _, issue := range moduleRule.Finish() {
			if c.shouldIgnore(rule, issue.Path) {
				continue
			}
			if c.allowCommentIgnores && suppressed[issue.Path].contains(issue.RuleName, issue.Position.Line) {
				continue
			}
			res = append(res, issue)
		}
	}

	c.logger.Info(ctx, "lint completed", slog.Int("issues", len(res)))

	return res, nil
}

// checkedFile is one file's findings from rules that look at it alone.
type checkedFile struct {
	issues       []IssueInfo
	suppressions suppressionRanges
	err          error
}

// checkFile applies suppressions and the rules that look at one file at a
// time. Those rules keep no state between files, so files can be checked
// concurrently.
func (c *Core) checkFile(ctx context.Context, path string, protoInfo ProtoInfo) checkedFile {
	suppressions, err := c.commentSuppressions(protoInfo)
	if err != nil {
		return checkedFile{err: fmt.Errorf("commentSuppressions: %w", err)}
	}
	var res []IssueInfo
	for _, rule := range c.rules {
		if ctx.Err() != nil {
			return checkedFile{err: ctx.Err()}
		}
		if _, ok := rule.(ModuleRule); ok || c.shouldIgnore(rule, path) {
			continue
		}
		results, err := rule.Validate(protoInfo)
		if err != nil {
			return checkedFile{err: fmt.Errorf("rule.Validate: %w", err)}
		}
		for _, result := range results {
			if c.allowCommentIgnores && suppressions.contains(result.RuleName, result.Position.Line) {
				continue
			}
			res = append(res, IssueInfo{Issue: result, Path: path})
		}
	}
	return checkedFile{issues: res, suppressions: suppressions}
}

// collectFile hands the file to module rules, in walk order. They see every
// file, so a group stays complete even when findings for some of its files
// are ignored.
func (c *Core) collectFile(protoInfo ProtoInfo) error {
	for _, rule := range c.rules {
		if moduleRule, ok := rule.(ModuleRule); ok {
			if err := moduleRule.Collect(protoInfo); err != nil {
				return fmt.Errorf("rule.Collect: %w", err)
			}
		}
	}
	return nil
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

// parseOnTheirOwn parses each linted file without linking, recording syntax
// errors per file and registering outlines for imports of linted files.
func (c *Core) parseOnTheirOwn(ctx context.Context, disk FS, paths []string, sources [][]byte, errs []error) error {
	return parallel.Run(ctx, len(paths), 0, func(i int) error {
		file, err := protoast.Parse(paths[i], sources[i])
		if err != nil {
			errs[i] = fmt.Errorf("Parse: %w", err)
			return nil
		}
		c.cacheLintedOutline(disk, paths[i], file)
		return nil
	})
}
