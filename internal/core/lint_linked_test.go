package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

// Engine test doubles for the linked lint engine: the per-file rules are
// stateless and share one instance across the whole corpus; the ModuleRule
// double records its lifecycle calls. Struct names drive LinkedRuleName, so
// IgnoreOnly keys and directive names use the UPPER_SNAKE spelling of these
// names (doubleRuleAlpha -> DOUBLE_RULE_ALPHA).

var (
	_ LintRule   = (*doubleRuleAlpha)(nil)
	_ LintRule   = (*doubleRuleBeta)(nil)
	_ ModuleRule = (*doubleModuleRule)(nil)
)

// doubleRuleAlpha reports one finding per message declaration.
type doubleRuleAlpha struct{}

func (d *doubleRuleAlpha) Message() string { return "alpha double finding" }

func (d *doubleRuleAlpha) Validate(info LinkedFileInfo) ([]Issue, error) {
	return messageIssues(info, d.Message(), "DOUBLE_RULE_ALPHA"), nil
}

// doubleRuleBeta reports under a rule name that sorts after alpha's; tests
// register it first so unsorted (registration-order) output would differ.
type doubleRuleBeta struct{}

func (d *doubleRuleBeta) Message() string { return "beta double finding" }

func (d *doubleRuleBeta) Validate(info LinkedFileInfo) ([]Issue, error) {
	return messageIssues(info, d.Message(), "DOUBLE_RULE_BETA"), nil
}

// messageIssues builds one finding per message declaration at its position.
func messageIssues(info LinkedFileInfo, message, ruleName string) []Issue {
	issues := make([]Issue, 0, len(info.File.Messages))
	for _, msg := range info.File.Messages {
		issues = append(issues, Issue{
			Position: msg.Pos,
			Message:  message,
			RuleName: ruleName,
		})
	}
	return issues
}

// doubleModuleRule is a ModuleRule double: Validate counts invocations (the
// engine must never call it), Collect records walk-order paths, Finish reports
// one cross-file finding anchored on the first collected path.
type doubleModuleRule struct {
	mu        sync.Mutex
	collected []string
	validates atomic.Int64
}

func (m *doubleModuleRule) Message() string { return "module double finding" }

func (m *doubleModuleRule) Validate(LinkedFileInfo) ([]Issue, error) {
	m.validates.Add(1)
	return nil, nil
}

func (m *doubleModuleRule) Collect(info LinkedFileInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collected = append(m.collected, info.Path)
	return nil
}

func (m *doubleModuleRule) Finish() []IssueInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.collected) == 0 {
		return nil
	}
	return []IssueInfo{{
		Path: m.collected[0],
		Issue: Issue{
			Position: Position{Line: 1, Column: 1},
			Message:  m.Message(),
			RuleName: "DOUBLE_MODULE_RULE",
		},
	}}
}

// writeCorpus writes the given corpus files into a fresh temp dir and returns
// its root.
func writeCorpus(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, src := range files {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(src), 0o644))
	}
	return root
}

// newLinkedEngine constructs the linked engine through the public Options
// surface. KnownLintRules feeds the suppression directive validation, so cells
// using directives pass their double rule names explicitly.
func newLinkedEngine(known []string, allowCommentIgnores bool, ignoreOnly map[string][]string, rules ...LintRule) *Core {
	return New(Options{
		Logger:              logger.NewNop(),
		LinkedRules:         rules,
		KnownLintRules:      known,
		AllowCommentIgnores: allowCommentIgnores,
		IgnoreOnly:          ignoreOnly,
	})
}

