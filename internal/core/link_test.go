package core

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/easyp-tech/easyprotoc/protoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/fs/fs"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestLinkSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		paths   []string
		sources [][]byte
		setup   func(t *testing.T, root string)
		newCore func(t *testing.T, root string) *Core
		assert  func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo)
	}{
		{
			name:    "compile_error",
			paths:   []string{"test.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n")},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 1)
				assert.Equal(t, CompileRuleName, issues[0].RuleName)
				assert.Equal(t, "test.proto", issues[0].Path)
				assert.Equal(t, "Required fields are not allowed in proto3.", issues[0].Message)
				assert.Equal(t, 4, issues[0].Position.Line)
				assert.Equal(t, 3, issues[0].Position.Column)
				assert.Empty(t, issues[0].Position.Filename)
				assert.Zero(t, issues[0].Position.Offset)
				assert.Empty(t, issues[0].SourceName)
			},
		},
		{
			name:    "compiling_with_wkt",
			paths:   []string{"event.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"google/protobuf/timestamp.proto\";\n\nmessage Event {\n  google.protobuf.Timestamp at = 1;\n}\n")},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.NotNil(t, linked)
				assert.Empty(t, issues)
			},
		},
		{
			name:    "missing_import",
			paths:   []string{"orphan.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"missing.proto\";\n")},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)
			},
		},
		{
			name:  "dep_file_semantic_error",
			paths: []string{"a.proto", "b.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  Ghost field = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 1)
				assert.Contains(t, issues[0].Message, "Ghost")
				assert.Equal(t, "b.proto", issues[0].Path)
				assert.Equal(t, CompileRuleName, issues[0].RuleName)
			},
		},
		{
			name:  "deterministic_order_twice",
			paths: []string{"z.proto", "a.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\npackage zexample;\nmessage bad_name_z {\n  required string s = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\npackage aexample;\nmessage bad_name_a {\n  required string s = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 2)
				assert.Equal(t, "a.proto", issues[0].Path)
				assert.Equal(t, "z.proto", issues[1].Path)

				linkedAgain, issuesAgain := engine.linkSources(t.Context(), walker, []string{"z.proto", "a.proto"},
					[][]byte{
						[]byte("syntax = \"proto3\";\npackage zexample;\nmessage bad_name_z {\n  required string s = 1;\n}\n"),
						[]byte("syntax = \"proto3\";\npackage aexample;\nmessage bad_name_a {\n  required string s = 1;\n}\n"),
					})
				require.Nil(t, linkedAgain)
				require.Equal(t, issues, issuesAgain)
			},
		},
		{
			name:  "multi_file_corpus",
			paths: []string{"clean.proto", "broken.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage fine;\n\nmessage Fine {\n  string name = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\npackage broken;\nmessage bad_name {\n  required string s = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 1)
				assert.Equal(t, "broken.proto", issues[0].Path)
				assert.Equal(t, CompileRuleName, issues[0].RuleName)
			},
		},
		{
			name:  "syntax_error_falls_back",
			paths: []string{"clean.proto", "truncated.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage fine;\n\nmessage Fine {\n  string name = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\n\npackage broken;\n\nmessage Truncated {\n  string name = 1;\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)
			},
		},
		{
			name:  "syntax_beats_compile",
			paths: []string{"a.proto", "b.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage broken;\n\nmessage Truncated {\n  string name = 1;\n"),
				[]byte("syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  Ghost field = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)
			},
		},
		{
			name:  "semantic_only_emits_compile",
			paths: []string{"b.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  Ghost field = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.NotEmpty(t, issues)
				for _, issue := range issues {
					assert.Equal(t, CompileRuleName, issue.RuleName)
				}
			},
		},
		{
			name:  "transitive_missing_depth2",
			paths: []string{"a.proto", "b.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\n\npackage dep;\n\nimport \"missing.proto\";\n\nmessage B {\n  string name = 1;\n}\n"),
			},
			setup: func(t *testing.T, root string) {
				writeImportSource(t, root, "a.proto", "syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n")
				writeImportSource(t, root, "b.proto", "syntax = \"proto3\";\n\npackage dep;\n\nimport \"missing.proto\";\n\nmessage B {\n  string name = 1;\n}\n")
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)

				// The corpus fell back because the per-file path it lands on
				// carries the typed missing-import error to the exit-2 route.
				// The direct importer is b.proto: the per-file read path is
				// one-level, so processing b.proto is where missing.proto
				// surfaces.
				src, err := os.ReadFile(filepath.Join(walker.RootPath(), "b.proto"))
				require.NoError(t, err)
				_, err = engine.protoInfoFromSource(t.Context(), walker, "b.proto", src)
				require.Error(t, err)
				var notFound *OpenImportFileError
				require.ErrorAs(t, err, &notFound)
				assert.Equal(t, "missing.proto", notFound.FileName)
			},
		},
		{
			name:  "blocked_import",
			paths: []string{"a.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n"),
			},
			setup: func(t *testing.T, root string) {
				writeImportSource(t, root, "a.proto", "syntax = \"proto3\";\n\npackage example;\n\nimport \"b.proto\";\n\nmessage A {\n  dep.B ok = 1;\n}\n")
				writeImportSource(t, root, "b.proto", "syntax = \"proto3\";\n\npackage dep;\n\nmessage B {\n  string name = 1;\n}\n")
			},
			newCore: func(t *testing.T, root string) *Core {
				// The importRoots loop (proto_info_read.go:127-137) resolves
				// imports without consulting importFileAllowed, so this cell's
				// Core carries empty importRoots: the denial is the only thing
				// standing between b.proto and the os.ErrNotExist fall-through.
				return &Core{
					logger: logger.NewNop(),
					importFileAllowed: func(path string) bool {
						return path != filepath.Join(root, "b.proto")
					},
				}
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)

				// Blocked imports fall back exactly like missing ones, and the
				// per-file path errors with the exit-2 carrier naming the dep.
				src, err := os.ReadFile(filepath.Join(walker.RootPath(), "a.proto"))
				require.NoError(t, err)
				_, err = engine.protoInfoFromSource(t.Context(), walker, "a.proto", src)
				require.Error(t, err)
				var notFound *OpenImportFileError
				require.ErrorAs(t, err, &notFound)
				assert.Equal(t, "b.proto", notFound.FileName)
			},
		},
		{
			name:  "wkt_resolves",
			paths: []string{"event.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"google/protobuf/timestamp.proto\";\n\nmessage Event {\n  google.protobuf.Timestamp at = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.NotNil(t, linked)
				assert.Empty(t, issues)
			},
		},
		{
			name:  "non_local_import_compile",
			paths: []string{"outside_user.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"../outside.proto\";\n\nmessage A {\n  string name = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.NotEmpty(t, issues)

				// A non-local import path fails with a plain error, not
				// os.ErrNotExist, so the missing flag stays false and the
				// etalon classifies the corpus as COMPILE (research A4).
				found := false
				for _, issue := range issues {
					if issue.RuleName == CompileRuleName && strings.Contains(issue.Message, "../outside.proto") {
						found = true
					}
				}
				assert.True(t, found, "expected a COMPILE issue referencing the non-local import path, got: %+v", issues)
			},
		},
		{
			name:  "tab_utf8_position",
			paths: []string{"broken.proto"},
			sources: [][]byte{
				mustReadFixture(t, "testdata/link/tab_utf8/broken.proto"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 1)
				assert.Equal(t, CompileRuleName, issues[0].RuleName)
				assert.Contains(t, issues[0].Message, "Required fields are not allowed")

				// Compiler LineCol convention (github.com/easyp-tech/easyprotoc/lexer/token.go:121-123):
				// 1-based line, 1-based column, tab stop 8, byte columns. The
				// error token on line 6 follows a leading tab (col0 0 -> 8) and
				// 28 more bytes (`/* ` + 8 two-byte Cyrillic letters + ` ` + a
				// 4-byte emoji + ` */ `), so it starts at col0 36 -> column 37.
				// The Phase 1 D-07 rune-column contract would put the same
				// offset at column 19 — the fixture provably diverges, so the
				// two coordinate systems cannot be silently mixed.
				assert.Equal(t, 6, issues[0].Position.Line)
				assert.Equal(t, 37, issues[0].Position.Column)
			},
		},
		{
			name:  "import_cycle",
			paths: []string{"a.proto", "b.proto"},
			sources: [][]byte{
				[]byte("syntax = \"proto3\";\n\npackage cycle;\n\nimport \"b.proto\";\n\nmessage A {\n  B b = 1;\n}\n"),
				[]byte("syntax = \"proto3\";\n\npackage cycle;\n\nimport \"a.proto\";\n\nmessage B {\n  A a = 1;\n}\n"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.NotEmpty(t, issues)

				// The driver hard-codes the recursive-import diagnostic at
				// Line 2, Col 1 (github.com/easyp-tech/easyprotoc/internal/compiler/driver.go:319-323 at the go.mod pin); do not "fix" the 2:1.
				var cycle *IssueInfo
				for i := range issues {
					if strings.Contains(issues[i].Message, "recursively imports itself") {
						cycle = &issues[i]
					}
				}
				require.NotNil(t, cycle, "expected a cycle diagnostic, got: %+v", issues)
				assert.Equal(t, 2, cycle.Position.Line)
				assert.Equal(t, 1, cycle.Position.Column)
			},
		},
		{
			name: "multi_file_matrix_order",
			// Natural input order names the erroneous dep file first, so any
			// unsorted emission would list b.proto before a.proto.
			paths: []string{"b.proto", "a.proto"},
			sources: [][]byte{
				mustReadFixture(t, "testdata/link/multi_file/b.proto"),
				mustReadFixture(t, "testdata/link/multi_file/a.proto"),
			},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Len(t, issues, 2)

				// D-03 ordering holds across files: by path, then line, then col.
				assert.Equal(t, "a.proto", issues[0].Path)
				assert.Equal(t, CompileRuleName, issues[0].RuleName)
				assert.Equal(t, "b.proto", issues[1].Path)
				assert.Equal(t, CompileRuleName, issues[1].RuleName)
				assert.True(t, issues[0].Path < issues[1].Path, "issues must be sorted by path: %+v", issues)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, root)
			}
			engine := &Core{logger: logger.NewNop(), importRoots: []string{root}}
			if tt.newCore != nil {
				engine = tt.newCore(t, root)
			}
			walker := fs.NewFSWalker(root, ".")

			linked, issues := engine.linkSources(t.Context(), walker, tt.paths, tt.sources)

			tt.assert(t, engine, walker, linked, issues)
		})
	}
}

