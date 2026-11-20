// Vendored from easyp-tech/compiler-lab (private), branch main, directory
// easyp-cli/internal/rules/lint_helpers.go, pinned SHA
// 08cef10909c0722f7042ba17b5c11951f9c6b6f6, vendored 2026-11-18.
//
// D-01 keeps the whole lint helper set (целиком, как в лаборатории) so the
// future compiler sync stays trivial and the Phase-4 rule-port diff stays
// minimal. Sync procedure: internal/easypc/VENDORED.md.
//
// Phase-4 re-alignment RESOLVED (04-07): the file compiles against the
// linked-engine contract and the ports consume their helpers. Consumed by
// the *_linked.go ports: groupedValues.add,
// firstExemptMismatches (the 04-02 main-first-exempt Finish shape).
// D-01-kept verbatim and unused until Phase 5 deletes them together with the
// old rule path: fileOption (lost its last consumer to 04-08's per-occurrence
// option iteration — 04-REVIEW WR-02), optionString, enumNumber,
// leadingCommentDocuments, isLintDirective, packagePosition, mismatches,
// packageHasVersionSuffix, versionNumber. The .golangci.yml paths exclusion
// for this file stays for exactly those unused helpers; Phase 5 removes
// helpers and exclusion in one consolidation commit.
//
// Delta from upstream (D-01 refinement, user-confirmed 2026-11-17; settled
// state after the Phase-4 re-alignment):
//
//  1. packagePosition — the lab signature consumed the lab's protoast-based
//     core.ProtoInfo.File; it now takes the protoast file directly
//     (packagePosition(file *protoast.File, path string)); the body
//     references info.File -> file and info.Path -> path are mechanical.
//  2. groupedValue.pos — no textual change was needed: main's core.Position
//     is an alias of protoast.Position (settled by 01-03), so the lab
//     declaration compiles verbatim.
//  3. groupedValues.mismatches — the 01-04 seam adaptation (5-arg AppendIssue
//     with the meta.Position conversion and the parser/meta import) is GONE:
//     with the ports on the linked contract, the helper is back to the
//     lab-faithful shape — mismatches(rule core.LintRule) calls the linked
//     constructor (nil, rule, v.pos, v.source) verbatim. mismatches remains
//     the report-everyone shape and is not reached by the ports (the ports'
//     Finish is firstExemptMismatches); it stays vendored for re-sync
//     fidelity.

package rules

