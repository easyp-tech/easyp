package core

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/easyp-tech/easyp/internal/core/path_helpers"
	"github.com/easyp-tech/easyp/internal/parallel"
	"github.com/easyp-tech/easyp/internal/protoast"
)

// lintWindow bounds how many parsed files are held at once: files are read
// and parsed in parallel a window at a time, then checked in walk order.
const lintWindow = 256

// Lint lints the proto files with the two-pass linked engine: pass 1
// reads every file once and links the corpus, pass 2 builds each file's model
// from the linked syntax trees and checks it, in windows; rules that look at
// files alone run concurrently, cross-file rules run serially in walk order.
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
	// When linking cannot classify the corpus, files are parsed on their own
	// so that syntax errors and missing imports surface as before.
	var linked *linkedFiles
	if !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) && len(paths) > 0 {
		linkedRaw, compileIssues := c.linkSources(ctx, fsWalker, paths, sources)
		if linkedRaw != nil {
			linked = &linkedFiles{linked: linkedRaw, outlines: make(map[string]*linkedOutline)}
		} else if err := c.parseOnTheirOwn(ctx, fsWalker, paths, sources, readErrs); err != nil {
			return nil, fmt.Errorf("fs.WalkDir: %w", err)
		} else if len(compileIssues) > 0 && !slices.ContainsFunc(readErrs, func(err error) bool { return err != nil }) {
			// Like buf, lint rules do not run on files that do not compile;
			// the COMPILE issues return before any suppression machinery runs
			// (research A6).
			c.logger.Info(ctx, "lint completed", slog.Int("issues", len(compileIssues)))
			return compileIssues, nil
		}
	}

	var res []IssueInfo
	suppressed := make(map[string]suppressionRanges, len(paths))
	infos := make([]ProtoInfo, min(lintWindow, len(paths)))
	infoErrs := make([]error, len(infos))
	checked := make([]checkedFile, len(infos))
	for start := 0; start < len(paths); start += lintWindow {
		window := paths[start:min(start+lintWindow, len(paths))]
		// Pass 2 builds models and resolves imports a window at a time. Rules
		// keep cross-file state, so they run below in the original order.
		err := parallel.Run(ctx, len(window), 0, func(i int) error {
			infos[i], infoErrs[i] = ProtoInfo{}, nil
			if readErrs[start+i] != nil {
				infoErrs[i] = readErrs[start+i]
				return nil
			}
			if linked != nil {
				infos[i], infoErrs[i] = c.protoInfoFromLink(fsWalker, linked, window[i])
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
			// The wrapper text pins main's user-visible error chain (REQ-2.4
			// main-wins philosophy): the callee name c.lintProtoInfoRead in the
			// format string is a deliberate parity pin, not the actual callee.
			if infoErrs[i] != nil {
				return nil, fmt.Errorf("fs.WalkDir: c.lintProtoInfoRead: %w", infoErrs[i])
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

	// Deterministic issue order: by path, then rule, then line, then column,
	// then source. This total sort is the one deliberate delta vs the etalon,
	// which leaves the drain order unsorted (D-01); SourceName is
	// user-visible in both output printers, so the sort stays total through
	// it. Message stays out: Message() equality is separately proven by the
	// carried per-rule tests.
	slices.SortStableFunc(res, func(a, b IssueInfo) int {
		return cmp.Or(
			strings.Compare(a.Path, b.Path),
			strings.Compare(a.RuleName, b.RuleName),
			cmp.Compare(a.Position.Line, b.Position.Line),
			cmp.Compare(a.Position.Column, b.Position.Column),
			strings.Compare(a.SourceName, b.SourceName),
		)
	})

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
func (c *Core) checkFile(ctx context.Context, path string, info ProtoInfo) checkedFile {
	suppressions, err := c.commentSuppressions(info)
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
		results, err := rule.Validate(info)
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

// collectFile hands the file to module rules, in walk order — except files
// ignored for that rule, which main never lets reach Validate and therefore
// never join the group (REQ-2.5 parity; see 04-REVIEW CR-01). The Finish
// drain's path filter below stays as-is: defense in depth for group rules,
// and the per-file RPC_REQUEST_RESPONSE_UNIQUE findings.
func (c *Core) collectFile(info ProtoInfo) error {
	for _, rule := range c.rules {
		if moduleRule, ok := rule.(ModuleRule); ok {
			if c.shouldIgnore(rule, info.Path) {
				continue
			}
			if err := moduleRule.Collect(info); err != nil {
				return fmt.Errorf("rule.Collect: %w", err)
			}
		}
	}
	return nil
}

// shouldIgnore reports whether a rule is ignore_only-suppressed on a path:
// an exact file match or a directory prefix match.
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
