package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseModuleMissingIdentityHasLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: "protobuf.mod:1: missing module directive"},
		{name: "roots block", raw: "roots (\n  proto\n)\n", want: "protobuf.mod:3: missing module directive"},
		{name: "comments", raw: "# first\n// second\n", want: "protobuf.mod:2: missing module directive"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseModule(strings.NewReader(tt.raw))

			require.EqualError(t, err, tt.want)
		})
	}
}
