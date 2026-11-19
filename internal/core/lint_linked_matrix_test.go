package core

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
)

// serialCollectModuleRule appends to an unsynchronized slice on every Collect:
// the engine runs Collect serially in the walk-order drain, so the append is
// safe without a lock — a concurrent Collect would trip the race detector on
// this exact slice (D-03 serialization proof).
type serialCollectModuleRule struct {
	collected []string
}

var _ ModuleRule = (*serialCollectModuleRule)(nil)

func (m *serialCollectModuleRule) Message() string { return "serial collect finding" }

func (m *serialCollectModuleRule) Validate(LinkedFileInfo) ([]Issue, error) {
	return nil, nil
}

func (m *serialCollectModuleRule) Collect(info LinkedFileInfo) error {
	m.collected = append(m.collected, info.Path)
	return nil
}

func (m *serialCollectModuleRule) Finish() []IssueInfo { return nil }

// matrixCorpus builds an 8-file single-package corpus: m0.proto declares the
// shared SharedBase message, m1..m7 import it and reference it, so every
// concurrent window worker resolves the m0 outline through the shared
// linkedFiles adapter. Message names stay unique across the shared package:
// proto files of one package share one namespace.
func matrixCorpus() map[string]string {
	files := map[string]string{
		"m0.proto": "syntax = \"proto3\";\n\npackage matrixpkg;\n\nmessage SharedBase {\n  string id = 1;\n}\n\nmessage M0Two {\n  string name = 1;\n}\n",
	}
	for i := 1; i < 8; i++ {
		files[fmt.Sprintf("m%d.proto", i)] = fmt.Sprintf("syntax = \"proto3\";\n\npackage matrixpkg;\n\nimport \"m0.proto\";\n\nmessage M%dOne {\n  SharedBase base = 1;\n}\n\nmessage M%dTwo {\n  string x = 1;\n}\n", i, i)
	}
	return files
}

