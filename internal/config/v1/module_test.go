package v1

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsModuleManifestAcceptsParsedSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "spaces", raw: "module github.com/acme/user\n"},
		{name: "tab", raw: "module\tgithub.com/acme/user\n"},
		{name: "unicode whitespace", raw: "module\u00a0github.com/acme/user\n"},
		{name: "comments", raw: "// comment\n\n\tmodule\tgithub.com/acme/user // comment\n"},
		{name: "roots before module", raw: "roots api\nmodule github.com/acme/user\n"},
		{name: "require block before module", raw: "require (\n github.com/acme/common v1.0.0\n)\nmodule github.com/acme/user\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseModule(strings.NewReader(tt.raw))
			require.NoError(t, err)
			assert.True(t, IsModuleManifest([]byte(tt.raw)))
		})
	}
}

func TestIsModuleManifestKeepsLegacySyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "empty", raw: "\n// comment\n"},
		{name: "direct", raw: "direct (\n github.com/acme/common@v1.0.0\n)\n"},
		{name: "module dependency", raw: "direct (\n module\n)\n"},
		{name: "malformed direct", raw: "direct (\n module github.com/acme/user\n)\n"},
		{name: "replace", raw: "replace module@v1.0.0 => ./local\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.False(t, IsModuleManifest([]byte(tt.raw)))
		})
	}
}

func TestIsModuleManifestPreservesV1ParseErrors(t *testing.T) {
	t.Parallel()

	raw := "module github.com/acme/user\nunknown directive\n"
	assert.True(t, IsModuleManifest([]byte(raw)))
	_, err := ParseModule(strings.NewReader(raw))
	require.ErrorContains(t, err, "unknown directive")
}

func TestParseModuleRoots(t *testing.T) {
	got, err := ParseModule(strings.NewReader(`// schema module
module github.com/acme/user
roots (
  api
  shared // comment
)
`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "github.com/acme/user" || !reflect.DeepEqual(got.Roots, []string{"api", "shared"}) {
		t.Fatalf("unexpected module: %#v", got)
	}
}

func TestParseModuleDefaultRoot(t *testing.T) {
	got, err := ParseModule(strings.NewReader("module github.com/acme/user\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Roots, []string{"."}) {
		t.Fatalf("roots = %v", got.Roots)
	}
}

func TestParseModuleRejectsUnknownDirective(t *testing.T) {
	if _, err := ParseModule(strings.NewReader("module github.com/acme/user\nplugin go\n")); err == nil {
		t.Fatal("expected unknown directive error")
	}
}

func TestParseModuleHTTPSIdentity(t *testing.T) {
	got, err := ParseModule(strings.NewReader("module https://example.com/acme/user.git\nrequire https://example.com/acme/common.git v1.2.3 // indirect\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "https://example.com/acme/user.git" || len(got.Requires) != 1 || got.Requires[0].Module != "https://example.com/acme/common.git" {
		t.Fatalf("unexpected module: %#v", got)
	}
}
