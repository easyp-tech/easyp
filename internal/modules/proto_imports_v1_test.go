package modules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1ImportsFollowProtobufSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		imports []string
	}{
		{
			name:   "ignore block comment",
			source: "syntax = \"proto3\";\n/*\nimport \"removed.proto\";\n*/\nmessage Item {}\n",
		},
		{
			name:    "same line imports",
			source:  `syntax = "proto3"; import "first.proto"; import "second.proto";`,
			imports: []string{"first.proto", "second.proto"},
		},
		{
			name:    "single quotes",
			source:  "syntax = \"proto3\";\nimport 'missing.proto';\n",
			imports: []string{"missing.proto"},
		},
		{
			name:    "escaped path",
			source:  "syntax = \"proto3\";\nimport \"external\\x2ftypes.proto\";\n",
			imports: []string{"external/types.proto"},
		},
		{
			name:    "comments inside import",
			source:  "syntax = \"proto3\";\nimport /* reason */ \"missing.proto\";\n",
			imports: []string{"missing.proto"},
		},
		{
			name:    "public and weak imports",
			source:  "syntax = \"proto3\";\nimport public \"public.proto\";\nimport weak \"weak.proto\";\n",
			imports: []string{"public.proto", "weak.proto"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			err := os.WriteFile(filepath.Join(root, "main.proto"), []byte(tt.source), 0o644)
			require.NoError(t, err)

			missing, err := unresolvedV1Imports(root, []string{"."})
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.imports, missing)

			imports, err := v1RootImports(root, []string{"."})
			require.NoError(t, err)
			want := make(map[string]bool, len(tt.imports))
			for _, path := range tt.imports {
				want[path] = true
			}
			assert.Equal(t, want, imports)
		})
	}
}

func TestV1ImportsRejectInvalidSyntax(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
	}{
		{name: "missing semicolon", source: `syntax = "proto3"; import "missing.proto"`},
		{name: "unterminated string", source: `syntax = "proto3"; import "missing.proto;`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeV1GenerateFixture(t, root, "main.proto", tt.source)

			_, missingErr := unresolvedV1Imports(root, []string{"."})
			_, importsErr := v1RootImports(root, []string{"."})

			assert.ErrorContains(t, missingErr, "main.proto")
			assert.ErrorContains(t, importsErr, "main.proto")
		})
	}
}
