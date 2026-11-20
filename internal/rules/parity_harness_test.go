package rules_test

// The batch parity harness (04-RESEARCH "Batch-Verification Mechanism"): it
// drives the live old engine and the dormant linked engine in-process over the
// same compiling fixture tree and compares sorted (path, rule, line, col,
// source) multisets. Both engines run identical suppression state:
// assertBatchParity feeds NewLinked's unwrapped ignoreOnly into the linked
// app's core.Options.IgnoreOnly (04-REVIEW WR-01), so ignore_only parity is
// testable by construction. Every later batch plan reuses requirePorted as its
// RED guard and assertBatchParity as its only comparison mechanism; corpus
// lint-diff stays the Phase-7 FINISH gate.

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/rules"
)

// requirePorted is every batch's RED guard: a batch's parity test fails HERE,
// on its own assertion, listing the unported names, while the package still
// compiles (a Go build-failure RED is invalid — never reference
// not-yet-existing types in RED tests). Every given name is probed so one
// failure reports the batch's full unported list.
func requirePorted(t *testing.T, names ...string) {
	t.Helper()

	var unported []string
	for _, name := range names {
		if _, _, err := rules.NewLinked(config.LintConfig{Use: []string{name}}); err != nil {
			unported = append(unported, fmt.Sprintf("%s (%v)", name, err))
		}
	}
	if len(unported) > 0 {
		t.Fatalf("rules not ported to the linked registry yet: %s", strings.Join(unported, "; "))
	}
}

// assertBatchParity runs one config over one compiling fixture tree through
// both engines — with identical ignore_only state: NewLinked's second return
// value feeds core.Options.IgnoreOnly on the linked app, so suppression parity
// is testable by construction (04-REVIEW WR-01) — and requires equal
// fingerprints. wantOldCount is the expected
// old-side finding count, asserted BEFORE the multiset compare so a vacuous
// both-empty run can never pass silently (T-04-02); fixtures must compile —
// on broken files the engines legitimately diverge (COMPILE routing), and
// both engines' errors surface here verbatim.
func assertBatchParity(t *testing.T, cfg config.LintConfig, files map[string]string, wantOldCount int) {
	t.Helper()

	root := t.TempDir()
	for name, src := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(src), 0o644))
	}

	oldRules, oldIgnoreOnly, err := rules.New(cfg)
	require.NoError(t, err)
	oldApp := core.New(core.Options{
		Logger:              logger.NewNop(),
		Rules:               oldRules,
		IgnoreOnly:          oldIgnoreOnly,
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	oldIssues, err := oldApp.Lint(t.Context(), fs.NewFSWalker(root, "."))
	require.NoError(t, err)

	newRules, newIgnoreOnly, err := rules.NewLinked(cfg)
	require.NoError(t, err)
	newApp := core.New(core.Options{
		Logger:              logger.NewNop(),
		LinkedRules:         newRules,
		IgnoreOnly:          newIgnoreOnly, // parity must include suppression state (04-REVIEW WR-01)
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	newIssues, err := newApp.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
	require.NoError(t, err)

	oldFp := fingerprint(oldIssues)
	require.Len(t, oldFp, wantOldCount, "old-side finding count (asserted before the multiset compare)")

	require.Equal(t, oldFp, fingerprint(newIssues),
		"old-vs-new parity: sorted (path, rule, line, col, source) multisets must match")
}

// fingerprint copies and sorts issues by the D-01 key (path, rule, line,
// column, source): the old engine's output order is nondeterministic (map
// iteration), so both sides sort before require.Equal. The parity contract is
// (path, rule, line, col, source) — SourceName is user-visible in both output
// printers (internal/api/lint.go text and JSON); Messages are not part of the
// contract because Message() equality is separately proven by the carried
// per-rule tests.
func fingerprint(infos []core.IssueInfo) []core.IssueInfo {
	out := make([]core.IssueInfo, len(infos))
	copy(out, infos)
	slices.SortStableFunc(out, func(a, b core.IssueInfo) int {
		return cmp.Or(
			strings.Compare(a.Path, b.Path),
			strings.Compare(a.RuleName, b.RuleName),
			cmp.Compare(a.Position.Line, b.Position.Line),
			cmp.Compare(a.Position.Column, b.Position.Column),
			strings.Compare(a.SourceName, b.SourceName),
		)
	})
	return out
}

// TestFingerprintSourceNameTiebreak pins the fifth sort term (04-REVIEW
// WR-04): issues tying on (path, rule, line, col) but differing only in
// SourceName must come out ordered by SourceName, and SourceName is part of
// the compared value — it is user-visible in both output printers
// (internal/api/lint.go text and JSON).
func TestFingerprintSourceNameTiebreak(t *testing.T) {
	t.Parallel()

	alpha := core.IssueInfo{
		Issue: core.Issue{
			Position:   core.Position{Line: 5, Column: 1},
			SourceName: "alpha",
			RuleName:   "RULE_ONE",
		},
		Path: "a/b.proto",
	}
	beta := core.IssueInfo{
		Issue: core.Issue{
			Position:   core.Position{Line: 5, Column: 1},
			SourceName: "beta",
			RuleName:   "RULE_ONE",
		},
		Path: "a/b.proto",
	}

	// Sort pin: fed in reverse order, the fingerprint must still come out
	// ordered by SourceName (RED until the fifth cmp.Or term lands).
	got := fingerprint([]core.IssueInfo{beta, alpha})
	require.Len(t, got, 2)
	require.Equal(t, "alpha", got[0].SourceName,
		"ties on (path, rule, line, col) must order by SourceName")
	require.Equal(t, "beta", got[1].SourceName)

	// Value-compare pin: two one-element fingerprints differing only in
	// SourceName are NotEqual.
	require.NotEqual(t, fingerprint([]core.IssueInfo{alpha}), fingerprint([]core.IssueInfo{beta}),
		"SourceName is part of the parity value")
}