// TestLintLinked drives the dormant two-pass linked engine end to end with
// test-double rules (research A2): corpus walk -> parallel read -> link ->
// windowed model build from linked trees -> central suppression filter ->
// indexed serial drain in walk order -> ModuleRule.Finish -> D-01 total sort.
func TestLintLinked(t *testing.T) {
	t.Parallel()

	t.Run("linked_corpus_deterministic", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage alpha;\n\nmessage AlphaOne {\n  string name = 1;\n}\n\nmessage AlphaTwo {\n  string id = 1;\n}\n",
			"b.proto": "syntax = \"proto3\";\n\npackage beta;\n\nmessage BetaOne {\n  string name = 1;\n}\n",
			"c.proto": "syntax = \"proto3\";\n\npackage gamma;\n\nimport \"b.proto\";\n\nmessage GammaOne {\n  b.BetaOne ref = 1;\n}\n",
		})
		alpha := &doubleRuleAlpha{}
		beta := &doubleRuleBeta{}
		module := &doubleModuleRule{}
		app := newLinkedEngine(
			[]string{"DOUBLE_RULE_ALPHA", "DOUBLE_RULE_BETA", "DOUBLE_MODULE_RULE"},
			false, nil, alpha, beta, module,
		)
		walker := fs.NewFSWalker(root, ".")

		first, err := app.LintLinked(t.Context(), walker)
		require.NoError(t, err)
		second, err := app.LintLinked(t.Context(), walker)
		require.NoError(t, err)

		// a.proto carries two messages, b.proto and c.proto one each: 2 rules
		// x 4 messages + one module finding = 9. Two runs over the same core
		// must return byte-identical slices (D-01 determinism, ROADMAP
		// criterion 1).
		require.Len(t, first, 9)
		require.Equal(t, first, second)

		var moduleIssues []IssueInfo
		for _, issue := range first {
			if issue.RuleName == "DOUBLE_MODULE_RULE" {
				moduleIssues = append(moduleIssues, issue)
			}
		}
		require.Len(t, moduleIssues, 1)
		assert.Equal(t, "a.proto", moduleIssues[0].Path)
	})

	t.Run("engine_sort_order", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"z.proto": "syntax = \"proto3\";\npackage zeta;\nmessage Zed {\n  string name = 1;\n}\n",
			"a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage Ace {\n  string name = 1;\n}\n",
		})
		// Registration order is deliberately anti-alphabetical: without the
		// D-01 total sort the walk-order drain would emit BETA before ALPHA
		// inside every file.
		app := newLinkedEngine(nil, false, nil, &doubleRuleBeta{}, &doubleRuleAlpha{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Both messages sit on line 3 of their file; the total order is
		// (path, rule, line, column).
		require.Len(t, got, 4)
		type orderKey struct {
			path string
			rule string
		}
		want := []orderKey{
			{"a.proto", "DOUBLE_RULE_ALPHA"},
			{"a.proto", "DOUBLE_RULE_BETA"},
			{"z.proto", "DOUBLE_RULE_ALPHA"},
			{"z.proto", "DOUBLE_RULE_BETA"},
		}
		for i, w := range want {
			assert.Equal(t, w.path, got[i].Path, "finding %d", i)
			assert.Equal(t, w.rule, got[i].RuleName, "finding %d", i)
			assert.Equal(t, 3, got[i].Position.Line, "finding %d", i)
		}
	})

	t.Run("compile_error_rules_skipped", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"test.proto": "syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n",
		})
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Like buf, rules do not run on files that do not compile: exactly the
		// etalon-shaped COMPILE issue comes back (link_test.go compile_error
		// precedent), before any suppression machinery (research A6).
		require.Len(t, got, 1)
		assert.Equal(t, CompileRuleName, got[0].RuleName)
		assert.Equal(t, "test.proto", got[0].Path)
		assert.Equal(t, "Required fields are not allowed in proto3.", got[0].Message)
		assert.Equal(t, 4, got[0].Position.Line)
		assert.Equal(t, 3, got[0].Position.Column)
	})

	t.Run("comment_suppression_central", func(t *testing.T) {
		t.Parallel()

		const single = "syntax = \"proto3\";\n\npackage alpha;\n\n// easyp:disable DOUBLE_RULE_ALPHA\nmessage SuppressedOne {\n  string name = 1;\n}\n\nmessage KeepMe {\n  string name = 1;\n}\n"

		for _, tc := range []struct {
			name      string
			allow     bool
			wantLines []int
		}{
			{name: "allow_filters_suppressed", allow: true, wantLines: []int{10}},
			{name: "disallow_keeps_findings", allow: false, wantLines: []int{6, 10}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				root := writeCorpus(t, map[string]string{"single.proto": single})
				app := newLinkedEngine([]string{"DOUBLE_RULE_ALPHA"}, tc.allow, nil, &doubleRuleAlpha{})
				got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
				require.NoError(t, err)

				// The disable comment annotates SuppressedOne (line 6), so the
				// central filter drops that file's line-6 finding; KeepMe
				// (line 10) stays. Both sides of AllowCommentIgnores (D-04).
				require.Len(t, got, len(tc.wantLines))
				for i, line := range tc.wantLines {
					assert.Equal(t, "single.proto", got[i].Path)
					assert.Equal(t, "DOUBLE_RULE_ALPHA", got[i].RuleName)
					assert.Equal(t, line, got[i].Position.Line)
				}
			})
		}
	})

	t.Run("module_rule_lifecycle", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage A {}\n",
			"b.proto": "syntax = \"proto3\";\npackage beta;\nmessage B {}\n",
			"c.proto": "syntax = \"proto3\";\npackage gamma;\nmessage C {}\n",
		})
		module := &doubleModuleRule{}
		app := newLinkedEngine(nil, false, nil, module)
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Validate is never called on a ModuleRule (D-03).
		assert.Zero(t, module.validates.Load())
		// Collect saw every corpus file exactly once, in walk order (serial
		// drain, D-03/D-05).
		module.mu.Lock()
		collected := append([]string(nil), module.collected...)
		module.mu.Unlock()
		assert.Equal(t, []string{"a.proto", "b.proto", "c.proto"}, collected)
		// The Finish finding passed the shouldIgnore/suppression re-filter and
		// landed in the output.
		require.Len(t, got, 1)
		assert.Equal(t, "a.proto", got[0].Path)
		assert.Equal(t, "DOUBLE_MODULE_RULE", got[0].RuleName)
	})

	t.Run("ignore_only_skips_per_file", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage A {}\n",
			"b.proto": "syntax = \"proto3\";\npackage beta;\nmessage B {}\n",
		})
		app := newLinkedEngine(nil, false, map[string][]string{"DOUBLE_RULE_ALPHA": {"b.proto"}}, &doubleRuleAlpha{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// The b.proto finding is gone; the a.proto finding stays.
		require.Len(t, got, 1)
		assert.Equal(t, "a.proto", got[0].Path)
		assert.Equal(t, "DOUBLE_RULE_ALPHA", got[0].RuleName)
	})

	t.Run("race_multi_file", func(t *testing.T) {
		t.Parallel()

		files := map[string]string{
			"f0.proto": "syntax = \"proto3\";\n\npackage racepkg;\n\nmessage SharedBase {\n  string id = 1;\n}\n\nmessage Second {\n  string name = 1;\n}\n",
		}
		for i := 1; i < 8; i++ {
			files[fmt.Sprintf("f%d.proto", i)] = "syntax = \"proto3\";\n\npackage racepkg;\n\nimport \"f0.proto\";\n\nmessage FiOne {\n  SharedBase base = 1;\n}\n\nmessage FiTwo {\n  string x = 1;\n}\n"
		}
		root := writeCorpus(t, files)
		// Shared instances across the whole corpus: every worker checks files
		// through the same stateless rules, and every f1..f7 model resolves
		// the f0 outline through the shared linkedFiles adapter (T-03-02
		// mutex + sync.Once discipline under -race).
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{}, &doubleModuleRule{})
		got, err := app.LintLinked(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// Exact counts, never range bounds: 8 files x 2 messages x 2 rules +
		// 1 module finding = 33.
		require.Len(t, got, 33)
		perFile := make(map[string]map[string]int)
		moduleCount := 0
		for _, issue := range got {
			switch issue.RuleName {
			case "DOUBLE_MODULE_RULE":
				moduleCount++
			case "DOUBLE_RULE_ALPHA", "DOUBLE_RULE_BETA":
				counts := perFile[issue.Path]
				if counts == nil {
					counts = make(map[string]int)
					perFile[issue.Path] = counts
				}
				counts[issue.RuleName]++
			default:
				t.Fatalf("unexpected rule name in output: %q", issue.RuleName)
			}
		}
		assert.Equal(t, 1, moduleCount)
		require.Len(t, perFile, 8)
		for name := range files {
			assert.Equal(t, map[string]int{"DOUBLE_RULE_ALPHA": 2, "DOUBLE_RULE_BETA": 2}, perFile[name], "file %s", name)
		}
	})
}