// TestLintLinkedMatrix is the parallel-correctness matrix for the linked lint
// engine: repeated-run byte determinism and the shared-instance statelessness
// audit under -race (ROADMAP criteria 1 and 4), the serial walk-order Collect
// lifecycle, ctx cancellation propagation, walk-order error precedence, and
// the COMPILE-class contract driven through the engine one level above the
// Phase 2 linkSources cells (mixed-corpus isolation, dep diagnostics,
// suppression bypass per research A6, compile-order determinism).
func TestLintLinkedMatrix(t *testing.T) {
	t.Parallel()

	t.Run("repeated_runs_byte_identical", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, matrixCorpus())
		// The double instances are shared across all five runs: repeated runs
		// must not depend on rule state or cache warmth.
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{}, &doubleModuleRule{})
		walker := fs.NewFSWalker(root, ".")

		var runs [][]IssueInfo
		for range 5 {
			got, err := app.LintLinked(t.Context(), walker)
			require.NoError(t, err)
			runs = append(runs, got)
		}

		// Exact count first (8 files x 2 messages x 2 rules + 1 module Finish
		// finding = 33), then full-slice equality: run[i] == run[0] byte for
		// byte (D-01 total sort over the indexed drain, ROADMAP criterion 1).
		require.Len(t, runs[0], 33)
		for i := 1; i < len(runs); i++ {
			require.Equal(t, runs[0], runs[i], "run %d differs from run 0", i)
		}
	})

	t.Run("race_shared_doubles_identity", func(t *testing.T) {
		t.Parallel()

		files := matrixCorpus()
		root := writeCorpus(t, files)
		// One shared instance per rule across the whole corpus: every
		// concurrent window worker validates through the same stateless
		// doubles, and the m0 outline is resolved concurrently through the
		// shared linkedFiles adapter (ROADMAP criterion 4 statelessness
		// audit; a stateful rule would trip the detector here).
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{}, &doubleModuleRule{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Exact totals, never range bounds: 8 files x 2 messages x 2 rules +
		// 1 module Finish finding = 33.
		require.Len(t, got, 33)

		// Message declaration lines: m0.proto declares at lines 5 and 9; the
		// other files' import line shifts the layout to 7 and 11.
		messageLines := map[string][]int{"m0.proto": {5, 9}}
		for i := 1; i < 8; i++ {
			messageLines[fmt.Sprintf("m%d.proto", i)] = []int{7, 11}
		}

		// Identity tallies: exact per-file rule counts and per-file line
		// counts. A finding whose Path, RuleName, or Line does not match the
		// file that declares its message cannot pass these exact maps — no
		// cross-file bleed, no lost or duplicated findings.
		perFileRules := make(map[string]map[string]int, len(files))
		perFileLines := make(map[string]map[int]int, len(files))
		moduleCount := 0
		for _, issue := range got {
			switch issue.RuleName {
			case "DOUBLE_MODULE_RULE":
				moduleCount++
			case "DOUBLE_RULE_ALPHA", "DOUBLE_RULE_BETA":
				rules := perFileRules[issue.Path]
				if rules == nil {
					rules = make(map[string]int)
					perFileRules[issue.Path] = rules
				}
				rules[issue.RuleName]++
				lines := perFileLines[issue.Path]
				if lines == nil {
					lines = make(map[int]int)
					perFileLines[issue.Path] = lines
				}
				lines[issue.Position.Line]++
			default:
				t.Fatalf("unexpected rule name in output: %q", issue.RuleName)
			}
		}
		assert.Equal(t, 1, moduleCount)
		require.Len(t, perFileRules, len(files))
		for name := range files {
			assert.Equal(t, map[string]int{"DOUBLE_RULE_ALPHA": 2, "DOUBLE_RULE_BETA": 2}, perFileRules[name], "file %s", name)
			wantLines := make(map[int]int, len(messageLines[name]))
			for _, line := range messageLines[name] {
				wantLines[line] = 2 // one finding per rule per message
			}
			assert.Equal(t, wantLines, perFileLines[name], "file %s finding lines", name)
		}
	})

	t.Run("collect_serial_walk_order", func(t *testing.T) {
		t.Parallel()

		files := matrixCorpus()
		root := writeCorpus(t, files)
		walker := fs.NewFSWalker(root, ".")

		// Walk order of the corpus: the walker lists the .proto files
		// lexically, and the engine passes them to Collect in that order.
		wantOrder := make([]string, 0, len(files))
		for _, name := range slices.Sorted(maps.Keys(files)) {
			wantOrder = append(wantOrder, name)
		}

		for run := range 2 {
			module := &serialCollectModuleRule{}
			app := newLinkedEngine(nil, false, nil, module)
			_, err := app.LintLinked(t.Context(), walker)
			require.NoError(t, err)

			// The recorded Collect sequence equals the walk order exactly,
			// once per run: the drain-loop placement serializes Collect —
			// completion order would be nondeterministic here.
			assert.Equal(t, wantOrder, module.collected, "run %d", run)
		}
	})

	t.Run("ctx_cancel_surfaces", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, matrixCorpus())
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		got, err := app.LintLinked(ctx, fs.NewFSWalker(root, "."))

		// Cancellation propagates out of the engine, never swallowed: the
		// walk loop and both parallel.Run sites return ctx.Err() under %w
		// wraps, so errors.Is sees the sentinel through the chain.
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, got)
	})

	t.Run("walk_order_error_precedence", func(t *testing.T) {
		t.Parallel()

		// Two syntax-error files: only the first in walk order may win the
		// error race — completion order would be nondeterministic under
		// parallel parsing, so this cell pins the indexed drain (main-wins
		// error parity).
		root := writeCorpus(t, map[string]string{
			"a_bad.proto": "syntax = \"proto3\";\n\npackage broken;\n\nmessage Truncated {\n  string name = 1;\n",
			"z_bad.proto": "syntax = \"proto3\";\n\npackage zeta;\n\nmessage Zed {\n  string x = 1;\n",
		})
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.Error(t, err)
		require.Nil(t, got)

		// a_bad.proto sorts first in walk order; the indexed drain reports
		// its error under the main-wins wrapper, never z_bad.proto's.
		require.ErrorContains(t, err, "fs.WalkDir: c.lintProtoInfoRead:")
		assert.Contains(t, err.Error(), "a_bad.proto")
		assert.NotContains(t, err.Error(), "z_bad.proto")
	})
}
