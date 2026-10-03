package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePolicyReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want PolicyReference
		err  string
	}{
		{name: "empty", raw: "", want: PolicyReference{}},
		{name: "whitespace", raw: "  \t\n", want: PolicyReference{}},
		{name: "local file", raw: "./base.yaml", want: PolicyReference{Raw: "./base.yaml", Local: true, Path: "base.yaml"}},
		{name: "local hidden directory", raw: "./.policies/base.yaml", want: PolicyReference{Raw: "./.policies/base.yaml", Local: true, Path: ".policies/base.yaml"}},
		{name: "local directory", raw: "../policies/", want: PolicyReference{Raw: "../policies/", Local: true, Path: "../policies"}},
		{name: "local parent", raw: "..", want: PolicyReference{Raw: "..", Local: true, Path: ".."}},
		{name: "exact module", raw: "github.com/acme/proto-policy", want: PolicyReference{Raw: "github.com/acme/proto-policy", Module: "github.com/acme/proto-policy"}},
		{name: "major module", raw: "github.com/acme/proto-policy/v2", want: PolicyReference{Raw: "github.com/acme/proto-policy/v2", Module: "github.com/acme/proto-policy/v2"}},
		{name: "fragment file", raw: "github.com/acme/proto-policy#base.yaml", want: PolicyReference{Raw: "github.com/acme/proto-policy#base.yaml", Module: "github.com/acme/proto-policy", Relative: "base.yaml", Fragment: true}},
		{name: "fragment directory", raw: "github.com/acme/proto-policy#lint/base", want: PolicyReference{Raw: "github.com/acme/proto-policy#lint/base", Module: "github.com/acme/proto-policy", Relative: "lint/base", Fragment: true}},
		{name: "major fragment", raw: "github.com/acme/proto-policy/v2#base.yaml", want: PolicyReference{Raw: "github.com/acme/proto-policy/v2#base.yaml", Module: "github.com/acme/proto-policy/v2", Relative: "base.yaml", Fragment: true}},
		{name: "rfc unfragmented", raw: "github.com/acme/proto-policy/lint-base", want: PolicyReference{Raw: "github.com/acme/proto-policy/lint-base", Module: "github.com/acme/proto-policy/lint-base"}},
		{name: "url module identity", raw: "https://example.test/acme/policies#base.yaml", want: PolicyReference{Raw: "https://example.test/acme/policies#base.yaml", Module: "https://example.test/acme/policies", Relative: "base.yaml", Fragment: true}},
		{name: "trailing module slash", raw: "github.com/acme/proto-policy/", want: PolicyReference{Raw: "github.com/acme/proto-policy/", Module: "github.com/acme/proto-policy"}},
		{name: "version is rejected", raw: "github.com/acme/proto-policy@v1.2.3", err: "must not repeat a version"},
		{name: "commit is rejected", raw: "github.com/acme/proto-policy#base.yaml@v1", err: "must not repeat a version"},
		{name: "absolute Git identity", raw: "/opt/policies/base.yaml", want: PolicyReference{Raw: "/opt/policies/base.yaml", Module: "/opt/policies/base.yaml"}},
		{name: "parent keyword", raw: "parent", err: "must name a module identity"},
		{name: "single element", raw: "policies", err: "must name a module identity"},
		{name: "parent segments", raw: "github.com/acme/../policies", err: "invalid module identity"},
		{name: "empty module", raw: "#base.yaml", err: "has no module identity"},
		{name: "repeated fragment", raw: "github.com/acme/policies#a#b", err: "more than one # fragment"},
		{name: "escaping fragment", raw: "github.com/acme/policies#../base.yaml", err: "must not leave its module"},
		{name: "absolute fragment", raw: "github.com/acme/policies#/base.yaml", err: "module-relative policy path"},
		{name: "fragment empty segment", raw: "github.com/acme/policies#lint//base.yaml", err: "portable module-relative policy path"},
		{name: "backslash", raw: `.policies\base.yaml`, err: "unsupported characters"},
		{name: "newline", raw: "github.com/acme/poli\ncies", err: "unsupported characters"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParsePolicyReference(tt.raw)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicyReferenceEmptyAndString(t *testing.T) {
	t.Parallel()

	empty, err := ParsePolicyReference("")
	require.NoError(t, err)
	assert.True(t, empty.Empty())
	assert.Empty(t, empty.String())

	reference, err := ParsePolicyReference(" ./base.yaml ")
	require.NoError(t, err)
	assert.False(t, reference.Empty())
	assert.Equal(t, "./base.yaml", reference.String())
}
