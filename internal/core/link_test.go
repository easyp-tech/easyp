package core

import (
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
		assert  func(t *testing.T, linked *protoc.Linked, issues []IssueInfo)
	}{
		{
			name:    "compile_error",
			paths:   []string{"test.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\npackage example;\nmessage bad_name {\n  required string s = 1;\n}\n")},
			assert: func(t *testing.T, linked *protoc.Linked, issues []IssueInfo) {
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
			assert: func(t *testing.T, linked *protoc.Linked, issues []IssueInfo) {
				require.NotNil(t, linked)
				assert.Empty(t, issues)
			},
		},
		{
			name:  "missing_import",
			paths: []string{"orphan.proto"},
			sources: [][]byte{[]byte("syntax = \"proto3\";\n\npackage example;\n\nimport \"missing.proto\";\n")},
			assert: func(t *testing.T, linked *protoc.Linked, issues []IssueInfo) {
				require.Nil(t, linked)
				require.Nil(t, issues)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			engine := &Core{logger: logger.NewNop(), importRoots: []string{root}}
			walker := fs.NewFSWalker(root, ".")

			linked, issues := engine.linkSources(t.Context(), walker, tt.paths, tt.sources)

			tt.assert(t, linked, issues)
		})
	}
}
