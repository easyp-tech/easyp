package rules_test

// The batch parity harness (04-RESEARCH "Batch-Verification Mechanism"): it
// drives the live old engine and the dormant linked engine in-process over the
// same compiling fixture tree and compares sorted (rule, path, line, col)
// multisets. Every later batch plan reuses requirePorted as its RED guard and
// assertBatchParity as its only comparison mechanism; corpus lint-diff stays
// the Phase-7 FINISH gate.

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
// both engines and requires equal fingerprints. wantOldCount is the expected
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

	newRules, _, err := rules.NewLinked(cfg)
	require.NoError(t, err)
	newApp := core.New(core.Options{
		Logger:              logger.NewNop(),
		LinkedRules:         newRules,
		AllowCommentIgnores: true,
		KnownLintRules:      rules.AllRuleNames(),
	})
	newIssues, err := newApp.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
	require.NoError(t, err)

	oldFp := fingerprint(oldIssues)
	require.Len(t, oldFp, wantOldCount, "old-side finding count (asserted before the multiset compare)")

	require.Equal(t, oldFp, fingerprint(newIssues),
		"old-vs-new parity: sorted (rule, path, line, col) multisets must match")
}

// fingerprint copies and sorts issues by the D-01 key (path, rule, line,
// column): the old engine's output order is nondeterministic (map iteration),
// so both sides sort before require.Equal. Messages are not part of the
// parity contract (rule, file, line, col is).
func fingerprint(infos []core.IssueInfo) []core.IssueInfo {
	out := make([]core.IssueInfo, len(infos))
	copy(out, infos)
	slices.SortStableFunc(out, func(a, b core.IssueInfo) int {
		return cmp.Or(
			strings.Compare(a.Path, b.Path),
			strings.Compare(a.RuleName, b.RuleName),
			cmp.Compare(a.Position.Line, b.Position.Line),
			cmp.Compare(a.Position.Column, b.Position.Column),
		)
	})
	return out
}