// mustReadFixture loads a testdata file relative to the package directory,
// following the position_parity_test.go fixture-loading pattern.
func mustReadFixture(t *testing.T, rel string) []byte {
	t.Helper()

	src, err := os.ReadFile(rel)
	require.NoError(t, err)
	return src
}

// TestLinkSourcesWarnDrop drives synthetic protoc.Diagnostic values through
// diagCollector — the exact handler linkSources registers with
// protoc.WithErrorHandler, the same function value protoc.Link invokes at
// runtime. The collector slice is diagnostics' only channel into the D-03
// sort and COMPILE issue construction, so a warn dropped at this handler
// structurally can never surface as a COMPILE issue (gap G-02-3, research
// A3). The test constructs diagnostic values; it never parses or matches
// diagnostic text (D-01).
func TestLinkSourcesWarnDrop(t *testing.T) {
	t.Parallel()

	t.Run("warn_dropped", func(t *testing.T) {
		t.Parallel()

		var mu sync.Mutex
		var diags []protoc.Diagnostic
		handler := diagCollector(&mu, &diags)

		handler(protoc.Diagnostic{Path: "test.proto", Line: 4, Col: 3, Message: "synthetic warning", IsWarn: true})

		require.Empty(t, diags)
	})

	t.Run("non_warn_collected", func(t *testing.T) {
		t.Parallel()

		var mu sync.Mutex
		var diags []protoc.Diagnostic
		handler := diagCollector(&mu, &diags)

		// The drop is IsWarn-selective, not a swallow-all: the same handler
		// construction collects a non-warn diagnostic.
		handler(protoc.Diagnostic{Path: "test.proto", Line: 4, Col: 3, Message: "synthetic error"})

		require.Equal(t, []protoc.Diagnostic{{Path: "test.proto", Line: 4, Col: 3, Message: "synthetic error"}}, diags)
	})

	t.Run("warn_mix_order", func(t *testing.T) {
		t.Parallel()

		var mu sync.Mutex
		var diags []protoc.Diagnostic
		handler := diagCollector(&mu, &diags)

		// A warn/error mix keeps exactly the error diagnostics, in arrival
		// order: the COMPILE input stream is neither shrunk nor corrupted.
		handler(protoc.Diagnostic{Path: "test.proto", Line: 1, Col: 1, Message: "synthetic warning first", IsWarn: true})
		handler(protoc.Diagnostic{Path: "test.proto", Line: 9, Col: 1, Message: "synthetic error at line 9"})
		handler(protoc.Diagnostic{Path: "test.proto", Line: 2, Col: 1, Message: "synthetic warning second", IsWarn: true})
		handler(protoc.Diagnostic{Path: "test.proto", Line: 12, Col: 1, Message: "synthetic error at line 12"})

		require.Len(t, diags, 2)
		require.Equal(t, []protoc.Diagnostic{
			{Path: "test.proto", Line: 9, Col: 1, Message: "synthetic error at line 9"},
			{Path: "test.proto", Line: 12, Col: 1, Message: "synthetic error at line 12"},
		}, diags)
	})

	t.Run("handler_race_mixed", func(t *testing.T) {
		t.Parallel()

		// Contention pin on the extracted seam (research Pitfall 7): the
		// mutex discipline must hold under concurrent mixed warn/error posts
		// — no lost updates, no warn leakage.
		const goroutines = 4
		const pairsPerGoroutine = 25

		var mu sync.Mutex
		var diags []protoc.Diagnostic
		handler := diagCollector(&mu, &diags)

		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < pairsPerGoroutine; i++ {
					handler(protoc.Diagnostic{Path: "race.proto", Line: g + 1, Col: 1, Message: "synthetic warning", IsWarn: true})
					handler(protoc.Diagnostic{Path: "race.proto", Line: g + 1, Col: 1, Message: "synthetic error"})
				}
			}(g)
		}
		wg.Wait()

		// Exactly the posted error count arrives: goroutines*pairsPerGoroutine
		// error posts, interleaving legal but none lost.
		require.Len(t, diags, goroutines*pairsPerGoroutine)

		// The Line value identifies the posting goroutine, so per-identity
		// counts must be exact — order is not asserted (interleaving is legal).
		counts := make(map[int]int, goroutines)
		for _, d := range diags {
			require.False(t, d.IsWarn)
			counts[d.Line]++
		}
		for g := 1; g <= goroutines; g++ {
			assert.Equal(t, pairsPerGoroutine, counts[g], "goroutine identity %d", g)
		}
	})
}
