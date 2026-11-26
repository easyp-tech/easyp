package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/easyp-tech/easyprotoc/protoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/easyp-tech/easyp/internal/sourceview"
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

func (m *serialCollectModuleRule) Validate(ProtoInfo) ([]Issue, error) {
	return nil, nil
}

func (m *serialCollectModuleRule) Collect(info ProtoInfo) error {
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

// TestLintMatrix is the parallel-correctness matrix for the linked lint
// engine: repeated-run byte determinism and the shared-instance statelessness
// audit under -race (ROADMAP criteria 1 and 4), the serial walk-order Collect
// lifecycle, ctx cancellation propagation, walk-order error precedence, and
// the COMPILE-class contract driven through the engine one level above the
// Phase 2 linkSources cells (mixed-corpus isolation, dep diagnostics,
// suppression bypass per research A6, compile-order determinism).
func TestLintMatrix(t *testing.T) {
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
			got, err := app.Lint(t.Context(), walker)
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
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
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
		wantOrder := slices.Sorted(maps.Keys(files))

		for run := range 2 {
			module := &serialCollectModuleRule{}
			app := newLinkedEngine(nil, false, nil, module)
			_, err := app.Lint(t.Context(), walker)
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
		got, err := app.Lint(ctx, fs.NewFSWalker(root, "."))

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
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
		require.Error(t, err)
		require.Nil(t, got)

		// a_bad.proto sorts first in walk order; the indexed drain reports
		// its error under the main-wins wrapper, never z_bad.proto's.
		require.ErrorContains(t, err, "fs.WalkDir: c.lintProtoInfoRead:")
		assert.Contains(t, err.Error(), "a_bad.proto")
		assert.NotContains(t, err.Error(), "z_bad.proto")
	})

	t.Run("mixed_corpus_compile_isolation", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"broken.proto":  "syntax = \"proto3\";\npackage broken;\nmessage bad_name {\n  required string s = 1;\n}\n",
			"clean_a.proto": "syntax = \"proto3\";\n\npackage fine;\n\nmessage FineOne {\n  string name = 1;\n}\n\nmessage FineTwo {\n  string id = 1;\n}\n",
			"clean_b.proto": "syntax = \"proto3\";\n\npackage other;\n\nmessage FineThree {\n  string id = 1;\n}\n",
		})
		// Doubles eligible for every clean declaration: if rules ran anywhere
		// on this corpus they would report these messages.
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{}, &doubleModuleRule{})
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// COMPILE issues isolate to the erroneous file (exact count, exact
		// path, etalon shape), and rules are skipped corpus-wide: zero double
		// findings anywhere, like buf (Phase 2 multi_file_corpus pin, one
		// level up through the engine).
		require.Len(t, got, 1)
		assert.Equal(t, CompileRuleName, got[0].RuleName)
		assert.Equal(t, "broken.proto", got[0].Path)
		assert.Equal(t, "Required fields are not allowed in proto3.", got[0].Message)
		assert.Equal(t, 4, got[0].Position.Line)
		assert.Equal(t, 3, got[0].Position.Column)
		assert.Empty(t, got[0].Position.Filename)
		assert.Zero(t, got[0].Position.Offset)
		assert.Empty(t, got[0].SourceName)
	})

	t.Run("dep_file_diagnostic_through_engine", func(t *testing.T) {
		t.Parallel()

		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n",
			"b.proto": "syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  Ghost field = 1;\n}\n",
		})
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{}, &doubleRuleBeta{})
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// The dep file's semantic error surfaces as a COMPILE issue carried
		// on the dep's own import path, with the verbatim compiler message
		// naming the undefined type (Phase 2 dep_file_semantic_error pin,
		// mirrored one level up through the engine).
		require.Len(t, got, 1)
		assert.Equal(t, CompileRuleName, got[0].RuleName)
		assert.Equal(t, "b.proto", got[0].Path)
		assert.Contains(t, got[0].Message, "Ghost")
	})

	t.Run("compile_bypasses_suppression", func(t *testing.T) {
		t.Parallel()

		// A6/IN-02 behavior pin: COMPILE issues return before any suppression
		// machinery runs — no shouldIgnore, no ignoreOnly, no comment
		// suppression range filters them. Whether COMPILE remarks SHOULD be
		// suppressible is the Phase 5 IN-02 decision (deferred to the flip
		// plan); this cell makes the current routing explicit so the flip
		// plan inherits the decision consciously, not accidentally.
		root := writeCorpus(t, map[string]string{
			"broken.proto": "syntax = \"proto3\";\npackage broken;\n\n// easyp:disable DOUBLE_RULE_ALPHA\nmessage bad_name {\n  required string s = 1;\n}\n",
			"clean.proto":  "syntax = \"proto3\";\n\npackage fine;\n\nmessage Fine {\n  string name = 1;\n}\n",
		})
		// Every suppression channel armed: comment ignores allowed, the
		// directive annotates the offending declaration, ignoreOnly names the
		// erroneous path, and the double would report the clean file if rules
		// ran at all.
		app := newLinkedEngine(
			[]string{"DOUBLE_RULE_ALPHA"},
			true,
			map[string][]string{"DOUBLE_RULE_ALPHA": {"broken.proto"}},
			&doubleRuleAlpha{},
		)
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// The COMPILE issue returns unchanged through every armed channel.
		require.Len(t, got, 1)
		assert.Equal(t, CompileRuleName, got[0].RuleName)
		assert.Equal(t, "broken.proto", got[0].Path)
		assert.Equal(t, "Required fields are not allowed in proto3.", got[0].Message)
	})

	t.Run("compile_deterministic_twice", func(t *testing.T) {
		t.Parallel()

		// Two erroneous files, the first carrying two diagnostics: three
		// COMPILE issues total, ordered by the Phase 2 D-03 sort (path, line,
		// col) composed with the engine's D-01 sort at exit.
		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage bad_name {\n  required string s = 1;\n  required string t = 2;\n}\n",
			"z.proto": "syntax = \"proto3\";\npackage zeta;\nmessage also_bad {\n  required string s = 1;\n}\n",
		})
		app := newLinkedEngine(nil, false, nil, &doubleRuleAlpha{})
		walker := fs.NewFSWalker(root, ".")

		first, err := app.Lint(t.Context(), walker)
		require.NoError(t, err)
		second, err := app.Lint(t.Context(), walker)
		require.NoError(t, err)

		// Byte-identical across runs, with the sort stability holding through
		// both layers: a.proto's two diagnostics stay in line order ahead of
		// z.proto's.
		require.Equal(t, first, second)
		require.Len(t, first, 3)
		wantOrder := []struct {
			path string
			line int
		}{
			{path: "a.proto", line: 4},
			{path: "a.proto", line: 5},
			{path: "z.proto", line: 4},
		}
		for i, w := range wantOrder {
			assert.Equal(t, w.path, first[i].Path, "finding %d", i)
			assert.Equal(t, CompileRuleName, first[i].RuleName, "finding %d", i)
			assert.Equal(t, w.line, first[i].Position.Line, "finding %d", i)
			assert.Equal(t, 3, first[i].Position.Column, "finding %d", i)
		}
	})
}

