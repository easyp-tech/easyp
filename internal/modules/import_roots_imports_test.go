package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntrinsicProtoImportsUsesProtobufDeclarations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		want   []string
	}{
		{name: "before_unterminated_body", source: `syntax = "proto3"; import "svc.proto"; message Broken {`, want: []string{"svc.proto"}},
		{name: "public_single_quote_escape", source: `import public 's\x76c.proto'; message Broken {`, want: []string{"svc.proto"}},
		{name: "weak_concatenated_literals", source: `import weak "svc/" /* comment */ "v1/svc.proto";`, want: []string{"svc/v1/svc.proto"}},
		{name: "comments_options_and_nested_body", source: `/* import "fake.proto"; */ option text = 'import "fake.proto";'; message Body { import "fake.proto"; } import "real.proto";`, want: []string{"real.proto"}},
		{name: "invalid_declaration_ignored", source: `import "missing-semicolon.proto" message Broken {`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, intrinsicProtoImports("fixture.proto", []byte(tt.source)))
		})
	}
}