import (
	"strconv"
	"strings"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

// fileOption scans file.Options and hands back its final entry with the
// given name (a last-wins collapse — the body stays lab-verbatim). Kept
// unused since 04-08: the seven option ports iterate info.File.Options per
// occurrence (main's shape, 04-REVIEW WR-02), so no port consumes this
// helper anymore; Phase 5 deletes it together with the old path (D-01
// whole-file vendoring for lab re-syncs).
func fileOption(file *protoast.File, name string) *protoast.Option {
	var found *protoast.Option
	for _, opt := range file.Options {
		if opt.Name == name {
			found = opt
		}
	}
	return found
}

// optionString returns a literal's value: strings unquoted, other literals as written.
func optionString(literal string) string {
	if unquoted, err := strconv.Unquote(literal); err == nil {
		return unquoted
	}
	if len(literal) >= 2 && literal[0] == '\'' && literal[len(literal)-1] == '\'' {
		return literal[1 : len(literal)-1]
	}
	return literal
}

// enumNumber parses an enum value number as written: decimal, hex or octal, with sign.
func enumNumber(number string) (int64, bool) {
	n, err := strconv.ParseInt(number, 0, 64)
	return n, err == nil
}

// leadingCommentDocuments reports whether the comment directly above a
// declaration documents it: some line must be non-empty and not a lint
// directive. Detached comments, separated by a blank line, do not count.
func leadingCommentDocuments(decl *protoast.Decl) bool {
	for _, comment := range decl.Leading {
		for _, line := range comment.Lines() {
			line = strings.TrimSpace(line)
			if line == "" || isLintDirective(line) {
				continue
			}
			return true
		}
	}
	return false
}

func isLintDirective(line string) bool {
	for _, prefix := range []string{"buf:lint:ignore", "nolint:", "easyp:"} {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// packagePosition is where package findings go: the package statement, or the
// file itself when the file has no package.
func packagePosition(file *protoast.File, path string) (core.Position, string) {
	if len(file.Packages) == 0 {
		return protoast.FilePosition(path), ""
	}
	pkg := file.Packages[0]
	return pkg.Pos, pkg.Name
}

// groupedValues collects one value per file within groups such as packages or
// directories, and reports every file of a group whose files disagree.
type groupedValues struct {
	groups map[string][]groupedValue
	order  []string
}

type groupedValue struct {
	path, value, source string
	pos                 core.Position
}

func (g *groupedValues) add(group string, value groupedValue) {
	if g.groups == nil {
		g.groups = make(map[string][]groupedValue)
	}
	if _, ok := g.groups[group]; !ok {
		g.order = append(g.order, group)
	}
	g.groups[group] = append(g.groups[group], value)
}

// mismatches returns an issue for every file of each group with more than one
// distinct value, then resets the collector. Lab-faithful report-everyone
// shape on the linked contract (04-07 re-alignment): the ports' Finish is
// firstExemptMismatches below.
func (g *groupedValues) mismatches(rule core.LintRule) []core.IssueInfo {
	var res []core.IssueInfo
	for _, group := range g.order {
		values := g.groups[group]
		distinct := make(map[string]bool)
		for _, v := range values {
			distinct[v.value] = true
		}
		if len(distinct) < 2 {
			continue
		}
		for _, v := range values {
			for _, issue := range core.AppendLinkedIssue(nil, rule, v.pos, v.source) {
				res = append(res, core.IssueInfo{Issue: issue, Path: v.path})
			}
		}
	}
	*g = groupedValues{}
	return res
}

// firstExemptMismatches is the ported group rules' Finish with main's
// first-exempt semantics: the first value of a group (walk order; add
// preserves insertion order) is exempt, and only later files whose value
// differs from it issue, each at its own position. The lab's mismatches above
// stays the lab-faithful report-everyone shape and is not reached by the
// ported rules.
func (g *groupedValues) firstExemptMismatches(rule core.LintRule) []core.IssueInfo {
	var res []core.IssueInfo
	for _, group := range g.order {
		first := g.groups[group][0]
		for _, v := range g.groups[group][1:] {
			if v.value == first.value {
				continue
			}
			for _, issue := range core.AppendLinkedIssue(nil, rule, v.pos, v.source) {
				res = append(res, core.IssueInfo{Issue: issue, Path: v.path})
			}
		}
	}
	*g = groupedValues{}
	return res
}

// packageHasVersionSuffix applies buf's package version rule: at least two
// components, the last one v<major>, v<major>test..., or
// v<major>[p<patch>](alpha|beta)[<minor>], with major >= 1.
func packageHasVersionSuffix(pkg string) bool {
	parts := strings.Split(pkg, ".")
	if pkg == "" || len(parts) < 2 {
		return false
	}
	component := parts[len(parts)-1]
	if len(component) < 2 || component[0] != 'v' {
		return false
	}
	version := component[1:]
	if strings.Contains(version, "test") {
		major, _, _ := strings.Cut(version, "test")
		return versionNumber(major, 1)
	}
	alpha, beta := strings.Contains(version, "alpha"), strings.Contains(version, "beta")
	if alpha && beta {
		return false
	}
	if !alpha && !beta {
		return versionNumber(version, 1)
	}
	level := "alpha"
	if beta {
		level = "beta"
	}
	majorPatch, minor, _ := strings.Cut(version, level)
	if minor != "" && !versionNumber(minor, 1) {
		return false
	}
	if major, patch, ok := strings.Cut(majorPatch, "p"); ok {
		return versionNumber(major, 1) && versionNumber(patch, 1)
	}
	return versionNumber(majorPatch, 1)
}

func versionNumber(s string, minimum int64) bool {
	n, err := strconv.ParseInt(s, 10, 64)
	return err == nil && n >= minimum
}