// TestLintErrorPathFidelity pins the linked engine's error-path
// classification (Phase 5 WR-01/WR-02/IN-01 fixes plus the IN-02
// keep-decision): a real import open/read failure must degrade the run
// exactly like the old per-file engine — a command error carrying the OS
// cause, never a misleading "not found"-shaped COMPILE remark — the linker's
// corpus branch must honor the same importFileAllowed policy openDiskSource
// applies on disk reads, cancellation must be honored at the resolver
// boundary like every sibling resolver, and COMPILE diagnostics stay outside
// the suppression machinery.
func TestLintErrorPathFidelity(t *testing.T) {
	t.Parallel()

	t.Run("unreadable_import_falls_back", func(t *testing.T) {
		t.Parallel()

		// WR-01: an import whose open fails for a non-notFound OS reason must
		// force the classifier fallback, so the run behaves like the old
		// per-file engine (whose readFilesFromImport errors through the same
		// openImportFile funnels) and the OS cause surfaces as a command
		// error — never as a misleading "Import ... was not found or had
		// errors." COMPILE remark masking the real failure.
		//
		// Fixture note (plan re-derivation against live source): the plan's
		// directory fixture cannot drive this path — FSWalker.Open refuses
		// non-regular files with ErrUnsupported before any read happens, and a
		// directory named dep.proto is itself walked as a corpus file, so
		// pass-1 readErrs skip linking before the resolver ever runs.
		// Permission bits are not root-safe in CI. An *os.PathError carrier
		// injected through Options.OpenSourceFile is deterministic, root-safe,
		// and exercises the identical openDiskSource funnel an OS open
		// failure would.
		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage alpha;\n\nimport \"dep.proto\";\n\nmessage A {\n  dep.B b = 1;\n}\n",
		})
		app := New(Options{
			Logger: logger.NewNop(),
			Rules:  []Rule{&doubleRuleAlpha{}},
			OpenSourceFile: func(p string) (io.ReadCloser, error) {
				if strings.HasSuffix(filepath.ToSlash(p), "/dep.proto") {
					return nil, &os.PathError{Op: "open", Path: p, Err: os.ErrPermission}
				}
				return os.Open(p)
			},
		})
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))

		// The fallback degrades nothing into COMPILE: the corpus falls back to
		// the per-file path, whose import read errors through the same
		// openImportFile funnel with the OS cause attached.
		require.Error(t, err)
		require.Nil(t, got)
		require.ErrorIs(t, err, os.ErrPermission)
		assert.Contains(t, err.Error(), "dep.proto")
	})

	t.Run("sentinel_import_failure_falls_back", func(t *testing.T) {
		t.Parallel()

		// WR-01 gap closure: the production open funnels the resolver actually
		// calls (SourceRoots.OpenSourceFile -> sourceview.OpenLocalSelected;
		// FSWalker.Open -> sourceview.OpenLocal) fail not only with
		// *os.PathError carriers (unreadable_import_falls_back) but also with
		// PathError-free sentinel shapes — a dangling source alias, a target
		// changed during open, a nested-repository crossing. Such a carrier
		// must force the classifier fallback exactly like the PathError
		// subset: the run degrades to the per-file path and the funnel cause
		// re-surfaces as a command error — never the misleading
		// "not found"-shaped COMPILE remark skipping every rule corpus-wide.
		//
		// Each carrier mirrors its production funnel shape exactly: the
		// OpenLocal wrap of SourceRoots.OpenSourceFile over the named
		// local.go shape, with no *os.PathError anywhere in the chain and the
		// not-exist cause of the dangling row interpolated as text only (a
		// wrapped fs.ErrNotExist would trigger the missing branch and prove
		// the wrong path; errors.Join discards nil elements, so the
		// production joins reduce to their sentinel-only shape when the
		// Close errors are nil).
		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage alpha;\n\nimport \"dep.proto\";\n\nmessage A {\n  dep.B b = 1;\n}\n",
		})

		tests := []struct {
			name     string
			carrier  error
			wantIs   []error
			wantText string
		}{
			{
				// local.go:33 + local.go:176 — a dangling source alias: the
				// OpenLocal wrap (SourceRoots.OpenSourceFile) over the
				// Resolve wrap over the dangling-alias message wrapping
				// ErrUnsupported; the not-exist cause rides the message as
				// text only.
				name: "dangling_alias_unsupported",
				carrier: fmt.Errorf("OpenLocal: %w", errors.Join(fmt.Errorf("Resolve: %w",
					fmt.Errorf("source alias %q is dangling: %v: %w", "dep.proto",
						&os.PathError{Op: "lstat", Path: "dep.proto", Err: os.ErrNotExist},
						sourceview.ErrUnsupported)))),
				wantIs:   []error{sourceview.ErrUnsupported},
				wantText: "dep.proto",
			},
			{
				// local.go:56 — the target was replaced between resolution
				// and opening: the OpenLocal wrap over the ErrChanged join.
				name:    "changed_during_open",
				carrier: fmt.Errorf("OpenLocal: %w", errors.Join(sourceview.ErrChanged)),
				wantIs:  []error{sourceview.ErrChanged},
			},
			{
				// local.go:209 — a nested-repository crossing: the OpenLocal
				// wrap over the nested-repository message wrapping the
				// ErrUnsupported+ErrNestedRepository join.
				name: "nested_repository_crossing",
				carrier: fmt.Errorf("OpenLocal: %w", fmt.Errorf(
					"source target %q crosses nested repository %q: %w",
					"dep.proto", "dep.repo",
					errors.Join(sourceview.ErrUnsupported, sourceview.ErrNestedRepository))),
				wantIs: []error{sourceview.ErrUnsupported, sourceview.ErrNestedRepository},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				app := New(Options{
					Logger: logger.NewNop(),
					Rules:  []Rule{&doubleRuleAlpha{}},
					OpenSourceFile: func(p string) (io.ReadCloser, error) {
						if strings.HasSuffix(filepath.ToSlash(p), "/dep.proto") {
							return nil, tt.carrier
						}
						return os.Open(p)
					},
				})
				got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))

				// The sentinel-shaped funnel failure degrades nothing into
				// COMPILE: the corpus falls back to the per-file path, whose
				// openImportFile wrap chain re-surfaces the same carrier as a
				// command error (REQ-2.4 parity semantics).
				require.Error(t, err)
				require.Nil(t, got)
				for _, want := range tt.wantIs {
					require.ErrorIs(t, err, want)
				}
				if tt.wantText != "" {
					assert.Contains(t, err.Error(), tt.wantText)
				}
			})
		}
	})

	t.Run("blocked_corpus_file_errors", func(t *testing.T) {
		t.Parallel()

		// WR-02: the linker's corpus branch (the linted-map hit) must honor
		// the same importFileAllowed policy openDiskSource applies on disk
		// reads — the rooted full path is computed exactly as openDiskSource
		// computes it — so a policy-excluded file can never be served from
		// memory. On denial the resolver stores missing and returns
		// protoc.ErrFileNotFound; the per-file fallback then yields
		// openDiskSource's exact os.ErrNotExist carrier.
		//
		// The denial itself is driven at the resolver boundary: pass 1 reads
		// corpus files through the same openDiskSource funnel, so an
		// engine-level run already errors before linking when OpenSourceFile
		// is set — the corpus-branch check is the defense in depth that keeps
		// the linker from ever serving a denied file the pass-1 funnel did not
		// consult.
		root := t.TempDir()
		files := map[string]string{
			"a.proto": "syntax = \"proto3\";\n\npackage alpha;\n\nmessage A {\n  string name = 1;\n}\n",
			"b.proto": "syntax = \"proto3\";\n\npackage beta;\n\nmessage B {\n  string id = 1;\n}\n",
		}
		for name, src := range files {
			require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(src), 0o644))
		}
		core := New(Options{
			Logger: logger.NewNop(),
			Rules:  []Rule{&doubleRuleAlpha{}},
			ImportFileAllowed: func(p string) bool {
				return !strings.HasSuffix(filepath.ToSlash(p), "/b.proto")
			},
			OpenSourceFile: func(p string) (io.ReadCloser, error) { return os.Open(p) },
		})
		walker := fs.NewFSWalker(root, ".")
		resolver := &lintResolver{
			core:   core,
			disk:   walker,
			linted: map[string][]byte{"b.proto": []byte(files["b.proto"])},
		}

		// The corpus branch denies the policy-excluded file: ErrFileNotFound,
		// no content, missing stored for the classifier.
		got, err := resolver.FindFileByPath(t.Context(), "b.proto")
		require.ErrorIs(t, err, protoc.ErrFileNotFound)
		require.Nil(t, got)
		require.True(t, resolver.missing.Load(), "denial must store missing for the classifier fallback")

		// An allowed corpus file is still served from memory unchanged.
		allowed := &lintResolver{
			core:   core,
			disk:   walker,
			linted: map[string][]byte{"a.proto": []byte(files["a.proto"])},
		}
		gotAllowed, errAllowed := allowed.FindFileByPath(t.Context(), "a.proto")
		require.NoError(t, errAllowed)
		require.Equal(t, files["a.proto"], string(gotAllowed))

		// The per-file funnel's denial carrier: openDiskSource answers the
		// same denial with the exact os.ErrNotExist PathError the old engine
		// surfaces (frozen-file behavior the fallback inherits).
		f, openErr := core.openDiskSource(walker, "b.proto")
		require.Error(t, openErr)
		require.ErrorIs(t, openErr, os.ErrNotExist)
		var pathErr *os.PathError
		require.ErrorAs(t, openErr, &pathErr)
		assert.Equal(t, "open", pathErr.Op)
		assert.Equal(t, "b.proto", pathErr.Path)
		require.Nil(t, f)

		// End to end, a policy-denied corpus file is a command error naming
		// the file, with no findings for it (pass-1 readSource consults the
		// same policy funnel).
		gotEngine, errEngine := core.Lint(t.Context(), walker)
		require.Error(t, errEngine)
		require.Nil(t, gotEngine)
		require.ErrorIs(t, errEngine, os.ErrNotExist)
		assert.Contains(t, errEngine.Error(), "b.proto")
	})

	t.Run("find_by_path_honors_ctx", func(t *testing.T) {
		t.Parallel()

		// IN-01: the ctx guard every sibling resolver carries
		// (github.com/easyp-tech/easyprotoc/protoc/resolver.go): a cancelled context returns
		// the context error before any file access, from the corpus branch and
		// the disk branch alike.
		root := writeCorpus(t, map[string]string{
			"a.proto": "syntax = \"proto3\";\npackage alpha;\nmessage A {}\n",
		})
		resolver := &lintResolver{
			core:   New(Options{Logger: logger.NewNop()}),
			disk:   fs.NewFSWalker(root, "."),
			linted: map[string][]byte{"cached.proto": []byte("cached content")},
		}
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		gotCached, errCached := resolver.FindFileByPath(ctx, "cached.proto")
		require.ErrorIs(t, errCached, context.Canceled)
		require.Nil(t, gotCached)

		gotDisk, errDisk := resolver.FindFileByPath(ctx, "a.proto")
		require.ErrorIs(t, errDisk, context.Canceled)
		require.Nil(t, gotDisk)
	})

	t.Run("ignored_broken_dependency_not_suppressed", func(t *testing.T) {
		t.Parallel()

		// IN-02 keep-decision pin (green by construction): a broken dependency
		// excluded from the lint corpus by Ignore still yields its COMPILE
		// diagnostics — compiler diagnostics are not rule findings and stay
		// outside the suppression machinery (the walk filter, ignoreOnly, and
		// comment directives). The broken dep carries the file's established
		// broken-file shape (proto3 required field), and the cell stays green
		// through the flip.
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "dep"), 0o755))
		files := map[string]string{
			"a.proto":          "syntax = \"proto3\";\n\npackage alpha;\n\nimport \"dep/broken.proto\";\n\nmessage A {\n  dep.B b = 1;\n}\n",
			"dep/broken.proto": "syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  required string s = 1;\n}\n",
		}
		for name, src := range files {
			require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(src), 0o644))
		}
		app := New(Options{
			Logger: logger.NewNop(),
			Rules:  []Rule{&doubleRuleAlpha{}},
			Ignore: []string{"dep"},
		})
		got, err := app.Lint(t.Context(), fs.NewFSWalker(root, "."))
		require.NoError(t, err)

		// The COMPILE diagnostic on the ignored, broken dependency is present
		// despite the ignore, and rules are skipped corpus-wide (zero double
		// findings) exactly as in the compile_bypasses_suppression cell.
		require.Len(t, got, 1)
		assert.Equal(t, CompileRuleName, got[0].RuleName)
		assert.Equal(t, "dep/broken.proto", got[0].Path)
	})
}
