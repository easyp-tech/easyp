package core

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
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
			name:  "compiling_with_wkt",
			paths: []string{"event.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"google/protobuf/timestamp.proto\";\n\nmessage Event {\n  google.protobuf.Timestamp at = 1;\n}\n")},
			assert: func(t *testing.T, engine *Core, walker *fs.FSWalker, linked *protoc.Linked, issues []IssueInfo) {
				require.NotNil(t, linked)
				assert.Empty(t, issues)
			},
		},
		{
			name:  "missing_import",
			paths: []string{"orphan.proto"},
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
				// The direct importer is b.proto: readFilesFromImport reads one
				// level, so processing b.proto is where missing.proto surfaces.
				_, err := engine.lintProtoInfoRead(t.Context(), walker, "b.proto")
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
				_, err := engine.lintProtoInfoRead(t.Context(), walker, "a.proto")
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
