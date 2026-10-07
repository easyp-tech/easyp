package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteProtoImportLiteralTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "regular", source: `syntax = "proto3"; import "svc.proto";`, want: `syntax = "proto3"; import "v1/svc.proto";`},
		{name: "public_comments", source: `syntax = "proto3"; import /* before */ public 'svc.proto' /* after */;`, want: `syntax = "proto3"; import /* before */ public "v1/svc.proto" /* after */;`},
		{name: "weak_escaped", source: `syntax = "proto3"; import weak "s\x76c.proto";`, want: `syntax = "proto3"; import weak "v1/svc.proto";`},
		{name: "concatenation", source: `syntax = "proto3"; import "s" /* keep */ 'vc.' "proto";`, want: `syntax = "proto3"; import "v1/svc.proto" /* keep */ "" "";`},
		{name: "crlf_non_import_strings", source: "syntax = \"proto3\";\r\n  import \"svc.proto\"; // svc.proto\r\noption java_package = \"svc.proto\";\r\nmessage M {}\r\n", want: "syntax = \"proto3\";\r\n  import \"v1/svc.proto\"; // svc.proto\r\noption java_package = \"svc.proto\";\r\nmessage M {}\r\n"},
		{name: "unicode_comment", source: "syntax = \"proto3\"; // комментарий\nimport \"svc.proto\";", want: "syntax = \"proto3\"; // комментарий\nimport \"v1/svc.proto\";"},
		{name: "duplicate_declarations", source: `syntax = "proto3"; import "svc.proto"; import public "svc.proto";`, want: `syntax = "proto3"; import "v1/svc.proto"; import public "v1/svc.proto";`},
		{name: "unrelated_import", source: `syntax = "proto3"; import "svc.proto"; import "other.proto";`, want: `syntax = "proto3"; import "v1/svc.proto"; import "other.proto";`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before := []byte(tt.source)
			updated, mappings, err := rewriteProtoImportLiterals("consumer.proto", before, func(name string) (string, error) {
				if name == "svc.proto" {
					return "v1/svc.proto", nil
				}
				return name, nil
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(updated))
			assert.Equal(t, [][2]string{{"svc.proto", "v1/svc.proto"}}, mappings)
			assert.Equal(t, tt.source, string(before))
		})
	}
}

func TestRewriteProtoImportsPreservesUnchangedSpelling(t *testing.T) {
	t.Parallel()
	before := []byte(`syntax = "proto3"; import "s\x76c." /* keep */ "proto";`)
	updated, mappings, err := rewriteProtoImportLiterals("consumer.proto", before, func(name string) (string, error) { return name, nil })

	require.NoError(t, err)
	assert.Equal(t, before, updated)
	assert.Empty(t, mappings)
}
